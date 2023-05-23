package tunnel

import "time"

type TaskChunk struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Dialect bool   `json:"dialect"`
	Hash    string `json:"hash"`
	Chunk   []byte `json:"chunk"`
}

type TaskDiff struct {
	Removes []int64      `json:"removes"` // 需要删除的配置 ID
	Updates []*TaskChunk `json:"updates"` // 需要更新的配置信息
}

// NotModified 与中心端比对没有差异
func (td TaskDiff) NotModified() bool {
	return len(td.Removes) == 0 && len(td.Updates) == 0
}

type TaskReport struct {
	Tasks []*TaskStatus `json:"tasks"`
}

type TaskStatus struct {
	ID      int64         `json:"id"`      // 配置 ID 由中心端下发
	Dialect bool          `json:"dialect"` // 是否是私有配置，由中心端下发
	Name    string        `json:"name"`    // 配置名称，由中心端下发
	Status  string        `json:"status"`  // 运行状态
	Hash    string        `json:"hash"`    // 配置哈希（目前是 MD5）
	Uptime  time.Time     `json:"uptime"`  // 配置启动时间
	From    string        `json:"from"`    // 配置来源
	Cause   string        `json:"cause"`   // 错误信息
	Runners []*TaskRunner `json:"runners"` // 任务内部模块运行状态
}

type TaskRunner struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
}
