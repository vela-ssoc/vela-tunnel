module github.com/vela-ssoc/vela-tunnel/_example

go 1.20

require (
	github.com/vela-ssoc/vela-common-mba v0.0.0-20230608065327-77f702856af6
	github.com/vela-ssoc/vela-tunnel v0.0.0-20230608071252-0fd220db2072
)

require github.com/gorilla/websocket v1.5.0 // indirect

replace github.com/vela-ssoc/vela-tunnel => ../
