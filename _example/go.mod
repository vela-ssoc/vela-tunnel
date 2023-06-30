module github.com/vela-ssoc/vela-tunnel/_example

go 1.20

require (
	github.com/vela-ssoc/vela-common-mba v0.0.0-20230630011300-0e6c1d3908e2
	github.com/vela-ssoc/vela-tunnel v0.0.0-20230628082800-ab2c917b5f56
)

require github.com/gorilla/websocket v1.5.0 // indirect

replace github.com/vela-ssoc/vela-tunnel => ../
