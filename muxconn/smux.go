package muxconn

import (
	"context"
	"net"
	"time"

	"github.com/xtaci/smux"
	"golang.org/x/time/rate"
)

// NewSMUX 基于 xtaci/smux 实现的多路复用。
func NewSMUX(parent context.Context, conn net.Conn, cfg *smux.Config, serverSide bool) (Muxer, error) {
	var err error
	mux := &smuxSession{
		stats:   newMUXStreamStats(),
		limiter: newUnlimit(),
		connAt:  time.Now(),
		parent:  parent,
	}
	if serverSide {
		mux.session, err = smux.Server(conn, cfg)
	} else {
		mux.session, err = smux.Client(conn, cfg)
	}
	if err != nil {
		return nil, err
	}

	return mux, nil
}

type smuxSession struct {
	session *smux.Session
	stats   *muxStreamStats
	limiter *rateLimiter
	connAt  time.Time // 创建连接的时间点
	parent  context.Context
}

func (m *smuxSession) Open(context.Context) (net.Conn, error) {
	return m.newConn(m.session.OpenStream())
}

func (m *smuxSession) Accept() (net.Conn, error)    { return m.newConn(m.session.AcceptStream()) }
func (m *smuxSession) Close() error                 { return m.session.Close() }
func (m *smuxSession) Addr() net.Addr               { return m.session.LocalAddr() }
func (m *smuxSession) RemoteAddr() net.Addr         { return m.session.RemoteAddr() }
func (m *smuxSession) IsClosed() bool               { return m.session.IsClosed() }
func (m *smuxSession) Limit() rate.Limit            { return m.limiter.Limit() }
func (m *smuxSession) SetLimit(bps rate.Limit) bool { return m.limiter.SetLimit(bps) }
func (m *smuxSession) NumStreams() (int64, int64)   { return m.stats.numStreams() }
func (m *smuxSession) Traffic() (uint64, uint64)    { return m.stats.mux.load() }
func (m *smuxSession) Library() (string, string)    { return "smux", "github.com/xtaci/smux" }
func (m *smuxSession) Streams() []Streamer          { return m.stats.actives() }
func (m *smuxSession) ConnectedAt() time.Time       { return m.connAt }

func (m *smuxSession) newConn(stm *smux.Stream, err error) (*smuxConn, error) {
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancelCause(m.parent)
	limit := m.limiter.warpReadWriter(ctx, stm)

	conn := &smuxConn{
		parent: m,
		stream: stm,
		limit:  limit,
		cancel: cancel,
	}
	m.stats.putConn(conn)

	return conn, nil
}
