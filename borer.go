package tunnel

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/user"
	"runtime"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vela-ssoc/backend-common/logback"
	"github.com/vela-ssoc/backend-common/netutil"
	"github.com/vela-ssoc/backend-common/opurl"
	"github.com/vela-ssoc/backend-common/spdy"
)

// borerTunnel 通道连接器
type borerTunnel struct {
	hide   Hide               // hide
	ident  Ident              // ident
	issue  Issue              // issue
	dialer dialer             // TCP 连接器
	muxer  spdy.Muxer         // 底层流复用
	client opurl.Client       // http 客户端
	stream netutil.Streamer   // 建立流式通道用
	slog   logback.Logger     // 日志输出组件
	parent context.Context    // parent context.Context
	ctx    context.Context    // context.Context
	cancel context.CancelFunc // context.CancelFunc
}

func (bt *borerTunnel) ID() int64 {
	return bt.issue.ID
}

func (bt *borerTunnel) Inet() net.IP {
	return bt.ident.Inet
}

func (bt *borerTunnel) Hide() Hide {
	return bt.hide
}

func (bt *borerTunnel) Ident() Ident {
	return bt.ident
}

func (bt *borerTunnel) Issue() Issue {
	return bt.issue
}

func (bt *borerTunnel) Listen() net.Listener {
	return bt.muxer
}

func (bt *borerTunnel) NodeName() string {
	return fmt.Sprintf("minion-%s-%d", bt.Inet(), bt.ID())
}

func (bt *borerTunnel) Reconnect(ctx context.Context) error {
	_ = bt.muxer.Close()
	bt.cancel()
	if ctx == nil {
		ctx = bt.parent
	}
	return bt.dial(ctx)
}

func (bt *borerTunnel) Fetch(ctx context.Context, op opurl.URLer, rd io.Reader) (*http.Response, error) {
	return bt.client.Fetch(ctx, op, nil, rd)
}

func (bt *borerTunnel) Oneway(ctx context.Context, op opurl.URLer, rd io.Reader) error {
	res, err := bt.client.Fetch(ctx, op, nil, rd)
	if err == nil {
		_ = res.Body.Close()
	}
	return err
}

func (bt *borerTunnel) JSON(ctx context.Context, op opurl.URLer, req any, reply any) error {
	return bt.client.JSON(ctx, op, nil, req, reply)
}

func (bt *borerTunnel) OnewayJSON(ctx context.Context, op opurl.URLer, req any) error {
	return bt.client.OnewayJSON(ctx, op, nil, req)
}

func (bt *borerTunnel) Attachment(ctx context.Context, op opurl.URLer) (opurl.Attachment, error) {
	return bt.client.Attachment(ctx, op)
}

func (bt *borerTunnel) Stream(op opurl.URLer, header http.Header) (*websocket.Conn, error) {
	addr := op.String()
	conn, _, err := bt.stream.Stream(addr, header)
	if err == nil {
		bt.slog.Infof("建立 stream (%s) 通道成功", addr)
	} else {
		bt.slog.Warnf("建立 stream (%s) 通道失败：%s", addr, err)
	}

	return conn, err
}

func (bt *borerTunnel) dialContext(context.Context, string, string) (net.Conn, error) {
	return bt.muxer.Dial()
}

func (bt *borerTunnel) heartbeat(inter time.Duration) {
	ticker := time.NewTicker(inter)
	defer ticker.Stop()

over:
	for {
		select {
		case <-bt.parent.Done():
			break over
		case <-ticker.C:
			if err := bt.Oneway(nil, opurl.OpPing, nil); err != nil {
				bt.slog.Warnf("心跳包发送出错：%v", err)
			}
		}
	}
}

func (bt *borerTunnel) dial(parent context.Context) error {
	bt.parent = parent
	bt.ctx, bt.cancel = context.WithCancel(parent)
	start := time.Now()

	bt.slog.Info("开始连接 broker ...")
	for {
		conn, addr, err := bt.dialer.iterDial(bt.ctx, 3*time.Second)
		if err != nil {
			du := bt.waitN(start)
			bt.slog.Warnf("连接 broker(%s) 发生错误: %s, %s 后重试", addr, err, du)
			if err = bt.sleepN(du); err != nil {
				return err
			}
			continue
		}
		ctx, cancel := context.WithTimeout(bt.ctx, 5*time.Second)
		ident, issue, err := bt.consult(ctx, conn, addr)
		cancel()
		if err == nil {
			bt.ident, bt.issue = ident, issue
			bt.muxer = spdy.Client(conn, spdy.WithEncrypt(issue.Passwd))
			bt.slog.Infof("连接 broker(%s) 成功", addr)
			return nil
		}

		du := bt.waitN(start)
		bt.slog.Warnf("与 broker(%s) 发生错误: %s, %s 后重试", addr, err, du)
		if err = bt.sleepN(du); err != nil {
			return err
		}
	}
}

// consult 当建立好 TCP 连接后进行应用层协商
func (bt *borerTunnel) consult(parent context.Context, conn net.Conn, addr *Address) (Ident, Issue, error) {
	ip := conn.LocalAddr().(*net.TCPAddr).IP
	mac := bt.dialer.lookupMAC(ip)

	ident := Ident{
		Semver: bt.hide.Semver,
		Inet:   ip,
		MAC:    mac.String(),
		Goos:   runtime.GOOS,
		Arch:   runtime.GOARCH,
		CPU:    runtime.NumCPU(),
		PID:    os.Getpid(),
		TimeAt: time.Now(),
	}
	ident.Hostname, _ = os.Hostname()
	ident.Workdir, _ = os.Getwd()
	ident.Executable, _ = os.Executable()
	if cu, _ := user.Current(); cu != nil {
		ident.Username = cu.Username
	}

	var issue Issue
	enc, err := ident.encrypt()
	if err != nil {
		return ident, issue, err
	}

	body := bytes.NewReader(enc)
	req := bt.client.NewRequest(parent, opurl.MonJoin, nil, body)
	host := addr.Name
	if host == "" {
		host, _, _ = net.SplitHostPort(addr.Addr)
	}
	req.Host = host
	if err = req.Write(conn); err != nil {
		return ident, issue, nil
	}

	res, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return ident, issue, nil
	}
	//goland:noinspection GoUnhandledErrorResult
	defer res.Body.Close()

	code := res.StatusCode
	if code != http.StatusAccepted {
		cause := make([]byte, 4096)
		n, _ := io.ReadFull(res.Body, cause)
		ret := struct {
			Message string `json:"message"`
		}{}
		exr := &opurl.Error{Code: code}
		if err = json.Unmarshal(cause[:n], &ret); err == nil {
			exr.Text = []byte(ret.Message)
		} else {
			exr.Text = cause[:n]
		}
		return ident, issue, exr
	}

	resp := make([]byte, 100*1024)
	n, _ := res.Body.Read(resp)
	err = issue.decrypt(resp[:n])

	return ident, issue, nil
}

func (bt *borerTunnel) waitN(start time.Time) time.Duration {
	since := time.Since(start)
	du := time.Second
	switch {
	case since > 12*time.Hour:
		du = 10 * time.Minute
	case since > time.Hour:
		du = time.Minute
	case since > 30*time.Minute:
		du = 30 * time.Second
	case since > 10*time.Minute:
		du = 10 * time.Second
	case since > 3*time.Minute:
		du = 3 * time.Second
	}
	return du
}

func (bt *borerTunnel) sleepN(du time.Duration) error {
	timer := time.NewTimer(du)
	defer timer.Stop()
	var err error
	ctx := bt.ctx
	select {
	case <-timer.C:
	case <-ctx.Done():
		err = ctx.Err()
	}
	return err
}
