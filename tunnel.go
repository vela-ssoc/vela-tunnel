package tunnel

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vela-ssoc/vela-common-mba/netutil"
)

// Tunneler agent 节点与 broker 的连接器
type Tunneler interface {
	// ID minion 节点 ID
	ID() int64

	// Inet 出口 IP
	Inet() net.IP

	// Hide 数据
	Hide() Hide

	// Ident 连接中心端的认证信息
	Ident() Ident

	// Issue 中心端认证成功后返回的数据
	Issue() Issue

	// BrkAddr 当前连接成功的 broker 节点地址
	BrkAddr() *Address

	// NodeName 节点业务名称，部分地方可能会用到
	NodeName() string

	// Fetch 请求响应式调用
	Fetch(context.Context, string, io.Reader, http.Header) (*http.Response, error)

	// Oneway 单向调用，不在乎返回值
	Oneway(context.Context, string, io.Reader, http.Header) error

	// JSON 请求与响应均为 json
	JSON(context.Context, string, any, any) error

	// OnewayJSON 请求数据格式化为 json 后发送，不关心不解析返回数据
	OnewayJSON(context.Context, string, any) error

	// Attachment 文件附件下载
	Attachment(context.Context, string) (*Attachment, error)

	// Stream 建立双向流
	Stream(ctx context.Context, path string, header http.Header) (*websocket.Conn, error)
}

type Server interface {
	Serve(ln net.Listener) error
}

var ErrEmptyAddress = errors.New("内网地址与外网地址不能全部为空")

func Dial(parent context.Context, hide Hide, srv Server, opts ...Option) (Tunneler, error) {
	if len(hide.Ethernet) == 0 && len(hide.Internet) == 0 {
		return nil, ErrEmptyAddress
	}

	opt := new(option)
	for _, fn := range opts {
		fn(opt)
	}
	if opt.slog == nil {
		opt.slog = new(stdLog)
	}
	if opt.coder == nil {
		opt.coder = new(stdJSON)
	}
	if opt.ntf == nil {
		opt.ntf = new(emptyNotify)
	}
	if opt.interval > 0 && (opt.interval < time.Minute || opt.interval > time.Hour) {
		opt.interval = 3 * time.Minute
	}

	// 对地址预先处理
	hide.Ethernet.Preformat()
	hide.Internet.Preformat()
	dial := newDialer(hide.Ethernet, hide.Internet)
	bt := &borerTunnel{
		hide:     hide,
		dialer:   dial,
		ntf:      opt.ntf,
		slog:     opt.slog,
		coder:    opt.coder,
		interval: opt.interval,
	}

	bt.stream = netutil.NewStream(bt.dialContext)        // 创建 stream 连接器
	trip := &http.Transport{DialContext: bt.dialContext} // 创建 HTTP 客户端
	bt.client = netutil.NewClient(trip)

	if err := bt.dial(parent); err != nil {
		bt.slog.Infof("连接 broker 失败：%v", err)
		return nil, err
	}

	// 连接成功
	if inter := opt.interval; inter > 0 { // 是否开启心跳
		go bt.heartbeat(inter)
	}

	// 开启监听
	if srv == nil {
		srv = &http.Server{}
	}
	go bt.serveHTTP(srv)

	return bt, nil
}
