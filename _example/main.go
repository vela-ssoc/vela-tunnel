package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vela-ssoc/vela-tunnel"
)

func main() {
	cares := []os.Signal{syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL, syscall.SIGINT}
	ctx, cancel := signal.NotifyContext(context.Background(), cares...)
	ntf := NewNotify(cancel)

	addr := &tunnel.Address{Addr: "172.31.61.168:8082"}
	hide := tunnel.Hide{
		Semver:   "0.0.1-example",
		Ethernet: tunnel.Addresses{addr},
	}

	srv := NewServer()
	tun, err := tunnel.Dial(ctx, hide, srv, tunnel.WithNotifier(ntf), tunnel.WithInterval(5*time.Second))
	if err != nil {
		log.Printf("tunnel 连接失败，结束运行：%v", err)
		return
	}
	name := tun.NodeName()
	ident, issue := tun.Ident(), tun.Issue()
	log.Printf("agent %s 连接成功！！！\nident：\n%s\nissue：\n%s\n", name, ident, issue)

	go func() {
		if exx := ProxyTCP("0.0.0.0:8066", tun); err != nil {
			log.Printf("TCP over websocket 代理出错：%s", exx)
		}
	}()

	go func() {
		time.Sleep(5 * time.Second)
		aaa := getAAA()
		tun.OnewayJSON(ctx, "/api/v1/broker/audit/risk", aaa)
	}()

	<-ctx.Done()
	log.Println("结束运行")
}

func getAAA() *AuditRiskRequest {
	return &AuditRiskRequest{
		Class:      "蜜罐应用",
		Level:      "高危",
		Payload:    "乒乒乓乓配",
		Subject:    "逐日",
		LocalIP:    "127.0.0.1",
		LocalPort:  8080,
		RemoteIP:   "3.3.3.3",
		RemotePort: 22336,
		FromCode:   "local.test",
		Region:     "",
		Reference:  "",
		Alert:      true,
		Time:       time.Now(),
	}
}

type AuditRiskRequest struct {
	// Class 风险类型
	// ["暴力破解", "病毒事件", "弱口令", "数据爬虫", "蜜罐应用", "web 攻击", "监控事件", "登录事件"]
	Class      string    `json:"class"       validate:"required"`
	Level      string    `json:"level"       validate:"required"` // 风险级别
	Payload    string    `json:"payload"`                         // 攻击载荷
	Subject    string    `json:"subject"     validate:"required"` // 风险事件主题
	LocalIP    string    `json:"local_ip"`                        // 本地 IP
	LocalPort  int       `json:"local_port"`                      // 本地端口
	RemoteIP   string    `json:"remote_ip"`                       // 远程 IP
	RemotePort int       `json:"remote_port"`                     // 远程端口
	FromCode   string    `json:"from_code"`                       // 来源模块
	Region     string    `json:"region"`                          // IP 归属地
	Reference  string    `json:"reference"`                       // 参考引用
	Alert      bool      `json:"alert"`                           // 是否需要发送告警
	Time       time.Time `json:"time"`                            // 风险产生的时间
}
