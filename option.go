package tunnel

import (
	"encoding/json"
	"io"
	"time"
)

// Option 方法
type Option func(*option)

// option 参数
type option struct {
	coder    Coder         // json 编解码器
	slog     Logger        // 日志输出组件
	ntf      Notifier      // 通道连接事件通知
	interval time.Duration // 心跳包发送间隔
}

// WithLogger 设置日志输出组件
func WithLogger(slog Logger) Option {
	return func(opt *option) {
		opt.slog = slog
	}
}

// WithInterval 设置心跳包间隔，如果不设置或该时间小于等于 0 则代表不发送心跳包
func WithInterval(interval time.Duration) Option {
	return func(opt *option) {
		opt.interval = interval
	}
}

func WithCoder(coder Coder) Option {
	return func(opt *option) {
		opt.coder = coder
	}
}

// WithNotifier 状态事件回调
func WithNotifier(ntf Notifier) Option {
	return func(opt *option) {
		opt.ntf = ntf
	}
}

type Coder interface {
	NewEncoder(io.Writer) interface{ Encode(any) error }
	NewDecoder(io.Reader) interface{ Decode(any) error }
}

type stdJSON struct{}

func (stdJSON) NewEncoder(w io.Writer) interface{ Encode(any) error } { return json.NewEncoder(w) }
func (stdJSON) NewDecoder(r io.Reader) interface{ Decode(any) error } { return json.NewDecoder(r) }
