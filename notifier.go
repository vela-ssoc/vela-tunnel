package tunnel

type Notifier interface {
	// Disconnect 连接关闭时的回调事件
	Disconnect(err error)

	// Reconnected 重连成功
	Reconnected(addr *Address)

	// Shutdown 连接遇到不可重试的错误，通道关闭程序结束。
	Shutdown(err error)
}

type emptyNotify struct{}

func (e emptyNotify) Disconnect(err error) {}

func (e emptyNotify) Reconnected(addr *Address) {}

func (e emptyNotify) Shutdown(err error) {}
