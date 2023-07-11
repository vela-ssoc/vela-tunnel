module github.com/vela-ssoc/vela-tunnel/_example

go 1.20

require (
	github.com/vela-ssoc/vela-common-mba v0.0.0-20230711072420-a2f072a189f3
	github.com/vela-ssoc/vela-tunnel v0.0.0-20230711074322-b23a329d3a86
)

require github.com/gorilla/websocket v1.5.0 // indirect

replace github.com/vela-ssoc/vela-tunnel => ../
