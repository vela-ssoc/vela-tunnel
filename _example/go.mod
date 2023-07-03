module github.com/vela-ssoc/vela-tunnel/_example

go 1.20

require (
	github.com/vela-ssoc/vela-common-mba v0.0.0-20230703055731-08f34aebb69e
	github.com/vela-ssoc/vela-tunnel v0.0.0-20230630070927-825b017a559b
)

require github.com/gorilla/websocket v1.5.0 // indirect

replace github.com/vela-ssoc/vela-tunnel => ../
