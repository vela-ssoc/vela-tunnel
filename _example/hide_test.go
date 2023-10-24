package main

import (
	"testing"

	tunnel "github.com/vela-ssoc/vela-tunnel"
)

func TestName(t *testing.T) {
	hide, err := tunnel.ReadHide("ssc-worker.exe")
	t.Log(hide)
	t.Log(err)
}
