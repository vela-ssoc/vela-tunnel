package tunnel_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/vela-ssoc/vela-common-mba/definition"
	tunnel "github.com/vela-ssoc/vela-tunnel"
)

func TestAgent(t *testing.T) {
	parent := context.Background()
	hide := definition.MHide{
		Addrs:  []string{"172.31.61.168:8082"},
		Semver: "0.0.1-local",
	}
	mux := http.NewServeMux()
	srv := &http.Server{Handler: mux}

	tun, err := tunnel.Dial(parent, hide, srv)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(tun.Inet())

	<-parent.Done()
}
