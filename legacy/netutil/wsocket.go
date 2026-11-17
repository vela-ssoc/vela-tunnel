package netutil

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Streamer interface {
	Stream(context.Context, string, http.Header) (*websocket.Conn, *http.Response, error)
}

func NewStream(dial func(context.Context, string, string) (net.Conn, error)) Streamer {
	sock := &websocket.Dialer{
		NetDialContext:    dial,
		HandshakeTimeout:  5 * time.Second,
		ReadBufferSize:    4 * 1024,
		WriteBufferSize:   4 * 1024,
		EnableCompression: true,
	}

	return &stream{
		sock: sock,
	}
}

type stream struct {
	sock *websocket.Dialer
}

func (stm *stream) Stream(ctx context.Context, addr string, header http.Header) (*websocket.Conn, *http.Response, error) {
	conn, res, err := stm.sock.DialContext(ctx, addr, header)
	if err == nil || res == nil {
		return conn, res, err
	}

	buf := make([]byte, 4096)
	code := res.StatusCode
	n, _ := res.Body.Read(buf)
	_ = res.Body.Close()

	exx := &HTTPError{Code: code, Body: buf[:n]}

	return nil, nil, exx
}
