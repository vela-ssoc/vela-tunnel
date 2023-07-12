module github.com/vela-ssoc/vela-tunnel/_example

go 1.20

require (
	github.com/vela-ssoc/vela-common-mba v0.0.0-20230712034230-0baa81b8b843
	github.com/vela-ssoc/vela-tunnel v0.0.0-20230712015101-c485ea619490
)

require github.com/gorilla/websocket v1.5.0 // indirect

replace github.com/vela-ssoc/vela-tunnel => ../
