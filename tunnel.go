package tunnel

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vela-ssoc/backend-common/httpclient"
	"github.com/vela-ssoc/backend-common/logback"
	"github.com/vela-ssoc/backend-common/netutil"
	"github.com/vela-ssoc/backend-common/opurl"
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

	// Listen 获取 net.Listener
	Listen() net.Listener

	// Reconnect 重连
	Reconnect(context.Context) error

	// Call 请求响应式调用
	Call(context.Context, opurl.URLer, io.Reader) (*http.Response, error)

	// Oneway 单向调用，不在乎返回值
	Oneway(context.Context, opurl.URLer, io.Reader) error

	// JSON 请求与响应均为 json
	JSON(context.Context, opurl.URLer, any, any) error

	// OnewayJSON 请求数据格式化为 json 后发送，不关心不解析返回数据
	OnewayJSON(context.Context, opurl.URLer, any) error

	// Attachment 文件附件下载
	Attachment(context.Context, opurl.URLer) (Attachment, error)

	// Stream 建立双向流
	Stream(opurl.URLer, http.Header) (*websocket.Conn, error)
}

var ErrEmptyAddress = errors.New("内网地址与外网地址不能全部为空")

func Dial(parent context.Context, hide Hide, opts ...Option) (Tunneler, error) {
	if len(hide.Ethernet) == 0 && len(hide.Internet) == 0 {
		return nil, ErrEmptyAddress
	}

	opt := new(option)
	for _, fn := range opts {
		fn(opt)
	}
	if opt.slog == nil {
		opt.slog = logback.Stdout()
	}
	if opt.interval > 0 && (opt.interval < 10*time.Second || opt.interval > time.Hour) {
		opt.interval = time.Minute
	}

	hide.Ethernet.Format()
	hide.Internet.Format()

	dial := newDialer(hide.Ethernet, hide.Internet)
	bt := &borerTunnel{
		hide:   hide,
		dialer: dial,
		slog:   opt.slog,
	}
	// 创建 stream 连接器
	bt.stream = netutil.Stream(bt.dialContext)
	// 创建 http 客户端
	tran := &http.Transport{DialContext: bt.dialContext}
	cli := &http.Client{Transport: tran}
	bt.client = httpclient.NewClient(cli)

	if err := bt.dial(parent); err != nil {
		bt.slog.Warnf("连接 broker 失败：%v", err)
		return nil, err
	}
	if inter := opt.interval; inter > 0 {
		go bt.heartbeat(inter)
	}

	return bt, nil
}
