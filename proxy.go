package tunnel

import (
	"net/http"
	"strings"
)

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type tunnelDo struct {
	tun    *borerTunnel
	prefix string
}

func (td *tunnelDo) Do(req *http.Request) (*http.Response, error) {
	path := req.URL.Path
	if strings.HasPrefix(path, "/") {
		req.URL.Path = td.prefix + path
	} else {
		req.URL.Path = td.prefix + "/" + path
	}

	return td.tun.client.Do(req)
}
