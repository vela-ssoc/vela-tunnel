module github.com/vela-ssoc/vela-tunnel/_example

go 1.20

require (
	github.com/vela-ssoc/vela-common-mba v0.0.0-20230830084957-da2ff0015ca5
	github.com/vela-ssoc/vela-tunnel v0.0.0-20231011011224-58f34b3ef227
)

require github.com/gorilla/websocket v1.5.0 // indirect

replace github.com/vela-ssoc/vela-tunnel => ../
