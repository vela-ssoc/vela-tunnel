package tunnel

import (
	"encoding/json"

	"github.com/vela-ssoc/ssoc-common-mba/ciphertext"
)

// Issue 认证成功后服务端返回的必要信息
type Issue struct {
	ID     int64  `json:"id"`     // agent ID
	Passwd []byte `json:"passwd"` // 通信数据加密的密钥
}

// String fmt.Stringer
func (iss Issue) String() string {
	dat, _ := json.MarshalIndent(iss, "", "    ")
	return string(dat)
}

// decrypt 数据解密
func (iss *Issue) decrypt(data []byte) error {
	return ciphertext.DecryptJSON(data, iss)
}
