package tunnel_test

import (
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"

	tunnel "github.com/vela-ssoc/vela-tunnel/v2"
)

func TestTunnel(t *testing.T) {
	srv := &http.Server{}

	ctx := context.Background()
	opt := tunnel.NewOption().
		Logger(slog.Default()).
		Server(srv).
		Identifier(tunnel.NewIdent(""))

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
