package tunnel

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/url"
	"os"
	"runtime"
	"time"

	"github.com/xtaci/smux"
)

func Open(ctx context.Context, cfg Config, opts ...Optioner) (Muxer, error) {
	if err := cfg.format(); err != nil {
		return nil, err
	}

	opts = append(opts, fallbackOption())
	opt := optionEval(opts...)

	req := &authRequest{
		PID:        os.Getpid(),
		Goos:       runtime.GOOS,
		Goarch:     runtime.GOARCH,
		Semver:     cfg.Semver,
		Unload:     cfg.Unload,
		Unstable:   cfg.Unstable,
		Customized: cfg.Customized,
	}
	req.Executable, _ = os.Executable()
	req.Workdir, _ = os.Getwd()
	req.Hostname, _ = os.Hostname()

	mux := new(safeMuxer)
	mc := &muxerClient{
		opt: opt,
		cfg: cfg,
		req: req,
		mux: mux,
		ctx: ctx,
	}

	if err := mc.open(); err != nil {
		return nil, err
	}

	ln := &muxerListener{mux: mux}
	go mc.serve(ln)

	return mux, nil
}

type muxerClient struct {
	opt     option
	cfg     Config
	req     *authRequest
	mux     *safeMuxer
	ctx     context.Context
	rebuild bool // 是否已经重新生成过机器码
}

func (mc *muxerClient) serve(ln net.Listener) {
	const sleep = 3 * time.Second

	mc.opt.notifier.Connected() // 首次连接成功回调函数。

	for {
		srv := mc.opt.server
		err := srv.Serve(ln)
		_ = ln.Close()

		attrs := []any{slog.Any("error", err), slog.Duration("timeout", sleep)}
		mc.log().Warn("agent 掉线了", attrs...)
		mc.opt.notifier.Disconnected(err) // 掉线回调函数。

		_ = mc.sleep(sleep)
		err = mc.open()
		if err != nil {
			mc.opt.notifier.Exited(err) // 退出回调函数。
			break
		}
		mc.opt.notifier.Reconnected() // 重连成功回调函数。
	}
}

func (mc *muxerClient) open() error {
	if mc.req.MachineID == "" {
		mc.req.MachineID = mc.opt.ident.MachineID(false)
	}

	var fails int
	startAt := time.Now()
	const timeout = 10 * time.Second
	addrs := mc.cfg.Addresses

	for {
		sess, _, err := mc.connects(addrs, timeout)
		if err == nil {
			mc.mux.store(sess)
			return nil
		}

		fails++
		sleep := mc.waitN(fails, startAt)
		mc.log().Warn("通道连接认证失败", "error", err, "sleep", sleep, "fails", fails, "addresses", addrs)
		if err = mc.sleep(sleep); err != nil {
			mc.log().Error("context 已取消，agent 隧道不再重连", "error", err, "fails", fails)
			return err
		}
	}
}

func (mc *muxerClient) connects(addrs []string, timeout time.Duration) (*smux.Session, *authResponse, error) {
	var errs []error
	for _, addr := range addrs {
		attrs := []any{slog.String("addr", addr), slog.Duration("timeout", timeout)}
		// 通过 websocket 拿到底层连接，并包装为 smux。
		sess, err := mc.openWebsocket(addr, timeout)
		if err != nil {
			errs = append(errs, err)
			attrs = append(attrs, slog.Any("error", err))
			mc.log().Warn("基础网络连接失败", attrs...)
			continue
		}

		resp, err1 := mc.authentication(sess, timeout)
		if err1 != nil {
			_ = sess.Close()
			errs = append(errs, err1)
			attrs = append(attrs, slog.Any("error", err1))
			mc.log().Warn("认证请求响应错误", attrs...)
			continue
		}

		attrs = append(attrs, slog.Any("agent_auth_response", resp))
		err2 := resp.checkError()
		if err2 == nil {
			mc.log().Info("通道连接认证成功", attrs...)
			return sess, resp, nil
		}

		_ = sess.Close()
		errs = append(errs, err2)
		attrs = append(attrs, slog.Any("error", err2))
		mc.log().Warn("认证失败", attrs...)

		if resp.duplicate() && !mc.rebuild {
			mc.rebuild = true
			before := mc.req.MachineID

			mc.log().Warn("当前机器码已经重复在线，准备 rebuild 机器码")
			after := mc.opt.ident.MachineID(true)
			mc.req.MachineID = after
			attrs = append(attrs, slog.String("before_machine_id", before), slog.String("after_machine_id", after))

			if before == after {
				mc.log().Info("前后机器码生成一致", attrs...)
			} else {
				mc.log().Warn("生成了新的机器码", attrs...)
			}
		}
	}

	return nil, nil, errors.Join(errs...)
}

func (mc *muxerClient) openWebsocket(addr string, timeout time.Duration) (*smux.Session, error) {
	ctx, cancel := context.WithTimeout(mc.ctx, timeout)
	defer cancel()

	dialer := mc.cfg.Dialer
	destURL := &url.URL{Scheme: "ws", Host: addr, Path: "/api/tunnel"}
	strURL := destURL.String()
	ws, _, err := dialer.DialContext(ctx, strURL, nil)
	if err != nil {
		return nil, err
	}

	conn := ws.NetConn()
	sess, err := smux.Client(conn, nil)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	return sess, nil
}

func (mc *muxerClient) log() *slog.Logger {
	if l := mc.opt.logger; l != nil {
		return l
	}

	return slog.Default()
}

func (mc *muxerClient) sleep(d time.Duration) error {
	ctx := mc.ctx
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

func (mc *muxerClient) waitN(fails int, startAt time.Time) time.Duration {
	if fails <= 30 {
		return 2 * time.Second
	} else if fails <= 100 {
		return 5 * time.Second
	} else if fails <= 200 {
		return 10 * time.Second
	}

	if du := time.Since(startAt); du <= 24*time.Hour {
		return 30 * time.Second
	}

	return time.Minute
}

func (mc *muxerClient) authentication(sess *smux.Session, timeout time.Duration) (*authResponse, error) {
	stm, err := sess.OpenStream()
	if err != nil {
		return nil, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer stm.Close()
	_ = stm.SetDeadline(time.Now().Add(timeout))

	req := mc.req
	switch adt := sess.LocalAddr().(type) {
	case *net.TCPAddr:
		req.Inet = adt.IP.String()
	case *net.UDPAddr:
		req.Inet = adt.IP.String()
	}
	if err = mc.writeRequest(stm, mc.req); err != nil {
		return nil, err
	}

	resp := new(authResponse)
	if err = mc.readResponse(stm, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (mc *muxerClient) writeRequest(stm *smux.Stream, v *authRequest) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	n := len(data)
	if n > 65535 {
		return io.ErrShortWrite
	}
	head := make([]byte, 4)
	binary.BigEndian.PutUint32(head, uint32(n))
	if _, err = stm.Write(head); err != nil {
		return err
	}
	_, err = stm.Write(data)

	return err
}

func (mc *muxerClient) readResponse(stm *smux.Stream, v any) error {
	head := make([]byte, 4)
	n, err := io.ReadFull(stm, head)
	if err != nil {
		return err
	} else if n != 4 {
		return io.ErrShortWrite
	}

	size := binary.BigEndian.Uint32(head)
	data := make([]byte, size)
	if n, err = io.ReadFull(stm, data); err != nil {
		return err
	} else if n != int(size) {
		return io.ErrShortWrite
	}

	return json.Unmarshal(data, v)
}
