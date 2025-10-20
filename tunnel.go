package tunnel

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"os/user"
	"runtime"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vela-ssoc/vela-common-mba/definition"
	"github.com/vela-ssoc/vela-common-mba/netutil"
)

// Tunneler agent 节点与 broker 的连接器
type Tunneler interface {
	// ID minion 节点 ID
	//
	// Deprecated: 应用层不应该关心 Inet。
	ID() int64

	// Inet 出口 IP
	//
	// Deprecated: 应用层不应该关心 Inet。
	Inet() net.IP

	// Hide 数据
	//
	// Deprecated: 应用层不应该关心 LocalAddr。
	Hide() definition.MHide

	// Ident 连接中心端的认证信息
	//
	// Deprecated: 应用层不应该关心 Ident。
	Ident() Ident

	// Issue 中心端认证成功后返回的数据
	//
	// Deprecated: 应用层不应该关心 Issue。
	Issue() Issue

	// BrkAddr 当前连接成功的 broker 节点地址。
	//
	// Deprecated: 应用层不应该关心 LocalAddr。
	BrkAddr() *Address

	// LocalAddr 当前 socket 连接的本地地址，无连接则返回 nil。
	//
	// Deprecated: 应用层不应该关心 LocalAddr。
	LocalAddr() net.Addr

	// RemoteAddr 当前 socket 连接的远端地址，无连接则返回 nil
	//
	// Deprecated: 应用层不应该关心 LocalAddr。
	RemoteAddr() net.Addr

	// NodeName 节点业务名称，部分地方可能会用到。
	//
	// Deprecated: 此接口无实际意义，请不要依赖该接口。
	NodeName() string

	// Doer 发送请求
	//
	// Deprecated: 请基于 DialContext 包装一个 http.Client。
	// 	示例：
	//		cli := &http.Client{
	//			Transport: &http.Transport{
	//				DialContext: tun.DialContext,
	//			},
	//		}
	//
	//		cli.Do(req)
	Doer(prefix string) Doer

	// Fetch 请求响应式调用
	//
	// Deprecated: 请基于 DialContext 包装一个 http.Client。
	// 	示例：
	//		cli := &http.Client{
	//			Transport: &http.Transport{
	//				DialContext: tun.DialContext,
	//			},
	//		}
	//
	//		cli.Do(req)
	Fetch(context.Context, string, io.Reader, http.Header) (*http.Response, error)

	// Oneway 单向调用，不在乎返回值
	//
	// Deprecated: 请基于 DialContext 包装一个 http.Client。
	// 	示例：
	//		cli := &http.Client{
	//			Transport: &http.Transport{
	//				DialContext: tun.DialContext,
	//			},
	//		}
	//
	//		cli.Do(req)
	Oneway(context.Context, string, io.Reader, http.Header) error

	// JSON 请求与响应均为 json
	//
	// Deprecated: 请基于 DialContext 包装一个 http.Client。
	// 	示例：
	//		cli := &http.Client{
	//			Transport: &http.Transport{
	//				DialContext: tun.DialContext,
	//			},
	//		}
	//
	//		cli.Do(req)
	JSON(context.Context, string, any, any) error

	// OnewayJSON 请求数据格式化为 json 后发送，不关心不解析返回数据
	//
	// Deprecated: 请基于 DialContext 包装一个 http.Client。
	// 	示例：
	//		cli := &http.Client{
	//			Transport: &http.Transport{
	//				DialContext: tun.DialContext,
	//			},
	//		}
	//
	//		cli.Do(req)
	OnewayJSON(context.Context, string, any) error

	// Attachment 文件附件下载
	//
	// Deprecated: 请基于 DialContext 包装一个 http.Client。
	// 	示例：
	//		cli := &http.Client{
	//			Transport: &http.Transport{
	//				DialContext: tun.DialContext,
	//			},
	//		}
	//
	//		cli.Do(req)
	Attachment(context.Context, string, ...time.Duration) (*Attachment, error)

	// Stream 建立双向流
	//
	// Deprecated: 请基于 DialContext 自行实现。
	Stream(ctx context.Context, path string, header http.Header) (*websocket.Conn, error)

	// StreamConn 建立 net.Conn
	//
	// Deprecated: 请基于 DialContext 自行实现。
	StreamConn(ctx context.Context, path string, header http.Header) (net.Conn, error)

	DialContext(ctx context.Context, network, addr string) (net.Conn, error)
}

type Server interface {
	Serve(ln net.Listener) error
}

// Dial 建立与服务端的通道连接。
// 如果有网络不可达问题，该方法会一直重连直至成功，或者遇到不可重试的错误。
func Dial(parent context.Context, hide definition.MHide, srv Server, opts ...Option) (Tunneler, error) {
	addrs := hide.Addrs
	if len(addrs) == 0 {
		return nil, errors.New("地址不能为空")
	}

	if parent == nil {
		parent = context.Background()
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
	if opt.ident == nil {
		opt.ident = NewMachineID(".ssoc-machine-id")
	}
	// 心跳间隔小于等于 0 时代表关闭定时心跳，此时中心端不会对该节点定期心跳监控。
	// 如果该值大于 0，则有效值在 1min - 20min 之间，如果参数不在有效区间则自动改为 1min。
	// 如果设置了心跳，服务端 3 倍心跳间隔仍未收到该节点的任何数据包，则会强制断开 socket 连接。
	// 客户端发送心跳如果连续 n 次错误，也会自己主动断开连接。
	// 具体 n 是几，可以查看 borerTunnel.heartbeat 方法中的定义。
	if opt.interval > 0 && (opt.interval < time.Minute || opt.interval > 20*time.Minute) {
		opt.interval = time.Minute
	}

	// 对地址预先处理
	dial := newDialer(addrs, hide.Servername)
	bt := &borerTunnel{
		hide:     hide,
		dialer:   dial,
		ntf:      opt.ntf,
		mident:   opt.ident,
		slog:     opt.slog,
		coder:    opt.coder,
		interval: opt.interval,
		parent:   parent,
	}
	bt.ident = bt.initIdent(hide)
	bt.ident.Interval = bt.interval
	bt.ident.MachineID = bt.mident.MachineID(false)

	bt.stream = netutil.NewStream(bt.dialContext)        // 创建 stream 连接器
	trip := &http.Transport{DialContext: bt.dialContext} // 创建 HTTP 客户端
	bt.client = netutil.NewClient(trip)

	if err := bt.dial(); err != nil {
		bt.slog.Infof("连接 broker 失败：%v", err)
		return nil, err
	}

	// 连接成功后是否开启心跳
	if du := bt.interval; du > 0 { // 是否开启心跳
		go bt.heartbeat(du)
	}

	// 开启监听
	if srv == nil {
		srv = &http.Server{
			Handler: http.NotFoundHandler(),
		}
	}
	go bt.serveHTTP(srv)

	return bt, nil
}

func (bt *borerTunnel) initIdent(hide definition.MHide) Ident {
	ident := Ident{
		CPU:        runtime.NumCPU(),
		PID:        os.Getpid(),
		Goos:       hide.Goos,
		Arch:       hide.Arch,
		Semver:     hide.Semver,
		Unload:     hide.Unload,
		Unstable:   hide.Unstable,
		Customized: hide.Customized,
	}
	if ident.Goos == "" {
		ident.Goos = runtime.GOOS
	}
	if ident.Arch == "" {
		ident.Arch = runtime.GOARCH
	}
	ident.Workdir, _ = os.Getwd()
	ident.Executable, _ = os.Executable()
	ident.Hostname, _ = os.Hostname()
	if cu, _ := user.Current(); cu != nil {
		ident.Username = cu.Username
	}

	return ident
}
