package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vela-ssoc/vela-common-mba/definition"
	"github.com/vela-ssoc/vela-tunnel"
)

func main() {
	cares := []os.Signal{syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL, syscall.SIGINT}
	ctx, cancel := signal.NotifyContext(context.Background(), cares...)
	ntf := NewNotify(cancel)

	hide := definition.MinionHide{
		Servername: "vela-ssoc-inline-yonghe.eastmoney.com",
		LAN:        []string{"10.228.162.244:1433"},
		Edition:    "0.0.0-unknown",
	}

	srv := NewServer()
	tun, err := tunnel.Dial(ctx, hide, srv, tunnel.WithNotifier(ntf), tunnel.WithInterval(5*time.Second))
	if err != nil {
		log.Printf("tunnel 连接失败，结束运行：%v", err)
		return
	}
	name := tun.NodeName()
	ident, issue := tun.Ident(), tun.Issue()
	log.Printf("agent %s 连接成功！！！\nident：\n%s\nissue：\n%s\n", name, ident, issue)

	go func() {
		if exx := ProxyTCP("0.0.0.0:8066", tun); err != nil {
			log.Printf("TCP over websocket 代理出错：%s", exx)
		}
	}()

	<-ctx.Done()
	log.Println("结束运行")
}
