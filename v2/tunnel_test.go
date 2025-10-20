package tunnel_test

import (
	"context"
	"testing"

	tunnel "github.com/vela-ssoc/vela-tunnel/v2"
)

func TestTunnel(t *testing.T) {
	cfg := tunnel.Config{
		Addresses: []string{"172.31.61.168:8082", "172.31.61.168:8083"},
		Semver:    "1.2.3-alpha",
	}
	ctx := context.Background()
	mux, err := tunnel.Open(ctx, cfg)
	if err != nil {
		t.Log(err)
		return
	}

	_ = mux
}
