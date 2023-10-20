package tunnel

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vela-ssoc/vela-common-mba/netutil"
	"github.com/vela-ssoc/vela-common-mba/smux"
)

// borerTunnel 通道连接器
type borerTunnel struct {
	hide     Hide               // hide
	ident    Ident              // ident
	issue    Issue              // issue
	ntf      Notifier           // 事件通知
	interval time.Duration      // 心跳间隔
	dialer   dialer             // TCP 连接器
	coder    Coder              // JSON 编解码器
	brkAddr  *Address           // 当前连接的 broker 节点地址
	laddr    net.Addr           // socket 连接本地地址
	raddr    net.Addr           // socket 连接的远端地址
	muxer    *smux.Session      // 底层流复用
	client   netutil.HTTPClient // http 客户端
	stream   netutil.Streamer   // 建立流式通道用
	slog     Logger             // 日志输出组件
	parent   context.Context    // parent context.Context
	ctx      context.Context    // context.Context
	cancel   context.CancelFunc // context.CancelFunc
	// muxer    spdy.Muxer         // 底层流复用
}

// ID 节点 ID
func (bt *borerTunnel) ID() int64 {
	return bt.issue.ID
}

// Inet 出口网卡的 IP 地址
func (bt *borerTunnel) Inet() net.IP {
	return bt.ident.Inet
}

// Hide 配置信息
func (bt *borerTunnel) Hide() Hide {
	return bt.hide
}

// Ident 认证信息
func (bt *borerTunnel) Ident() Ident {
	return bt.ident
}

// Issue 中心端认证成功后返回的信息
func (bt *borerTunnel) Issue() Issue {
	return bt.issue
}

// BrkAddr 当前连接的 broker 地址
func (bt *borerTunnel) BrkAddr() *Address {
	return bt.brkAddr
}

func (bt *borerTunnel) LocalAddr() net.Addr {
	return bt.laddr
}

func (bt *borerTunnel) RemoteAddr() net.Addr {
	return bt.raddr
}

// NodeName 生成的节点名字
func (bt *borerTunnel) NodeName() string {
	return fmt.Sprintf("minion-%s-%d", bt.Inet(), bt.ID())
}

// Fetch 发送 HTTP 请求
func (bt *borerTunnel) Fetch(ctx context.Context, path string, rd io.Reader, header http.Header) (*http.Response, error) {
	return bt.fetch(ctx, http.MethodPost, path, rd, header)
}

// Oneway 单向请求，不关心返回的数据
func (bt *borerTunnel) Oneway(ctx context.Context, path string, rd io.Reader, header http.Header) error {
	res, err := bt.fetch(ctx, http.MethodPost, path, rd, header)
	if err == nil {
		return res.Body.Close()
	}
	return err
}

// JSON 发送的数据进行 json 序列化，返回的报文会 json 反序列化
func (bt *borerTunnel) JSON(ctx context.Context, path string, body, resp any) error {
	res, err := bt.fetchJSON(ctx, path, body)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer res.Body.Close()

	return bt.coder.NewDecoder(res.Body).Decode(resp)
}

// OnewayJSON 单向请求 json 数据，不关心返回数据
func (bt *borerTunnel) OnewayJSON(ctx context.Context, path string, req any) error {
	res, err := bt.fetchJSON(ctx, path, req)
	if err == nil {
		_ = res.Body.Close()
	}
	return err
}

// Attachment 下载文件
func (bt *borerTunnel) Attachment(parent context.Context, path string, timeouts ...time.Duration) (*Attachment, error) {
	if parent == nil {
		parent = context.Background()
	}
	timeout := 10 * time.Minute
	if len(timeouts) > 0 && timeouts[0] > 0 {
		timeout = timeouts[0]
	}

	ctx, cancel := context.WithTimeout(parent, timeout)
	res, err := bt.fetch(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		cancel()
		return nil, err
	}
	att := &Attachment{
		code:   res.StatusCode,
		body:   res.Body,
		cancel: cancel,
	}
	disposition := res.Header.Get("Content-Disposition")
	if _, params, _ := mime.ParseMediaType(disposition); params != nil {
		att.filename = params["filename"]
		att.hash = params["hash"]
	}

	return att, nil
}

// Stream 建立双向流
func (bt *borerTunnel) Stream(ctx context.Context, path string, header http.Header) (*websocket.Conn, error) {
	addr := bt.wsURL(path)
	conn, _, err := bt.stream.Stream(ctx, addr, header)
	return conn, err
}

// StreamConn 建立双向流
func (bt *borerTunnel) StreamConn(ctx context.Context, path string, header http.Header) (net.Conn, error) {
	ws, err := bt.Stream(ctx, path, header)
	if err != nil {
		return nil, err
	}
	conn := &websocketConn{ws: ws, rd: websocket.JoinMessages(ws, "")}
	return conn, nil
}

// Doer 带前缀的客户端
func (bt *borerTunnel) Doer(prefix string) Doer {
	prefix = strings.TrimRight(prefix, "/")
	return &tunnelDo{
		tun:    bt,
		prefix: prefix,
	}
}

func (bt *borerTunnel) fetchJSON(ctx context.Context, path string, req any) (*http.Response, error) {
	buf := new(bytes.Buffer)
	if err := bt.coder.NewEncoder(buf).Encode(req); err != nil {
		return nil, err
	}
	header := http.Header{
		"Content-Type": []string{"application/json; charset=utf-8"},
		"Accept":       []string{"application/json"},
	}
	return bt.fetch(ctx, http.MethodPost, path, buf, header)
}

func (bt *borerTunnel) fetch(ctx context.Context, method, path string, rd io.Reader, header http.Header) (*http.Response, error) {
	addr := bt.httpURL(path)
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}
	return bt.client.Fetch(ctx, method, addr, rd, header)
}

func (bt *borerTunnel) httpURL(path string) string {
	return bt.newURL("http", path)
}

func (bt *borerTunnel) wsURL(path string) string {
	return bt.newURL("ws", path)
}

// newURL 构造 URL
func (bt *borerTunnel) newURL(scheme, path string) string {
	sn := strings.SplitN(path, "?", 2)
	u := &url.URL{Scheme: scheme, Host: "soc", Path: sn[0]}
	if len(sn) == 2 {
		u.RawQuery = sn[1]
	}
	return u.String()
}

func (bt *borerTunnel) dialContext(context.Context, string, string) (net.Conn, error) {
	if stream, err := bt.muxer.OpenStream(); err != nil {
		return nil, err // 防止 *smux.Stream(nil)
	} else {
		return stream, nil
	}
}

func (bt *borerTunnel) heartbeat(inter time.Duration) {
	const maximum = 5      // 心跳连续错误次数
	timeout := time.Minute // 每次心跳包发送的超时时间
	var total uint64       // 心跳包发送失败总次数
	var sum int            // 心跳包发送失败连续次数
	var over bool          // 是否终止不再发送心跳包

	ticker := time.NewTicker(inter)
	defer ticker.Stop()

	for !over {
		select {
		case <-bt.parent.Done():
			over = true
		case <-ticker.C:
			err := bt.heartbeatSend(timeout)
			if err == nil {
				sum = 0 // 发送成功就将连续错误次数置为 0
				break
			}

			total++
			sum++
			if sum >= maximum {
				sum = 0
				bt.slog.Warnf("连续 %d 次（总共失败 %d 次）心跳包发送失败：%s，主动断开连接", sum, total, err)
				_ = bt.muxer.Close()
			} else {
				bt.slog.Warnf("心跳包连续第 %d 次（总共失败 %d 次）发送失败：%s", sum, total, err)
			}
		}
	}
}

func (bt *borerTunnel) heartbeatSend(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(bt.parent, timeout)
	defer cancel()

	return bt.Oneway(ctx, "/api/v1/minion/ping", nil, nil)
}

func (bt *borerTunnel) dial(parent context.Context) error {
	bt.parent = parent
	bt.ctx, bt.cancel = context.WithCancel(parent)
	start := time.Now()
	timeout := 5 * time.Second

	bt.slog.Infof("准备连接 broker ...")
	for {
		conn, addr, err := bt.dialer.iterDial(bt.ctx, timeout)
		if err != nil {
			du := bt.waitN(start)
			bt.slog.Warnf("连接 broker(%s) 发生错误: %s, %s 后重试", addr, err, du)
			if err = bt.parkN(du); err != nil {
				return err
			}
			continue
		}
		ctx, cancel := context.WithTimeout(bt.ctx, timeout)
		ident, issue, err := bt.handshake(ctx, conn, addr)
		cancel()
		if err == nil {
			bt.ident, bt.issue, bt.brkAddr = ident, issue, addr
			bt.laddr, bt.raddr = conn.LocalAddr(), conn.RemoteAddr()
			cfg := smux.DefaultConfig()
			cfg.Passwd = issue.Passwd
			bt.muxer = smux.Client(conn, cfg)
			bt.slog.Infof("连接 broker(%s) 成功", addr)
			return nil
		}

		_ = conn.Close()                                                    // 握手协商失败就关闭连接
		if exx, ok := err.(*netutil.HTTPError); ok && exx.NotAcceptable() { // NotAcceptable 代表节点已被删除
			return exx
		}

		du := bt.waitN(start)
		bt.slog.Warnf("与 broker(%s) 发生错误: %s, %s 后重试", addr, err, du)
		if err = bt.parkN(du); err != nil {
			return err
		}
	}
}

// handshake 握手协商
func (bt *borerTunnel) handshake(parent context.Context, conn net.Conn, addr *Address) (Ident, Issue, error) {
	inet := bt.localInet(conn.LocalAddr())
	mac := bt.dialer.lookupMAC(inet)

	ident := Ident{
		Semver:   bt.hide.Semver,
		Inet:     inet,
		MAC:      mac.String(),
		Goos:     runtime.GOOS,
		Arch:     runtime.GOARCH,
		CPU:      runtime.NumCPU(),
		PID:      os.Getpid(),
		Interval: bt.interval,
		TimeAt:   time.Now(),
		Unload:   bt.hide.Unload,
	}
	ident.Hostname, _ = os.Hostname()
	ident.Workdir, _ = os.Getwd()
	ident.Executable, _ = os.Executable()
	if cu, _ := user.Current(); cu != nil {
		ident.Username = cu.Username
	}

	var issue Issue
	enc, err := ident.encrypt()
	if err != nil {
		return ident, issue, err
	}

	body := bytes.NewReader(enc)
	req, err := bt.client.NewRequest(parent, http.MethodConnect, "/api/v1/minion", body, nil)
	if err != nil {
		return ident, issue, err
	}

	req.Host = addr.Name // 设置 Host
	if err = req.Write(conn); err != nil {
		return ident, issue, err
	}

	res, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return ident, issue, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer res.Body.Close()

	resp := make([]byte, 100*1024) // 100KiB
	code := res.StatusCode
	if code != http.StatusAccepted {
		n, _ := io.ReadFull(res.Body, resp)
		exr := &netutil.HTTPError{Code: code, Body: resp[:n]}
		return ident, issue, exr
	}

	n, err := io.ReadFull(res.Body, resp)
	if err == nil || err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
		err = issue.decrypt(resp[:n])
	}

	return ident, issue, err
}

func (*borerTunnel) localInet(addr net.Addr) net.IP {
	switch a := addr.(type) {
	case *net.TCPAddr:
		return a.IP
	case *net.UDPAddr:
		return a.IP
	case *net.IPNet:
		return a.IP
	case *net.IPAddr:
		return a.IP
	default:
		return nil
	}
}

// waitN 计算下次客户端重试等待间隔。
//
// 时长：0  3min 10min 30min        1h         12h                      ∞
// 图例：└──┴────┴───────┴──────────┴───────────┴───────────────────────┘
// 结果： 3s  10s   30s      1min        5min              10min
func (*borerTunnel) waitN(start time.Time) time.Duration {
	interval := time.Since(start)
	switch {
	case interval < 3*time.Minute:
		return 3 * time.Second
	case interval < 10*time.Minute:
		return 10 * time.Second
	case interval < 30*time.Minute:
		return 30 * time.Second
	case interval < time.Hour:
		return time.Minute
	case interval < 12*time.Hour:
		return 5 * time.Minute
	default:
		return 10 * time.Minute
	}
}

// parkN 协程休眠
func (bt *borerTunnel) parkN(du time.Duration) error {
	timer := time.NewTimer(du)
	defer timer.Stop()
	ctx := bt.ctx
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (bt *borerTunnel) serveHTTP(srv Server) {
	ctx := bt.parent
	ntf := bt.ntf
	gap := 5 * time.Second

	var err error
	for {
		before := time.Now()
		ln := bt.muxer
		err = srv.Serve(ln) // 如果连接正常则会阻塞在此
		bt.slog.Warnf("连接断开：%s", err)
		ntf.Disconnect(err) // 断开连接通知回调

		// 防止出现连接成功立马断开的情况，如果连接成功立马断开，间隔过短就歇一会再试。
		if du := gap - time.Since(before); du > time.Second {
			bt.slog.Warnf("稍等 %s 后重连", du)
			if err = bt.parkN(du); err != nil {
				bt.slog.Warnf("连接已经断开不再重连：%s", err)
				break
			}
		}

		bt.slog.Warnf("即将重连...")
		if err = bt.dial(ctx); err != nil {
			bt.slog.Warnf("重连失败退出：%s", err)
			break
		}
		bt.slog.Infof("重连成功")
		addr := bt.brkAddr
		ntf.Reconnected(addr) // 重连成功通知回调
	}

	bt.slog.Warnf("连接已经断开不再重连：%s", err)
	ntf.Shutdown(err)
}
