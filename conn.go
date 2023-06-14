package tunnel

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type websocketConn struct {
	ws  *websocket.Conn
	rd  io.Reader
	rwu sync.Mutex
	wmu sync.Mutex
}

func (wc *websocketConn) Read(b []byte) (int, error) {
	wc.rwu.Lock()
	n, err := wc.rd.Read(b)
	wc.rwu.Unlock()
	return n, err
}

func (wc *websocketConn) Write(b []byte) (int, error) {
	n := len(b)
	wc.wmu.Lock()
	err := wc.ws.WriteMessage(websocket.BinaryMessage, b)
	wc.wmu.Unlock()

	return n, err
}

func (wc *websocketConn) Close() error {
	return wc.ws.Close()
}

func (wc *websocketConn) LocalAddr() net.Addr {
	return wc.ws.LocalAddr()
}

func (wc *websocketConn) RemoteAddr() net.Addr {
	return wc.ws.RemoteAddr()
}

func (wc *websocketConn) SetDeadline(t time.Time) error {
	err := wc.ws.SetReadDeadline(t)
	if err == nil {
		err = wc.ws.SetWriteDeadline(t)
	}
	return err
}

func (wc *websocketConn) SetReadDeadline(t time.Time) error {
	return wc.ws.SetReadDeadline(t)
}

func (wc *websocketConn) SetWriteDeadline(t time.Time) error {
	return wc.ws.SetWriteDeadline(t)
}
