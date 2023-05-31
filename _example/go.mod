module github.com/vela-ssoc/vela-tunnel/_example

go 1.20

require (
	github.com/vela-ssoc/vela-common-mba v0.0.0-20230529100030-bcf504ceadba
	github.com/vela-ssoc/vela-tunnel v0.0.0-20230529123622-7b67aad716aa
)

require github.com/gorilla/websocket v1.5.0 // indirect

replace github.com/vela-ssoc/vela-tunnel => ../
