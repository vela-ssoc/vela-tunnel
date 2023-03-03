package tunnel

import (
	"net"
	"strings"
)

// Hide 是 minion 节点的配置文件，正式发布时都会被隐写在二进制执行文件中，
// minion 启动时会读取自身文件中的隐写内容，解析出配置参数，所以叫 Hide。
// 注意：实际线上正式发布后，只能从自身读出隐写配置，强烈不建议使用开发模式
// 读取配置。
type Hide struct {
	Semver   string    `json:"semver"`   // 节点版本号
	Ethernet Addresses `json:"ethernet"` // 内网连接地址
	Internet Addresses `json:"internet"` // 外网连接地址
}

// Address broker 的服务地址
type Address struct {
	// TLS 服务端是否开启了 TLS
	TLS bool `json:"tls"`

	// Addr 服务端连接地址，格式为 域名、域名+端口、IP、IP+端口。
	// 如：ssoc.lan / ssoc.lan:8899 / 10.10.10.18 / 10.10.10.18:8443
	// 如果未写明端口，则开启 TLS 是 443，未开启则是 80
	Addr string `json:"addr"`

	// Name 是 TLS 证书验证的 servername，只有开启 TLS 连接时下才有作用
	Name string `json:"name"`

	eth bool
}

// String fmt.Stringer
func (ad Address) String() string {
	build := new(strings.Builder)
	if ad.eth {
		build.WriteString("eth ")
	}
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

// Format 对地址进行格式化处理，即：如果地址内有显式端口号，
// 则根据是否开启 TLS 补充默认端口号
func (ads Addresses) Format() {
	for _, ad := range ads {
		addr := ad.Addr
		_, port, err := net.SplitHostPort(addr)
		if err == nil && port != "" {
			continue
		}
		if ad.TLS {
			ad.Addr = addr + ":443"
		} else {
			ad.Addr = addr + ":80"
		}
	}
}
