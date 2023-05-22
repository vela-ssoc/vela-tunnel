package tunnel

import (
	"context"
	"errors"
	"time"
)

// Processor 中心端回调事件处理器
type Processor interface {
	// Substance 任务配置删除/修改时的处理器
	Substance(ctx context.Context, removes []int64, updates []*TaskChunk) ([]*TaskStatus, error)

	// ThirdUpdate 三方文件修改时会回调该接口
	ThirdUpdate(ctx context.Context, id int64) error

	// ThirdRemove 三方文件删除时会回调该接口
	ThirdRemove(ctx context.Context, id int64) error

	// Disconnect 连接关闭时的回调事件
	Disconnect(err error)

	// Reconnected 重连成功
	Reconnected()

	// Shutdown 连接遇到不可重试的错误，通道关闭程序结束。
	Shutdown(err error)
}

type TaskChunk struct {
	ID      int64  `json:"id,string"`
	Name    string `json:"name"`
	Dialect bool   `json:"dialect"`
	Hash    string `json:"hash"`
	Chunk   string `json:"chunk"`
}

type TaskStatus struct {
	ID      int64         `json:"id,string"`
	Dialect bool          `json:"dialect"`
	Name    string        `json:"name"`
	Status  string        `json:"status"`
	Hash    string        `json:"hash"`
	Uptime  time.Time     `json:"uptime"`
	From    string        `json:"from"`
	Runners []*TaskRunner `json:"runners"`
}

type TaskRunner struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

type noopProc struct{}

func (noopProc) Substance(ctx context.Context, removes []int64, updates []*TaskChunk) ([]*TaskStatus, error) {
	return nil, errors.New("non-implement substance event")
}

func (noopProc) ThirdUpdate(ctx context.Context, id int64) error {
	return errors.New("non-implement third update")
}

func (noopProc) ThirdRemove(ctx context.Context, id int64) error {
	return errors.New("non-implement third remove")
}

func (p noopProc) Disconnect(err error) {
}

func (p noopProc) Reconnected() {
}

func (p noopProc) Shutdown(err error) {
}
