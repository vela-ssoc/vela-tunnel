module github.com/vela-ssoc/vela-tunnel/_example

go 1.20

require (
	github.com/vela-ssoc/vela-common-mba v0.0.0-20230625060802-676019c3eadf
	github.com/vela-ssoc/vela-tunnel v0.0.0-20230625013504-d084d79e3f1d
)

require github.com/gorilla/websocket v1.5.0 // indirect

replace github.com/vela-ssoc/vela-tunnel => ../
