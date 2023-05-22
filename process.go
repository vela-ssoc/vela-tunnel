package tunnel

import (
	"context"
	"errors"
	"time"
)

type Processor interface {
	Substance(ctx context.Context, removes []int64, updates []*TaskChunk) ([]*TaskStatus, error)
	ThirdUpdate(ctx context.Context, id int64) error
	ThirdRemove(ctx context.Context, id int64) error
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
