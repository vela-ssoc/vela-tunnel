module github.com/vela-ssoc/vela-tunnel/_example

go 1.20

require (
	github.com/vela-ssoc/vela-common-mba v0.0.0-20230625070742-d83d8ab68906
	github.com/vela-ssoc/vela-tunnel v0.0.0-20230625101049-b22b89c756bf
)

require github.com/gorilla/websocket v1.5.0 // indirect

replace github.com/vela-ssoc/vela-tunnel => ../
