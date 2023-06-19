package main_test

import (
	"context"
	"testing"

	"github.com/olivere/elastic/v7"

	"github.com/vela-ssoc/vela-tunnel"
)

func TestRead(t *testing.T) {
	raw, hide, err := tunnel.ReadHide("ssc-amd64-upx.exe")
	t.Log(raw)
	t.Log(hide)
	t.Log(err)
	var tun tunnel.Tunneler

	doer := tun.Doer("/api/v1/forward/elastic")
	cli, err := elastic.NewClient(elastic.SetHttpClient(doer))

	// demo 查询所有索引
	cli.Aliases().Do(context.Background())
}

func add(a, b int) int {
	return a + b
}
