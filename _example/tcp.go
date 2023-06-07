package main

import (
	"context"
	"log"
	"net"
	"net/url"

	"github.com/vela-ssoc/vela-common-mba/netutil"
	"github.com/vela-ssoc/vela-tunnel"
)

func ProxyTCP(local string, tun tunnel.Tunneler) error {
	tow := tcpOverWebsocket{local: local, tun: tun}
	return tow.Serve()
}

type tcpOverWebsocket struct {
	local string // 监听本地 0.0.0.0:8066
	tun   tunnel.Tunneler
}

func (tow *tcpOverWebsocket) Serve() error {
	lis, err := net.Listen("tcp", tow.local)
	if err != nil {
		return err
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			// TODO 对错误友好处理，这里是 demo 就写的简单粗暴了
			return err
		}
		go tow.serve(conn)
	}
}

func (tow *tcpOverWebsocket) serve(conn net.Conn) {
	//goland:noinspection GoUnhandledErrorResult
	defer conn.Close()

	query := url.Values{"address": []string{"tcp://61.129.129.241:22"}}
	path := "/api/v1/broker/stream/tunnel?" + query.Encode()
	stm, err := tow.tun.Stream(context.Background(), path, nil)
	if err != nil {
		log.Printf("stream 建立失败：%s", err)
		return
	}

	netutil.Pipe(conn, stm)
}
