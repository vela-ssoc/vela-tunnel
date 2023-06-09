module github.com/vela-ssoc/vela-tunnel/_example

go 1.20

require (
	github.com/vela-ssoc/vela-common-mba v0.0.0-20230609032438-11353935e235
	github.com/vela-ssoc/vela-tunnel v0.0.0-20230609021943-6fd4e95e7efb
)

require github.com/gorilla/websocket v1.5.0 // indirect

replace github.com/vela-ssoc/vela-tunnel => ../
