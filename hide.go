package tunnel

import (
	"os"
	"strings"

	"github.com/vela-ssoc/vela-common-mba/ciphertext"
	"github.com/vela-ssoc/vela-common-mba/definition"
)

func ReadHide(filename ...string) (definition.MinionHide, error) {
	var name string
	if len(filename) > 0 && filename[0] != "" {
		name = filename[0]
	} else {
		name = os.Args[0]
	}

	var hide definition.MinionHide
	err := ciphertext.DecryptFile(name, &hide)

	return hide, err
}

// Address broker 的服务地址
type Address struct {
	// TLS 服务端是否开启了 TLS
	TLS bool `json:"tls" yaml:"tls"`

	// Addr 服务端地址，只需要填写地址或地址端口号，不需要路径
	// Example:
	//  	- example.com
	//  	- example.com:8080
	//		- 10.10.10.2
	// 		- 10.10.10.2:8443
	// 如果没有显式指明端口号，则开启 TLS 默认为 443，未开启 TLS 默认为 80
	Addr string `json:"addr" yaml:"addr"`

	// Name 主机名或 TLS SNI 名称
	// 无论是否开启 TLS，在发起 HTTP 请求时该 Name 都会被设置为 Host。
	// 当开启 TLS 时该 Name 会被设置为校验证书的 Servername。
	// 如果该字段为空，则默认使用 Addr 的地址作为主机名。
	Name string `json:"name" yaml:"name"`
}

// String fmt.Stringer
func (ad Address) String() string {
	build := new(strings.Builder)
	if ad.TLS {
		build.WriteString("tls://")
	} else {
		build.WriteString("tcp://")
	}
	build.WriteString(ad.Addr)

	if name := ad.Name; name != "" {
		build.WriteString("(name: ")
		build.WriteString(name)
		build.WriteByte(')')
	}

	return build.String()
}

// Addresses broker 地址切片
type Addresses []*Address
