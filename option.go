package tunnel

import (
	"time"

	"github.com/vela-ssoc/backend-common/logback"
)

// Option 方法
type Option func(*option)

// option 参数
type option struct {
	slog     logback.Logger // 日志输出组件
	interval time.Duration  // 心跳包发送间隔
}

// WithLogger 设置日志输出组件
func WithLogger(slog logback.Logger) Option {
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
