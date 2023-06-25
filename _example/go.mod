module github.com/vela-ssoc/vela-tunnel/_example

go 1.20

require (
	github.com/vela-ssoc/vela-common-mba v0.0.0-20230621095900-4f52d5f629a9
	github.com/vela-ssoc/vela-tunnel v0.0.0-20230621123722-9fdf6e19cd83
)

require github.com/gorilla/websocket v1.5.0 // indirect

replace github.com/vela-ssoc/vela-tunnel => ../
