package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/vela-ssoc/vela-tunnel"
)

func NewServer() *http.Server {
	mux := http.NewServeMux()

	todo := &todoHandle{}
	prefix := "/api/v1"
	mux.HandleFunc(prefix+"/agent/task/status", todo.TaskStatus)
	mux.HandleFunc(prefix+"/agent/task/diff", todo.TaskDiff)
	mux.HandleFunc(prefix+"/agent/third/diff", todo.TaskThird)

	return &http.Server{Handler: mux}
}

type todoHandle struct{}

func (*todoHandle) TaskStatus(w http.ResponseWriter, r *http.Request) {
	log.Printf("[TODO] 上报 agent 节点的运行状态")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// TODO 采集任务运行状态并返回

	res := &tunnel.TaskReport{}
	_ = json.NewEncoder(w).Encode(res)
}

func (*todoHandle) TaskDiff(w http.ResponseWriter, r *http.Request) {
	var req tunnel.TaskDiff
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.NotModified() {
		return
	}

	log.Printf("[TODO] 执行配置差异并返回最新状态")

	// TODO 删除/执行配置后并采集最新的任务状态并返回
	buf := bytes.NewBufferString("变更配置，")
	if size := len(req.Removes); size != 0 {
		buf.WriteString(fmt.Sprintf("删除 %d 个配置 ID：", size))
		for _, id := range req.Removes {
			buf.WriteString(strconv.FormatInt(id, 10) + ", ")
		}
	}
	if size := len(req.Updates); size != 0 {
		buf.WriteString(fmt.Sprintf("更新 %d 个配置：", size))
		for _, up := range req.Updates {
			buf.WriteString(fmt.Sprintf("%s，", up.Name))
		}
	}
	log.Printf("配置差异：%s", buf)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	res := &tunnel.TaskReport{}
	_ = json.NewEncoder(w).Encode(res)
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
