module github.com/vela-ssoc/vela-tunnel/_example

go 1.20

require (
	github.com/olivere/elastic/v7 v7.0.32
	github.com/vela-ssoc/vela-common-mba v0.0.0-20230613075657-284f14246a56
	github.com/vela-ssoc/vela-tunnel v0.0.0-20230614073201-8e45cf18d510
)

require (
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/pkg/errors v0.9.1 // indirect
)

replace github.com/vela-ssoc/vela-tunnel => ../
