module github.com/vela-ssoc/vela-tunnel/_example

go 1.20

require (
	github.com/vela-ssoc/vela-common-mba v0.0.0-20230719021516-03e1b06fa5c8
	github.com/vela-ssoc/vela-tunnel v0.0.0-20230828101517-469c00987832
)

require github.com/gorilla/websocket v1.5.0 // indirect

replace github.com/vela-ssoc/vela-tunnel => ../
