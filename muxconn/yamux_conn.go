package muxconn

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/hashicorp/yamux"
)

type yamuxConn struct {
	parent *yamuxSession
	stream *yamux.Stream
	stats  *streamStats
	limit  io.ReadWriter
	cancel context.CancelCauseFunc
}

func (c *yamuxConn) Read(b []byte) (int, error) {
	n, err := c.limit.Read(b)
	c.stats.incrRX(n)

	return n, err
}

func (c *yamuxConn) Write(b []byte) (int, error) {
	n, err := c.limit.Write(b)
	c.stats.incrTX(n)

	return n, err
}

func (c *yamuxConn) Close() error {
	c.parent.stats.delConn(c)
	c.cancel(net.ErrClosed)

	return c.stream.Close()
}

func (c *yamuxConn) LocalAddr() net.Addr                { return c.stream.LocalAddr() }
func (c *yamuxConn) RemoteAddr() net.Addr               { return c.stream.RemoteAddr() }
func (c *yamuxConn) SetDeadline(t time.Time) error      { return c.stream.SetDeadline(t) }
func (c *yamuxConn) SetReadDeadline(t time.Time) error  { return c.stream.SetReadDeadline(t) }
func (c *yamuxConn) SetWriteDeadline(t time.Time) error { return c.stream.SetWriteDeadline(t) }
func (c *yamuxConn) Stats() *StreamStats                { return c.stats.stats() }
func (c *yamuxConn) setTrunkStats(s *streamStats)       { c.stats = s }
