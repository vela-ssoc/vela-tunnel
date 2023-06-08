module github.com/vela-ssoc/vela-tunnel/_example

go 1.20

require (
	github.com/vela-ssoc/vela-common-mba v0.0.0-20230608012308-0ced553011c7
	github.com/vela-ssoc/vela-tunnel v0.0.0-20230607090956-98d11d192266
)

require github.com/gorilla/websocket v1.5.0 // indirect

replace github.com/vela-ssoc/vela-tunnel => ../
