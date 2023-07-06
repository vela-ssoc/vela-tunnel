module github.com/vela-ssoc/vela-tunnel/_example

go 1.20

require (
	github.com/vela-ssoc/vela-common-mba v0.0.0-20230706050807-99f8ad5a1a39
	github.com/vela-ssoc/vela-tunnel v0.0.0-20230703061942-d78044cab433
)

require github.com/gorilla/websocket v1.5.0 // indirect

replace github.com/vela-ssoc/vela-tunnel => ../
