package main

import (
	"context"
	"log"

	"github.com/vela-ssoc/vela-tunnel"
)

func NewNotify(cancel context.CancelFunc) tunnel.Notifier {
	return &notify{
		cancel: cancel,
	}
}

type notify struct {
	cancel context.CancelFunc
}

func (n *notify) Disconnect(err error) {
	log.Printf("tunnel 断开了连接：%v", err)
}

func (n *notify) Reconnected(addr *tunnel.Address) {
	log.Printf("tunnel 断开了重连成功：%s", addr)
}

func (n *notify) Shutdown(err error) {
	log.Printf("tunnel 连接失败，遇到不可重试的错误，程序即将退出：%s", err)
	n.cancel()
}
