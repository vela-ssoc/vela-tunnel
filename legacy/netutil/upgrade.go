package netutil

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func Upgrade(errFn func(http.ResponseWriter, *http.Request, int, error)) websocket.Upgrader {
	return websocket.Upgrader{
		HandshakeTimeout:  5 * time.Second,
		ReadBufferSize:    4 * 1024,
		WriteBufferSize:   4 * 1024,
		Error:             errFn,
		CheckOrigin:       func(*http.Request) bool { return true },
		EnableCompression: true,
	}
}
