package main

import (
	"context"
	"log"
	"net"

	"github.com/vela-ssoc/vela-common-mba/netutil"
	"github.com/vela-ssoc/vela-tunnel"
)

func ProxyTCP(local string, tun tunnel.Tunneler) error {
	tow := tcpOverWebsocket{local: local, tun: tun}
	return tow.Serve()
}

type tcpOverWebsocket struct {
	local string
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

	path := "/api/v1/broker/stream/tunnel?address=baidu.com:443"
	stm, err := tow.tun.Stream(context.Background(), path, nil)
	if err != nil {
		log.Printf("stream 建立失败：%s", err)
		return
	}

	netutil.ConnSockPIPE(conn, stm)
}
