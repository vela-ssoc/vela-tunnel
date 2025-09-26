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
	ident    Identifier    // 机器码生成器
	interval time.Duration // 心跳包发送间隔
}

// WithLogger 设置日志输出组件
func WithLogger(slog Logger) Option {
	return func(opt *option) {
		opt.slog = slog
	}
}

// WithInterval 设置心跳包间隔，如果不设置或该时间小于等于 0 则代表不发送心跳包。
// 心跳只是一种异常断开的兜底机制，由于生产环境节点较多，心跳间隔设置的太短也会给
// 中心端增加处理压力。
func WithInterval(interval time.Duration) Option {
	return func(opt *option) {
		opt.interval = interval
	}
}

// WithCoder 自定义 json 编解码器
func WithCoder(coder Coder) Option {
	return func(opt *option) {
		opt.coder = coder
	}
}

// WithNotifier tunnel 连接事件回调
func WithNotifier(ntf Notifier) Option {
	return func(opt *option) {
		opt.ntf = ntf
	}
}

// WithIdentifier 机器码生成器
func WithIdentifier(ident Identifier) Option {
	return func(opt *option) {
		opt.ident = ident
	}
}

type Coder interface {
	NewEncoder(io.Writer) interface{ Encode(any) error }
	NewDecoder(io.Reader) interface{ Decode(any) error }
}

// stdJSON 标准库 encoding/json 实现的 json 编解码器。
type stdJSON struct{}

func (stdJSON) NewEncoder(w io.Writer) interface{ Encode(any) error } { return json.NewEncoder(w) }
func (stdJSON) NewDecoder(r io.Reader) interface{ Decode(any) error } { return json.NewDecoder(r) }

// Identifier agent 身份唯一标识。
type Identifier interface {

	// MachineID 获取机器码，机器码是 agent 节点的唯一标识。
	//
	// 在实际环境中，业务方的会进行动态扩缩容，他们会用基础镜像克隆出多个实例，这个基础镜像可能
	// 已经运行过 ssoc-agent，机器码和环境已经初始化过了，而 agent 自身并不知道自己是克隆体，
	// 但是 ssoc 服务就会任务节点在重复连接而拒绝上线。
	//
	// 针对上述问题，设计一种机器码冲突避让策略：
	// 首先：生成机器码时一定要根据操作系统环境生成，不能用时间戳、UUID 等随机的参数作为机器码。
	// 例如：使用计算操作系统的 machine-id + hostname + mac + ip 哈希值作为机器码。即便是
	// 镜像克隆出来的机器，它们的 machine-id 一样，但是 hostname mac ip 不太可能一样，因为
	// 扩缩容的机器一般都同处一个局域网，如果 mac ip 一样，这台机器大概率无法正常联网工作的。
	//
	// 虽然有稳定的生成算法，生成机器码生成后要保存在本地磁盘，如果没有指定要 recreate 可以直接
	// 读取本地缓存的机器码，为什么要保存在本地呢？因为 agent 的 ip 可能是 DHCP，hostname 也
	// 可能被修改，但机器还是那台机器，如果不缓存每次都生成，久而久之会导致服务端留存大量无效节点。
	//
	// 总结：大致思路就是每次上线时如果服务端检测到重复连接，agent 就 recreate 重新生成机器码，
	// 每次上线时 recreate 至多会被调用一次，recreate 后的机器码可能还是原来的机器码，这说明
	// agent 环境没有发生变化。如果 recreate 后还是重复上线，也不会再次 recreate，一般说明存
	// 在其它问题。
	MachineID(recreate bool) string
}
