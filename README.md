# 安全平台 agent 节点通信连接

## 代码示例

```go
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/vela-ssoc/backend-common/encipher"
	"github.com/vela-ssoc/backend-common/logback"
	"github.com/vela-ssoc/vela-tunnel"
)

func main() {
	var hide tunnel.Hide

	// -----[ 实际生产环境要从自身执行文件中读取 hide 信息 ]-----
	{
		_ = encipher.ReadFile(os.Args[0], &hide)
	}

	// -----[ 测试开发调试手动设置 hide 信息 ]-----
	{
		addr := tunnel.Addresses{
			{TLS: true, Addr: "172.31.61.168", Name: "local.eastmoney.com"},
			{Addr: "172.31.61.168:8180"},
		}
		hide.Semver = "0.0.1-delve"
		hide.Ethernet = addr
	}

	// 监听停止信号
	cares := []os.Signal{syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL, syscall.SIGINT}
	ctx, cancel := signal.NotifyContext(context.Background(), cares...)
	defer cancel()
	slog := logback.Stdout()
	slog.Info("按 Ctrl+C 结束运行")

	tun, err := tunnel.Dial(ctx, hide, tunnel.WithLogger(slog), tunnel.WithInterval(time.Minute))
	if err != nil {
		slog.Errorf("连接 broker 发生错误：%s", err)
		return
	}

	ident := tun.Ident()
	issue := tun.Issue()
	slog.Infof("上报的 ident 如下：\n%s\n下发的 issue 内容如下：\n%s\n", ident, issue)

	// 建立双向流
	// tun.Stream(opurl.Kafka, nil)
	// 请求响应
	// tun.JSON(nil, opurl.OpPing, req, res)

	// 起个名字用于后期便于排查错误
	// Example: tunnel-172.36.18.18-184309536616640003
	node := "tunnel-" + ident.Inet.String() + "-" + strconv.FormatInt(issue.ID, 10)

	// 根据实际业务注册 http router
	mux := http.NewServeMux()
	mux.HandleFunc("/api/echo", func(w http.ResponseWriter, r *http.Request) {
		slog.Infof("%s 收到了中心端下发的消息", node)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("我是消息已经收到了！！！"))
	})

	errch := make(chan error, 1)
	dw := &daemonWatch{
		slog:   slog,
		handle: mux,
		tun:    tun,
		parent: ctx,
		errch:  errch,
	}
	go dw.Run()

	select {
	case err = <-errch:
	case <-ctx.Done():
	}

	slog.Warnf("程序执行结束：%v", err)
}

type daemonWatch struct {
	slog   logback.Logger
	handle http.Handler
	tun    tunnel.Tunneler
	parent context.Context
	errch  chan<- error
}

func (dw *daemonWatch) Run() {
	var err error
	srv := &http.Server{Handler: dw.handle}

over:
	for {
		lis := dw.tun.Listen()
		err = srv.Serve(lis)
		if pe := dw.parent.Err(); pe != nil {
			err = pe
			break over
		}
		dw.slog.Warnf("连接已经断开，即将重连：%s", err)
		if err = dw.tun.Reconnect(dw.parent); err != nil {
			break over
		}
	}

	dw.slog.Warnf("断开连接：%s", err)
	dw.errch <- err
}

```