package muxconn

import (
	"context"
	"net"
	"time"

	"golang.org/x/time/rate"
)

type Muxer interface {
	net.Listener

	// Open 开启一个虚拟连接。
	Open(context.Context) (net.Conn, error)

	// RemoteAddr 远端节点地址。
	RemoteAddr() net.Addr

	// IsClosed 通道是否关闭。
	IsClosed() bool

	// Limit 获取当前的传输速率限制。
	// bps = rate.Inf 代表不限制速率。
	Limit() (bps rate.Limit)

	// SetLimit 设置最大传输速率，bytes/s。
	// bps = rate.Inf 代表不限制速率。
	SetLimit(bps rate.Limit) bool

	Streams() []Streamer

	// NumStreams 返回累计创建的总连接数和当前活跃的连接数。
	NumStreams() (cumulative, active int64)

	// Traffic 通道总收发数据字节数。
	Traffic() (rx, tx uint64)

	// Library 底层连接所使用的库，方便排查调试。
	Library() (name, module string)

	// ConnectedAt Muxer 的创建时间，大致可以标识通道建立时间。
	ConnectedAt() time.Time
}

type Streamer interface {
	net.Conn
	Stats() *StreamStats
	setTrunkStats(*streamStats)
}
