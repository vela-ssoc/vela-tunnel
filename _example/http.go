package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func NewServer() *http.Server {
	mux := http.NewServeMux()

	todo := &todoHandle{}
	prefix := "/api/v1"
	mux.HandleFunc(prefix+"/agent/task/status", todo.TaskStatus)
	mux.HandleFunc(prefix+"/agent/task/diff", todo.TaskDiff)
	mux.HandleFunc(prefix+"/agent/third/diff", todo.TaskThird)
	mux.HandleFunc(prefix+"/agent/notice/command", todo.Command)
	mux.HandleFunc(prefix+"/agent/notice/upgrade", todo.Upgrade)
	mux.HandleFunc(prefix+"/agent/startup", todo.Startup)

	return &http.Server{Handler: mux}
}

type todoHandle struct{}

func (*todoHandle) TaskStatus(w http.ResponseWriter, r *http.Request) {
	log.Printf("[TODO] 上报 agent 节点的运行状态")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// TODO 采集任务运行状态并返回
}

func (*todoHandle) TaskDiff(w http.ResponseWriter, r *http.Request) {
	log.Printf("[TODO] 执行配置差异并返回最新状态")

	// TODO 删除/执行配置后并采集最新的任务状态并返回
}

func (*todoHandle) TaskThird(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string `json:"name"`
		Event string `json:"event"` // update-更新 delete-删除
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	// TODO 收到事件指令后。先查看本 agent 是否用到了该这个三方文件，
	// 	如果没有用到可以忽略该事件。如果用到了在根据 event 做相应的事件处理。

	log.Printf("[TODO] 三方文件 %s 发生了 %s 事件", req.Name, req.Event)

	w.WriteHeader(http.StatusOK)
}

func (*todoHandle) Upgrade(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Semver string `json:"semver"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	log.Printf("[TODO] agent 检查自身二进制是否需要升级：%s", req.Semver)
	w.WriteHeader(http.StatusOK)
}

func (h *todoHandle) Command(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Cmd string `json:"cmd"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	log.Printf("[TODO] 收到 %s 命令事件", req.Cmd)

	w.WriteHeader(http.StatusOK)
}

func (h *todoHandle) Startup(w http.ResponseWriter, r *http.Request) {
	log.Print("[TODO] startup 发生了变更")
}
