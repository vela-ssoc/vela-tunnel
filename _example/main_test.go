package main_test

import (
	"testing"

	"github.com/vela-ssoc/vela-tunnel"
)

func TestRead(t *testing.T) {
	raw, hide, err := tunnel.ReadHide("ssc-amd64-upx.exe")
	t.Log(raw)
	t.Log(hide)
	t.Log(err)
}
