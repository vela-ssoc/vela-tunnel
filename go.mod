module github.com/vela-ssoc/vela-tunnel

go 1.24.0

require (
	github.com/gorilla/websocket v1.5.3
	github.com/vela-ssoc/vela-common-mba v0.0.0-20251210091356-7c0c9896a277
	golang.org/x/sys v0.41.0
)

// golang.org/x/sys v0.41.0 之后的版本要求 go1.25，ssoc 为了老系统的兼容，
// 编译版本是 go1.24，为了避免传染到 ssoc 模块编译，特此锁定。
replace golang.org/x/sys => golang.org/x/sys v0.41.0
