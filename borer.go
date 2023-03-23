package tunnel

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/user"
	"runtime"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vela-ssoc/backend-common/httpx"
	"github.com/vela-ssoc/backend-common/logback"
	"github.com/vela-ssoc/backend-common/problem"
	"github.com/vela-ssoc/backend-common/spdy"
	"github.com/vela-ssoc/backend-common/transmit"
	"github.com/vela-ssoc/backend-common/transmit/opcode"
)

// borerTunnel 通道连接器
type borerTunnel struct {
	hide    Hide               // hide
	ident   Ident              // ident
	issue   Issue              // issue
	dialer  dialer             // TCP 连接器
	coder   JSONCoder          // JSON 编解码器
	brkAddr *Address           // 当前连接的 broker 节点地址
	muxer   spdy.Muxer         // 底层流复用
	client  transmit.Client    // http 客户端
	stream  transmit.Streamer  // 建立流式通道用
	slog    logback.Logger     // 日志输出组件
	parent  context.Context    // parent context.Context
	ctx     context.Context    // context.Context
	cancel  context.CancelFunc // context.CancelFunc
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

// Listen net.Listener
func (bt *borerTunnel) Listen() net.Listener {
	return bt.muxer
}

// NodeName 生成的节点名字
func (bt *borerTunnel) NodeName() string {
	return fmt.Sprintf("minion-%s-%d", bt.Inet(), bt.ID())
}

// Reconnect 断开连接并重连
func (bt *borerTunnel) Reconnect(ctx context.Context) error {
	_ = bt.muxer.Close()
	bt.cancel()
	if ctx == nil {
		ctx = bt.parent
	}
	return bt.dial(ctx)
}

// Fetch 发送 HTTP 请求
func (bt *borerTunnel) Fetch(ctx context.Context, op opcode.URLer, rd io.Reader, header http.Header) (*http.Response, error) {
	return bt.client.Fetch(ctx, op, rd, header)
}

// Oneway 单向请求，不关心返回的数据
func (bt *borerTunnel) Oneway(ctx context.Context, op opcode.URLer, rd io.Reader, header http.Header) error {
	res, err := bt.client.Fetch(ctx, op, rd, header)
	if err == nil {
		return res.Body.Close()
	}
	return err
}

// JSON 发送的数据进行 json 序列化，返回的报文会 json 反序列化
func (bt *borerTunnel) JSON(ctx context.Context, op opcode.URLer, body, resp any) error {
	res, err := bt.fetchJSON(ctx, op, body)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer res.Body.Close()

	return bt.coder.Unmarshal(res.Body, resp)
}

// OnewayJSON 单向请求 json 数据，不关心返回数据
func (bt *borerTunnel) OnewayJSON(ctx context.Context, op opcode.URLer, req any) error {
	res, err := bt.fetchJSON(ctx, op, req)
	if err == nil {
		_ = res.Body.Close()
	}
	return err
}

// Attachment 下载文件
func (bt *borerTunnel) Attachment(ctx context.Context, op opcode.URLer) (transmit.Attachment, error) {
	return bt.client.Attachment(ctx, op)
}

// Stream 建立双向流
func (bt *borerTunnel) Stream(op opcode.URLer, header http.Header) (*websocket.Conn, error) {
	conn, _, err := bt.stream.Stream(op, header)
	return conn, err
}

func (bt *borerTunnel) fetchJSON(ctx context.Context, op opcode.URLer, req any) (*http.Response, error) {
	buf := new(bytes.Buffer)
	if err := bt.coder.Marshal(buf, req); err != nil {
		return nil, err
	}
	header := http.Header{
		"Content-Type": []string{"application/json; charset=utf-8"},
		"Accept":       []string{"application/json"},
	}
	return bt.client.Fetch(ctx, op, buf, header)
}

func (bt *borerTunnel) dialContext(context.Context, string, string) (net.Conn, error) {
	return bt.muxer.Dial()
}

func (bt *borerTunnel) heartbeat(inter time.Duration) {
	ticker := time.NewTicker(inter)
	defer ticker.Stop()

over:
	for {
		select {
		case <-bt.parent.Done():
			break over
		case <-ticker.C:
			if err := bt.Oneway(nil, opcode.EndpointPing, nil, nil); err != nil {
				bt.slog.Warnf("心跳包发送出错：%v", err)
			}
		}
	}
}

func (bt *borerTunnel) dial(parent context.Context) error {
	bt.parent = parent
	bt.ctx, bt.cancel = context.WithCancel(parent)
	start := time.Now()

	bt.slog.Info("开始连接 broker ...")
	for {
		conn, addr, err := bt.dialer.iterDial(bt.ctx, 3*time.Second)
		if err != nil {
			du := bt.waitN(start)
			bt.slog.Warnf("连接 broker(%s) 发生错误: %s, %s 后重试", addr, err, du)
			if err = bt.sleepN(du); err != nil {
				return err
			}
			continue
		}
		ctx, cancel := context.WithTimeout(bt.ctx, 5*time.Second)
		ident, issue, err := bt.consult(ctx, conn, addr)
		cancel()
		if err == nil {
			bt.ident, bt.issue, bt.brkAddr = ident, issue, addr
			bt.muxer = spdy.Client(conn, spdy.WithEncrypt(issue.Passwd))
			bt.slog.Infof("连接 broker(%s) 成功", addr)
			return nil
		}

		if he, ok := err.(*httpx.Error); ok && he.NotAcceptable() {
			return he
		} else if pde, ok := err.(problem.Detail); ok && pde.Status == http.StatusNotAcceptable {
			return pde
		} else if pe, ok := err.(*problem.Detail); ok && pe.Status == http.StatusNotAcceptable {
			return pe
		}

		du := bt.waitN(start)
		bt.slog.Warnf("与 broker(%s) 发生错误: %s, %s 后重试", addr, err, du)
		if err = bt.sleepN(du); err != nil {
			return err
		}
	}
}

// consult 当建立好 TCP 连接后进行应用层协商
func (bt *borerTunnel) consult(parent context.Context, conn net.Conn, addr *Address) (Ident, Issue, error) {
	ip := conn.LocalAddr().(*net.TCPAddr).IP
	mac := bt.dialer.lookupMAC(ip)

	ident := Ident{
		Semver: bt.hide.Semver,
		Inet:   ip,
		MAC:    mac.String(),
		Goos:   runtime.GOOS,
		Arch:   runtime.GOARCH,
		CPU:    runtime.NumCPU(),
		PID:    os.Getpid(),
		TimeAt: time.Now(),
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
	req := bt.client.NewRequest(parent, opcode.EndpointMinion, body, nil)
	host := addr.Name
	if host == "" {
		host, _, _ = net.SplitHostPort(addr.Addr)
	}
	req.Host = host
	if err = req.Write(conn); err != nil {
		return ident, issue, nil
	}

	res, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return ident, issue, nil
	}
	//goland:noinspection GoUnhandledErrorResult
	defer res.Body.Close()

	resp := make([]byte, 10*1024)
	code := res.StatusCode
	if code != http.StatusAccepted {
		n, _ := io.ReadFull(res.Body, resp)
		pd := new(problem.Detail)
		if err = json.Unmarshal(resp[:n], pd); err == nil {
			return ident, issue, pd
		}
		exr := &httpx.Error{Code: code, Body: resp[:n]}
		return ident, issue, exr
	}

	n, _ := res.Body.Read(resp)
	err = issue.decrypt(resp[:n])

	return ident, issue, nil
}

// waitN 计算需要休眠多久
func (bt *borerTunnel) waitN(start time.Time) time.Duration {
	since := time.Since(start)
	du := time.Second
	switch {
	case since > 12*time.Hour:
		du = 10 * time.Minute
	case since > time.Hour:
		du = time.Minute
	case since > 30*time.Minute:
		du = 30 * time.Second
	case since > 10*time.Minute:
		du = 10 * time.Second
	case since > 3*time.Minute:
		du = 3 * time.Second
	}
	return du
}

// sleepN 协程休眠
func (bt *borerTunnel) sleepN(du time.Duration) error {
	timer := time.NewTimer(du)
	defer timer.Stop()
	var err error
	ctx := bt.ctx
	select {
	case <-timer.C:
	case <-ctx.Done():
		err = ctx.Err()
	}
	return err
}
