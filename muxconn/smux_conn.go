package muxconn

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/xtaci/smux"
)

type smuxConn struct {
	parent *smuxSession
	stream *smux.Stream
	stats  *streamStats
	limit  io.ReadWriter
	cancel context.CancelCauseFunc
}

func (c *smuxConn) Read(b []byte) (int, error) {
	n, err := c.limit.Read(b)
	c.stats.incrRX(n)

	return n, err
}

func (c *smuxConn) Write(b []byte) (int, error) {
	n, err := c.limit.Write(b)
	c.stats.incrTX(n)

	return n, err
}

func (c *smuxConn) Close() error {
	c.parent.stats.delConn(c)
	c.cancel(net.ErrClosed)

	return c.stream.Close()
}

func (c *smuxConn) LocalAddr() net.Addr                { return c.stream.LocalAddr() }
func (c *smuxConn) RemoteAddr() net.Addr               { return c.stream.RemoteAddr() }
func (c *smuxConn) SetDeadline(t time.Time) error      { return c.stream.SetDeadline(t) }
func (c *smuxConn) SetReadDeadline(t time.Time) error  { return c.stream.SetReadDeadline(t) }
func (c *smuxConn) SetWriteDeadline(t time.Time) error { return c.stream.SetWriteDeadline(t) }
func (c *smuxConn) Stats() *StreamStats                { return c.stats.stats() }
func (c *smuxConn) setTrunkStats(s *streamStats)       { c.stats = s }
