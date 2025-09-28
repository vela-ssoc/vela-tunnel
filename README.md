# 安全平台 agent 节点通信连接

## Go 版本锁定

[Go1.24](https://go.dev/doc/go1.24#linux) 仅支持 kernel 3.2 或更高版本，为了保证兼容性、稳定性，`vela-tunnel` 将长期固定在 go1.24.x 版本。（2025-05-08）

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
	"github.com/vela-ssoc/vela-common-mba/definition"
	"github.com/vela-ssoc/vela-tunnel"
	"github.com/xgfone/ship/v5"
)

func main() {
	ctx := context.Background()
	// 测试环境自己手动输入参数
	{
		
		hide := definition.MHide{
			Semver:   "0.0.1-alpha",
			Addrs: []string{"ssoc-broker.example.com:8443"},
		}
	}

	// 正式环境读取隐写的数据
	{
		raw, hide, err := tunnel.ReadHide()
	}

	// 下面是常用的几种 HTTP 服务框架如何实现 tunnel.Server 接口的示例，
	// 根据实际情况任选其一。
	var server tunnel.Server
	server = fromFastHTTP()        // https://github.com/valyala/fasthttp
	server = fromFastHTTPRouter()  // https://github.com/fasthttp/router
	server = fromFastHTTPRouting() // https://github.com/qiangxue/fasthttp-routing
	server = fromFiber()           // https://github.com/gofiber/fiber
	server = fromShip()            // https://github.com/xgfone/ship
	server = fromGin()             // https://github.com/gin-gonic/gin
	server = fromStd()             // https://github.com/golang/go

	// tunnel.Dial 是阻塞式连接。如果连接失败，Dial 会一直重试直至成功或遇到不可重试的错误。
	// 也就是说当 tunnel.Dial 方法返回时，如果 err != nil 代表着该 agent 没有必要继续启动了。
	//
	// 不可重试的错误目前只有以下几种种情况：
	// 		1. 中心端将 agent 删除，agent 没有必要再运行了。
	// 		2. 参数配置错误
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

// 标准库：https://github.com/golang/go
func fromStd() tunnel.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello"))
	})
	srv := &http.Server{Handler: mux}
	return srv
}

```

## 自定义 json 编解码

```go
// sonicJSON 以 [sonic] 为例自定义实现 JSON 编解码器，
// 实现 tunnel.Coder 接口。
//
// [sonic]: https://github.com/bytedance/sonic
type sonicJSON struct {
    api sonic.API
}

func (s sonicJSON) NewEncoder(w io.Writer) interface{ Encode(any) error } { return s.api.NewEncoder(w) }
func (s sonicJSON) NewDecoder(r io.Reader) interface{ Decode(any) error } { return s.api.NewDecoder(r) }

coder := &sonicJSON{api: sonic.ConfigStd}
tun, err := tunnel.Dial(ctx, hide, srv, tunnel.WithCoder(coder))

```
