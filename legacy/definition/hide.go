package definition

import (
	"encoding/json"
	"time"
)

type MHide struct {
	// Servername 服务端域名。
	// 此处有两个作用：
	// 		TLS 连接下，用于证书认证校验。
	//
	Servername string `json:"servername"`

	// Addrs 代理节点地址，告诉 agent 客户端连接哪台代理节点。
	Addrs []string `json:"addrs"`

	// Semver agent 二进制发行版本。
	Semver string `json:"semver"`

	// Hash 文件原始哈希，不计算隐写数据。
	Hash string `json:"hash"`

	// Size 文件原始大小，不计算隐写数据。
	Size int64 `json:"size"`

	// Tags 首次下载时的隐写标签，
	// 例如：首次部署时，
	Tags []string `json:"tags"`

	// Goos 操作系统。
	Goos string `json:"goos"`

	// Arch CPU 架构。
	Arch string `json:"arch"`

	// Unload 是否开启静默模式，仅对新注册上线的节点有效。
	Unload bool `json:"unload"`

	// Unstable 是否不稳定版本。
	Unstable bool `json:"unstable"`

	// Customized 定制版本标记，为空代表标准版或叫通用版。
	Customized string `json:"customized"`

	// DownloadAt 下载时间。
	DownloadAt time.Time `json:"download_at"`

	// VIP 代理节点公网地址。
	//
	// Deprecated: use Addrs.
	VIP []string `json:"vip"`

	// LAN 代理节点内网地址。
	//
	// Deprecated: use Addrs.
	LAN []string `json:"lan"`

	// Edition 版本号。
	//
	// Deprecated: use Semver.
	Edition string `json:"edition"`
}

func (m MHide) String() string {
	raw, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(raw)
}
