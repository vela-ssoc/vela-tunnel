module github.com/vela-ssoc/vela-tunnel/_example

go 1.20

require (
	github.com/vela-ssoc/vela-common-mba v0.0.0-20231129020857-de2b2be3073a
	github.com/vela-ssoc/vela-tunnel v0.0.0-20231114083858-0bba503285cd
)

require (
	github.com/gorilla/websocket v1.5.1 // indirect
	golang.org/x/net v0.19.0 // indirect
)

replace github.com/vela-ssoc/vela-tunnel => ../
