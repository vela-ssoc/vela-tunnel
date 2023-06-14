package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/vela-ssoc/vela-tunnel"
)

func main() {
	cares := []os.Signal{syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL, syscall.SIGINT}
	ctx, cancel := signal.NotifyContext(context.Background(), cares...)
	ntf := NewNotify(cancel)

	addr := &tunnel.Address{Addr: "172.31.61.168:8082"}
	hide := tunnel.Hide{
		Semver:   "0.0.1-example",
		Ethernet: tunnel.Addresses{addr},
	}

	srv := NewServer()
	tun, err := tunnel.Dial(ctx, hide, srv, tunnel.WithNotifier(ntf), tunnel.WithInterval(time.Minute))
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

	doer := tun.Doer("/api/v1/forward/elastic")
	esc, err := elastic.NewClient(elastic.SetHttpClient(doer))
	log.Println(esc)
	log.Println(err)

	res, err := esc.Aliases().Do(context.Background())
	log.Println(res, err)

	<-ctx.Done()
	log.Println("结束运行")
}
