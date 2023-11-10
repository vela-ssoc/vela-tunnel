module github.com/vela-ssoc/vela-tunnel/_example

go 1.20

require (
	github.com/vela-ssoc/vela-common-mba v0.0.0-20231110031019-8401a542e951
	github.com/vela-ssoc/vela-tunnel v0.0.0-20231024095349-f7882f28425f
)

require (
	github.com/gorilla/websocket v1.5.1 // indirect
	golang.org/x/net v0.18.0 // indirect
)

replace github.com/vela-ssoc/vela-tunnel => ../
