package netutil

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func Pipe(a net.Conn, b *websocket.Conn) PipeStat {
	bw := warpRWC(b)

	var ret PipeStat
	wg := new(sync.WaitGroup)
	wg.Add(1)

	go func() {
		defer wg.Done()
		wn, err := io.CopyBuffer(bw, a, make([]byte, 1024))
		_ = bw.Close()
		ret.Atob = wn
		ret.AErr = err
	}()

	wn, err := io.CopyBuffer(a, bw, make([]byte, 1024))
	_ = a.Close()
	ret.Btoa = wn
	ret.BErr = err
	wg.Wait()

	return ret
}

func PipeWebsocket(a, b *websocket.Conn) PipeStat {
	aw := warpRWC(a)
	bw := warpRWC(b)

	ret := PipeStat{Start: time.Now()}
	wg := new(sync.WaitGroup)
	wg.Add(1)

	go func() {
		defer wg.Done()
		wn, err := io.CopyBuffer(bw, aw, make([]byte, 1024))
		_ = bw.Close()
		ret.Atob = wn
		ret.AErr = err
	}()

	wn, err := io.CopyBuffer(aw, bw, make([]byte, 1024))
	_ = aw.Close()
	ret.Btoa = wn
	ret.BErr = err
	wg.Wait()
	ret.End = time.Now()

	return ret
}

type PipeStat struct {
	Atob  int64 // a 发送给 b 的字节数
	Btoa  int64 // b 发送给 a 的字节数
	AErr  error
	BErr  error
	Start time.Time // 数据交换开始时间
	End   time.Time // 执行结束时间
}

func (s PipeStat) String() string {
	tpl := "通道数据交换 a->b %d bytes, b->a %d bytes, 时长: %s"
	return fmt.Sprintf(tpl, s.Atob, s.Btoa, s.End.Sub(s.Start))
}

func warpRWC(c *websocket.Conn) io.ReadWriteCloser {
	return &websocketRWC{
		c: c,
		r: websocket.JoinMessages(c, ""),
	}
}

type websocketRWC struct {
	c *websocket.Conn
	r io.Reader
}

func (w *websocketRWC) Read(p []byte) (int, error) {
	return w.r.Read(p)
}

func (w *websocketRWC) Write(p []byte) (int, error) {
	n := len(p)
	var err error
	if n != 0 {
		err = w.c.WriteMessage(websocket.BinaryMessage, p)
	}

	return n, err
}

func (w *websocketRWC) Close() error { return w.c.Close() }
