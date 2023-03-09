package tunnel

import (
	"encoding/json"
	"io"
	"time"

	"github.com/vela-ssoc/backend-common/logback"
)

// JSONCoder 自定义 json 编解码器
type JSONCoder interface {
	// Marshal 将 struct 序列化为 JSON 并写入到 io.Writer 中
	Marshal(io.Writer, any) error

	// Unmarshal 流式读取 io.Reader 并反序列化为 struct
	Unmarshal(io.Reader, any) error
}

// Option 方法
type Option func(*option)

// option 参数
type option struct {
	coder    JSONCoder      // json 编解码器
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

func WithJSONCoder(coder JSONCoder) Option {
	return func(opt *option) {
		opt.coder = coder
	}
}

type stdJSON struct{}

func (stdJSON) Marshal(w io.Writer, a any) error   { return json.NewEncoder(w).Encode(a) }
func (stdJSON) Unmarshal(r io.Reader, a any) error { return json.NewDecoder(r).Decode(a) }
