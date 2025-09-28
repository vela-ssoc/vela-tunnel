package machine

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
)

// ID returns the platform specific machine id of the current host OS.
// Regard the returned id as "confidential" and consider using ProtectedID() instead.
func ID() (string, error) {
	id, err := machineID()
	if err != nil {
		return "", fmt.Errorf("machineid: %v", err)
	}
	return id, nil
}

// run wraps `exec.Command` with easy access to stdout and stderr.
func run(stdout, stderr io.Writer, cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Stdin = os.Stdin
	c.Stdout = stdout
	c.Stderr = stderr
	return c.Run()
}

func readFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func trim(s string) string {
	return strings.TrimSpace(strings.Trim(s, "\n"))
}

func NewGenerate(cachefile string) Generate {
	return Generate{
		cachefile: cachefile,
	}
}

type Generate struct {
	cachefile string
}

func (g Generate) MachineID(rebuild bool) string {
	if f := g.cachefile; f != "" && !rebuild {
		if raw, _ := os.ReadFile(f); len(raw) != 0 {
			return string(raw)
		}
	}

	return g.generateAndSaveCache()
}

func (g Generate) generateAndSaveCache() string {
	mid := g.generate()
	if f := g.cachefile; f != "" {
		_ = os.WriteFile(f, []byte(mid), 0644)
	}

	return mid
}

func (g Generate) generate() string {
	mid, _ := ID()
	hostname, _ := os.Hostname()
	networks := g.networks()
	addr := strings.Join(networks, ",")
	str := strings.Join([]string{mid, hostname, addr}, "|")
	sum := sha1.Sum([]byte(str))

	return hex.EncodeToString(sum[:])
}

func (g Generate) networks() []string {
	results := make([]string, 0, 10)
	faces, _ := net.Interfaces()
	for _, face := range faces {
		// 跳过换回网卡和未启用的网卡
		if face.Flags&net.FlagUp == 0 ||
			face.Flags&net.FlagLoopback != 0 {
			continue
		}

		var ips []string
		addrs, _ := face.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}

			// 过滤无效地址
			if ip == nil ||
				ip.IsLoopback() ||
				ip.IsMulticast() ||
				ip.IsUnspecified() {
				continue
			}

			if ip4 := ip.To4(); ip4 != nil {
				ips = append(ips, ip.String())
			} else if ip.To16() != nil {
				// 排除 IPv6 链路本地地址（fe80::/10），如不需要可移除此条件
				if ip.IsLinkLocalUnicast() {
					continue
				}
				ips = append(ips, ip.String())
			}
			if len(ips) != 0 {
				results = append(results, face.HardwareAddr.String())
				results = append(results, ips...)
			}
		}
	}

	return results
}
