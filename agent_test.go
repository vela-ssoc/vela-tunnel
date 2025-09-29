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

	tun, err := tunnel.Dial(parent, hide, srv, tunnel.WithNotifier(&notifier{t: t}))
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(tun.Inet())

	<-parent.Done()
}

type notifier struct {
	t *testing.T
}

func (n notifier) Connected(addr *tunnel.Address) {
	n.t.Log("连接成功了", addr.String())
}

func (n notifier) Disconnect(err error) {

}

func (n notifier) Reconnected(addr *tunnel.Address) {
}

func (n notifier) Shutdown(err error) {
}
