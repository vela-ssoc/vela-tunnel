package tunnel

import (
	"encoding/json"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/vela-ssoc/vela-common-mba/ciphertext"
	"github.com/vela-ssoc/vela-common-mba/definition"
)

// Hide 是 minion 节点的配置文件，正式发布时都会被隐写在二进制执行文件中，
// minion 启动时会读取自身文件中的隐写内容，解析出配置参数，所以叫 Hide。
// 注意：实际线上正式发布后，只能从自身读出隐写配置，强烈不建议使用开发模式
// 读取配置。
type Hide struct {
	// Semver 版本号，要遵循 [SemVer] 语义化版本
	//
	// [SemVer]: https://semver.org/lang/zh-CN/
	Semver string `json:"semver"`

	// Unload 是否开启静默模式，仅在节点注册时有效
	Unload bool `json:"unload"`

	// Ethernet 内网连接地址
	Ethernet Addresses `json:"ethernet"`

	// Internet 外网连接地址
	Internet Addresses `json:"internet"`
}

func (h Hide) String() string {
	raw, _ := json.MarshalIndent(h, "", "  ")
	return string(raw)
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

	// eth 是否是内网配置
	eth bool
}

func (ad Address) Ethernet() bool {
	return ad.eth
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

// Preformat 对地址进行格式化处理，即：如果地址内有显式端口号，
// 则根据是否开启 TLS 补充默认端口号
func (ads Addresses) Preformat() {
	for _, ad := range ads {
		addr := ad.Addr
		host, port, err := net.SplitHostPort(addr)
		if err == nil && port != "" {
			if ad.Name == "" {
				ad.Name = host
			}
			continue
		}
		if ad.Name == "" {
			ad.Name = addr
		}
		if ad.TLS {
			ad.Addr = addr + ":443"
		} else {
			ad.Addr = addr + ":80"
		}
	}
}

type RawHide definition.MinionHide

func (h RawHide) String() string {
	return definition.MinionHide(h).String()
}

func ReadHide(names ...string) (RawHide, Hide, error) {
	name := os.Args[0]
	if len(names) != 0 && names[0] != "" {
		name = names[0]
	}

	var raw RawHide
	var hide Hide
	if err := ciphertext.DecryptFile(name, &raw); err != nil {
		return raw, hide, err
	}

	// 将老的转为新的
	servername := raw.Servername
	hide.Semver = raw.Edition
	hide.Unload = raw.Unload
	for _, s := range raw.LAN {
		addr := parseURL(s, servername)
		hide.Ethernet = append(hide.Ethernet, addr)
	}
	for _, s := range raw.VIP {
		addr := parseURL(s, servername)
		hide.Internet = append(hide.Internet, addr)
	}

	return raw, hide, nil
}

func parseURL(rawURL, servername string) *Address {
	u, err := url.Parse(rawURL)
	if err != nil {
		return &Address{Addr: rawURL}
	}

	if servername == "" {
		sn, _, _ := net.SplitHostPort(u.Host)
		if sn == "" {
			servername = u.Host
		} else {
			servername = sn
		}
	}

	return &Address{
		TLS:  u.Scheme == "wss",
		Addr: u.Host,
		Name: servername,
	}
}
