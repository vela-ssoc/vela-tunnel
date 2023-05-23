package tunnel

// Notifier 通道 掉线/重连成功/关闭 事件通知器
type Notifier interface {
	// Disconnect 连接关闭时的回调事件
	Disconnect(err error)

	// Reconnected 重连成功
	Reconnected(addr *Address)

	// Shutdown 连接遇到不可重试的错误，通道关闭程序结束。
	Shutdown(err error)
}

type emptyNotify struct{}

func (e emptyNotify) Disconnect(error) {}

func (e emptyNotify) Reconnected(*Address) {}

func (e emptyNotify) Shutdown(error) {}
