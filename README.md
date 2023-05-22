# 安全平台 agent 节点通信连接

## [接口文档](https://vela-ssoc.github.io/vela-tunnel/)

## 代码示例

```go
package main

import (
	"context"

	"github.com/vela-ssoc/vela-tunnel"
)

func main() {
	// proc 要实现 tunnel.Processor 接口
	proc := new(velaProcess)
	ctx := context.Background()
	hide := tunnel.Hide{} // 自行注入参数
	tun, err := tunnel.Dial(ctx, hide, proc)
	if err != nil {
		panic(err)
	}

	// 连接成功
	// TODO 业务逻辑
}

type velaProcess struct{}

func (v velaProcess) Substance(ctx context.Context, removes []int64, updates []*tunnel.TaskChunk) ([]*tunnel.TaskStatus, error) {
	// TODO implement me
	panic("implement me")
}

func (v velaProcess) ThirdUpdate(ctx context.Context, id int64) error {
	// TODO implement me
	panic("implement me")
}

func (v velaProcess) ThirdRemove(ctx context.Context, id int64) error {
	// TODO implement me
	panic("implement me")
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
