package muxconn

import (
	"context"
	"net"
	"time"

	"github.com/vela-ssoc/vela-common-mba/smux"
	"golang.org/x/time/rate"
)

func NewVela(parent context.Context, conn net.Conn, cfg *smux.Config, serverSide bool) (Muxer, error) {
	mux := &velaSession{
		stats:   newMUXStreamStats(),
		limiter: newUnlimit(),
		connAt:  time.Now(),
		parent:  parent,
	}
	if serverSide {
		mux.session = smux.Server(conn, cfg)
	} else {
		mux.session = smux.Client(conn, cfg)
	}

	return mux, nil
}

type velaSession struct {
	session *smux.Session
	stats   *muxStreamStats
	limiter *rateLimiter
	connAt  time.Time // 创建连接的时间点
	parent  context.Context
}

func (m *velaSession) Open(context.Context) (net.Conn, error) {
	return m.newConn(m.session.OpenStream())
}

func (m *velaSession) Accept() (net.Conn, error)    { return m.newConn(m.session.AcceptStream()) }
func (m *velaSession) Close() error                 { return m.session.Close() }
func (m *velaSession) Addr() net.Addr               { return m.session.LocalAddr() }
func (m *velaSession) RemoteAddr() net.Addr         { return m.session.RemoteAddr() }
func (m *velaSession) IsClosed() bool               { return m.session.IsClosed() }
func (m *velaSession) Limit() rate.Limit            { return m.limiter.Limit() }
func (m *velaSession) SetLimit(bps rate.Limit) bool { return m.limiter.SetLimit(bps) }
func (m *velaSession) NumStreams() (int64, int64)   { return m.stats.numStreams() }
func (m *velaSession) Traffic() (uint64, uint64)    { return m.stats.mux.load() }
func (m *velaSession) Library() (string, string)    { return "smux", "github.com/xtaci/smux" }
func (m *velaSession) Streams() []Streamer          { return m.stats.actives() }
func (m *velaSession) ConnectedAt() time.Time       { return m.connAt }

func (m *velaSession) newConn(stm *smux.Stream, err error) (*velaConn, error) {
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancelCause(m.parent)
	limit := m.limiter.warpReadWriter(ctx, stm)

	conn := &velaConn{
		parent: m,
		stream: stm,
		limit:  limit,
		cancel: cancel,
	}
	m.stats.putConn(conn)

	return conn, nil
}
