package tunnel_test

import (
	"context"
	"net"
	"net/http"
	"net/url"
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
		t.Errorf("连接通道时发生不可重试错误：%v", err)
		return
	}
	t.Log("通道连接成功了")

	// ------------------------[ 业务自定义层 ]---------------------------

	// ⚠️ 这个 httpclient DialContext 会判断 host 如果是 broker.ssoc.internal 就走内部 tunnel，
	// 否则就走公网网络，如果有隔离需求请自行隔离。
	const internalHost = "broker.ssoc.internal"
	systemDialer := new(net.Dialer)
	multiHTTPClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				if host, _, _ := net.SplitHostPort(addr); host == internalHost {
					return mux.OpenConn(ctx)
				}

				return systemDialer.DialContext(ctx, network, addr)
			},
		},
	}

	{
		// 可以将该方法拿出来公用。
		buildURL := func(path string) *url.URL {
			return &url.URL{
				Scheme: "http",
				Host:   internalHost,
				Path:   path,
			}
		}
		// 走 tunnel 通道调用 broker 的内部接口
		// ⚠️ 协议一定要是 http 或 ws，
		// ⚠️ host 一定要和 httpclient 的 DialContext 判断一致。
		reqURL := buildURL("/foo/bar")
		resp, _ := multiHTTPClient.Get(reqURL.String())
		if resp != nil {
			_ = resp.Body.Close()
		}
	}
	{
		// 向外部发送 http 请求。
		resp, _ := multiHTTPClient.Get("https://example.com/foo/bar")
		if resp != nil {
			_ = resp.Body.Close()
		}
	}

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
