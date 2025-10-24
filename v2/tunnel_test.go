package tunnel_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/vela-ssoc/vela-tunnel/v2"
)

func TestTunnel(t *testing.T) {
	ctx := context.Background()
	opt := tunnel.NewOption().
		Server(yourHTTPServer()).        // 业务服务器（HTTP 服务）
		Identifier(tunnel.NewIdent("")). // 机器码生成器
		Notifier(&connectNotifier{t: t}) // 通道状态变化通知

	cfg := tunnel.Config{
		Addresses: []string{"broker.example.com:8082", "127.0.0.1:8082"},
		Semver:    "1.2.3-alpha",
	}
	mux, err := tunnel.Open(ctx, cfg, opt)
	if err != nil {
		t.Log(err)
		return
	}

	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		var last bool
		for range ticker.C {
			closed := mux.IsClosed()
			if closed != last {
				last = closed
				if closed {
					t.Log("糟糕，通道掉线了。")
				} else {
					t.Log("太好了，通道重连成功了。")
				}
			}
		}
	}()

	_ = mux
	time.Sleep(time.Hour)
}

func yourHTTPServer() tunnel.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/path/to/router", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("处理成功"))
	})

	return &http.Server{
		Handler: mux,
	}
}

type connectNotifier struct {
	t *testing.T
}

func (c *connectNotifier) Connected() {
	c.t.Log("agent 首次连接上线成功")
}

func (c *connectNotifier) Disconnected(err error) {
	c.t.Logf("agent 掉线了: %v", err)
}

func (c *connectNotifier) Reconnected() {
	c.t.Log("agent 重连成功了")
}

func (c *connectNotifier) Exited(err error) {
	c.t.Errorf("【严重】agent 退出不再尝试重连: %v", err)
}
