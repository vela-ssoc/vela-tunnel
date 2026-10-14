package muxconn

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/vela-ssoc/vela-common-mba/smux"
)

type velaConn struct {
	parent *velaSession
	stream *smux.Stream
	stats  *streamStats
	limit  io.ReadWriter
	cancel context.CancelCauseFunc
}

func (c *velaConn) Read(b []byte) (int, error) {
	n, err := c.limit.Read(b)
	c.stats.incrRX(n)

	return n, err
}

func (c *velaConn) Write(b []byte) (int, error) {
	n, err := c.limit.Write(b)
	c.stats.incrTX(n)

	return n, err
}

func (c *velaConn) Close() error {
	c.parent.stats.delConn(c)
	c.cancel(net.ErrClosed)

	return c.stream.Close()
}

func (c *velaConn) LocalAddr() net.Addr                { return c.stream.LocalAddr() }
func (c *velaConn) RemoteAddr() net.Addr               { return c.stream.RemoteAddr() }
func (c *velaConn) SetDeadline(t time.Time) error      { return c.stream.SetDeadline(t) }
func (c *velaConn) SetReadDeadline(t time.Time) error  { return c.stream.SetReadDeadline(t) }
func (c *velaConn) SetWriteDeadline(t time.Time) error { return c.stream.SetWriteDeadline(t) }
func (c *velaConn) Stats() *StreamStats                { return c.stats.stats() }
func (c *velaConn) setTrunkStats(s *streamStats)       { c.stats = s }
