package netutil

import (
	"net/http"
	"net/http/httputil"
)

type Forwarder interface {
	Forward(http.ResponseWriter, *http.Request)
}

func NewForward(trip http.RoundTripper, errFn func(http.ResponseWriter, *http.Request, error)) Forwarder {
	px := &httputil.ReverseProxy{
		Transport:    trip,
		ErrorHandler: errFn,
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetXForwarded()
		},
	}
	return &forward{
		px: px,
	}
}

type forward struct {
	px *httputil.ReverseProxy
}

func (f *forward) Forward(w http.ResponseWriter, r *http.Request) {
	f.px.ServeHTTP(w, r)
}
