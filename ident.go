package tunnel

import (
	"encoding/json"
	"net"
	"time"

	"github.com/vela-ssoc/backend-common/encipher"
)

// Ident minion 节点握手认证时需要携带的信息，
type Ident struct {
	Semver     string    `json:"semver"`     // 节点版本
	Inet       net.IP    `json:"inet"`       // 内网出口 IP
	MAC        string    `json:"mac"`        // 出口 IP 所在网卡的 MAC 地址
	Goos       string    `json:"goos"`       // 操作系统 runtime.GOOS
	Arch       string    `json:"arch"`       // 操作系统架构 runtime.GOARCH
	CPU        int       `json:"cpu"`        // CPU 核心数
	PID        int       `json:"pid"`        // 进程 PID
	Workdir    string    `json:"workdir"`    // 工作目录
	Executable string    `json:"executable"` // 执行路径
	Username   string    `json:"username"`   // 当前操作系统用户名
	Hostname   string    `json:"hostname"`   // 主机名
	TimeAt     time.Time `json:"time_at"`    // 当前时间，暂无太大意义
}

// String fmt.Stringer
func (ident Ident) String() string {
	dat, _ := json.MarshalIndent(ident, "", "    ")
	return string(dat)
}

// encrypt 身份信息加密
func (ident Ident) encrypt() ([]byte, error) {
	return encipher.EncryptJSON(ident)
}
