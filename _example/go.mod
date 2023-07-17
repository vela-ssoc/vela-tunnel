module github.com/vela-ssoc/vela-tunnel/_example

go 1.20

require (
	github.com/vela-ssoc/vela-common-mba v0.0.0-20230714095322-ce45e2df93a2
	github.com/vela-ssoc/vela-tunnel v0.0.0-20230714082745-8bc61a0a9211
)

require github.com/gorilla/websocket v1.5.0 // indirect

replace github.com/vela-ssoc/vela-tunnel => ../
