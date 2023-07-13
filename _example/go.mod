module github.com/vela-ssoc/vela-tunnel/_example

go 1.20

require (
	github.com/vela-ssoc/vela-common-mba v0.0.0-20230713115403-0429001948f8
	github.com/vela-ssoc/vela-tunnel v0.0.0-20230712045958-d443c918c4e2
)

require github.com/gorilla/websocket v1.5.0 // indirect

replace github.com/vela-ssoc/vela-tunnel => ../
