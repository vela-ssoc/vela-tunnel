package muxconn

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type DialWebsocketOptions struct {
	// Dialer websocket 拨号器。
	Dialer *websocket.Dialer

	// PerTimeout 每次连接的超时时间。
	PerTimeout time.Duration

	// Logger 日志输出。
	Logger *slog.Logger
}

func Dial(parent context.Context, addrs []string, opts *DialWebsocketOptions) (Muxer, error) {
	if opts == nil {
		opts = new(DialWebsocketOptions)
	}

	return opts.dialContext(parent, addrs)
}

func (d DialWebsocketOptions) dialContext(parent context.Context, addrs []string) (Muxer, error) {
	if len(addrs) == 0 {
		return nil, errors.New("连接地址不能为空")
	}

	var errs []error
	for _, addr := range addrs {
		pu, err := url.Parse(addr)
		if err != nil {
			errs = append(errs, err)
			d.log().Warn("地址格式错误", "addr", addr, "err", err)
			continue
		}
		switch pu.Scheme {
		case "http":
			pu.Scheme = "ws"
		case "https":
			pu.Scheme = "wss"
		}

		mux, err2 := d.dialServer(parent, pu)
		if err2 != nil {
			errs = append(errs, err2)
			d.log().Warn("连接失败", "addr", addr, "err", err2)
			continue
		}

		proto, module := mux.Library()
		d.log().Info("连接成功", "addr", addr, "protocol", proto, "module", module)

		return mux, nil
	}

	err := errors.Join(errs...)
	d.log().Error("全部连接失败", "addr", addrs, "err", err)

	return nil, err
}

func (d DialWebsocketOptions) dialServer(parent context.Context, dest *url.URL) (Muxer, error) {
	strURL := dest.String()
	dialer := d.websocketDialer()

	ctx, cancel := context.WithTimeout(parent, d.perTimeout())
	defer cancel()

	ws, _, err := dialer.DialContext(ctx, strURL, nil)
	if err != nil {
		return nil, err
	}
	subprotocol := ws.Subprotocol()
	if subprotocol == "" {
		_ = ws.Close()
		return nil, websocket.ErrBadHandshake
	}

	var mux Muxer
	conn := ws.NetConn()
	switch subprotocol {
	case "yamux":
		mux, err = NewYaMUX(parent, conn, nil, false)
	case "vela":
		mux, err = NewVela(parent, conn, nil, false)
	default:
		mux, err = NewSMUX(parent, conn, nil, false)
	}
	if err != nil {
		_ = ws.Close()
		return nil, err
	}

	return mux, nil
}

func (d DialWebsocketOptions) log() *slog.Logger {
	if d.Logger != nil {
		return d.Logger
	}

	return slog.Default()
}

func (d DialWebsocketOptions) websocketDialer() *websocket.Dialer {
	if d.Dialer != nil {
		return d.Dialer
	}

	return &websocket.Dialer{
		Subprotocols: []string{"smux", "yamux", "vela"},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
}

func (d DialWebsocketOptions) perTimeout() time.Duration {
	if d.PerTimeout > 0 {
		return d.PerTimeout
	}

	return 10 * time.Second
}
