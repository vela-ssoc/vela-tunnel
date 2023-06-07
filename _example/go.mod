module github.com/vela-ssoc/vela-tunnel/_example

go 1.20

require (
	github.com/vela-ssoc/vela-common-mba v0.0.0-20230607031609-7da38fca26ed
	github.com/vela-ssoc/vela-tunnel v0.0.0-20230605030553-3eddcf87f5ae
)

require github.com/gorilla/websocket v1.5.0 // indirect

replace github.com/vela-ssoc/vela-tunnel => ../
