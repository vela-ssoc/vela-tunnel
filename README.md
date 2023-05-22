# 安全平台 agent 节点通信连接

## [接口文档](https://vela-ssoc.github.io/vela-tunnel/)

## 代码示例

```go
package main

import (
	"context"
	"net"
	"net/http"

	"github.com/fasthttp/router"
	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
	"github.com/vela-ssoc/vela-tunnel"
	"github.com/xgfone/ship/v5"
)

func main() {
	ctx := context.Background()
	var hide tunnel.Hide

	var server tunnel.Server

	// 下面是常用的几种 HTTP 服务框架如何实现 tunnel.Server 接口的示例，
	// 根据实际情况任选其一。
	server = fromFastHTTP()        // https://github.com/valyala/fasthttp
	server = fromFastHTTPRouter()  // https://github.com/fasthttp/router
	server = fromFastHTTPRouting() // https://github.com/qiangxue/fasthttp-routing
	server = fromFiber()           // https://github.com/gofiber/fiber
	server = fromShip()            // https://github.com/xgfone/ship
	server = fromGin()             // https://github.com/gin-gonic/gin
	server = fromStd()             // https://github.com/golang/go

	tun, err := tunnel.Dial(ctx, hide, server)
}

// https://github.com/valyala/fasthttp/
func fromFastHTTP() tunnel.Server {
	fn := func(ctx *fasthttp.RequestCtx) {
		_, _ = ctx.WriteString("PONG")
	}

	srv := &fasthttp.Server{Handler: fn}

	return srv
}

// https://github.com/fasthttp/router
func fromFastHTTPRouter() tunnel.Server {
	r := router.New()
	r.GET("/ping", func(ctx *fasthttp.RequestCtx) {
		_, _ = ctx.WriteString("PONG")
	})

	srv := &fasthttp.Server{Handler: r.Handler}
	return srv
}

// https://github.com/qiangxue/fasthttp-routing
func fromFastHTTPRouting() tunnel.Server {
	r := routing.New()
	r.Get("/ping", func(c *routing.Context) error {
		_, err := c.WriteString("PONG")
		return err
	})

	srv := &fasthttp.Server{Handler: r.HandleRequest}

	return srv
}

// https://github.com/gofiber/fiber
func fromFiber() tunnel.Server {
	app := fiber.New()
	app.Get("/ping", func(ctx *fiber.Ctx) error {
		_, err := ctx.WriteString("PONG")
		return err
	})
	return &fiberServer{app: app}
}

type fiberServer struct {
	app *fiber.App
}

func (fb *fiberServer) Serve(ln net.Listener) error {
	return fb.app.Listener(ln)
}

// https://github.com/xgfone/ship
func fromShip() tunnel.Server {
	sh := ship.Default()
	sh.Route("/ping").GET(func(c *ship.Context) error {
		return c.Text(http.StatusOK, "PONG")
	})
	return &http.Server{Handler: sh}
}

// https://github.com/gin-gonic/gin
func fromGin() tunnel.Server {
	g := gin.Default()
	g.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "PONG")
	})
	srv := &http.Server{Handler: g}
	return srv
}

// 标准库
func fromStd() tunnel.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello"))
	})
	return new(http.Server)
}

```

## 自定义 json 编解码

```go
// sonicJSON 以 bytedance/sonic 为例实现 Coder 接口
type sonicJSON struct {
    api sonic.API
}

func (s sonicJSON) NewEncoder(w io.Writer) interface{ Encode(any) error } { return s.api.NewEncoder(w) }
func (s sonicJSON) NewDecoder(r io.Reader) interface{ Decode(any) error } { return s.api.NewDecoder(r) }

coder := &sonicJSON{api: sonic.ConfigStd}
tun, err := tunnel.Dial(ctx, hide, proc, tunnel.WithCoder(coder))

```
