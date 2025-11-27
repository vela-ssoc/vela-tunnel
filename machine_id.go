package tunnel

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"net"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func NewMachineID(file string, log ...Logger) Identifier {
	dnd := &defaultNodeID{file: file}
	if len(log) > 0 {
		dnd.log = log[0]
	}

	return dnd
}

type defaultNodeID struct {
	file string
	log  Logger
}

func (dnd *defaultNodeID) MachineID(rebuild bool) string {
	if !rebuild {
		dnd.getLog().Infof("准备从缓存中加载机器码")
		if mid := dnd.readFile(); mid != "" {
			dnd.getLog().Infof("从缓存中读取到了机器码 " + mid)
			return mid
		}
		dnd.getLog().Warnf("从缓存中未找到机器码")
	}

	dnd.getLog().Warnf("准备计算机器码")
	hostid, _ := machineID()
	hostname, _ := os.Hostname()
	macs := dnd.hardwareAddrs()
	smac := strings.Join(macs, ",")
	dnd.getLog().Infof("得到设备信息 hostid=" + hostid + ", hostname=" + hostname + ", mac=" + smac)

	input := strings.Join([]string{hostid, hostname, smac}, "-")
	sum := sha1.Sum([]byte(input))
	mid := hex.EncodeToString(sum[:])
	dnd.writeFile(mid) // 缓存机器码
	dnd.getLog().Infof("计算得到的机器码为 " + mid)

	return mid
}

func (dnd *defaultNodeID) readFile() string {
	if dnd.file == "" {
		return ""
	}
	dat, _ := os.ReadFile(dnd.file)

	return string(dat)
}

func (dnd *defaultNodeID) writeFile(mid string) {
	if dnd.file != "" {
		_ = os.WriteFile(dnd.file, []byte(mid), 0o600)
	}
}

func (dnd *defaultNodeID) getLog() Logger {
	if dnd.log != nil {
		return dnd.log
	}

	return new(discordLog)
}

func (dnd *defaultNodeID) hardwareAddrs() []string {
	uniq := make(map[string]struct{}, 8)
	macs := make([]string, 0, 8)
	faces, _ := net.Interfaces()
	virtuals := virtualNetworks()
	for _, face := range faces {
		name := face.Name
		flags := face.Flags
		if (flags & net.FlagUp) == 0 {
			dnd.getLog().Infof("跳过未启用的网卡 " + name)
			continue
		}
		if (flags & net.FlagLoopback) != 0 {
			dnd.getLog().Infof("跳过环回网卡 " + name)
			continue
		}
		if (flags & net.FlagPointToPoint) != 0 {
			dnd.getLog().Infof("跳过点对点网卡 " + name)
			continue
		}
		hw := face.HardwareAddr
		if len(hw) == 0 {
			dnd.getLog().Infof("跳过无 MAC 地址的网卡 " + name)
			continue
		}
		zeroMAC := true
		for _, b := range hw {
			if zeroMAC = b != 0; !zeroMAC {
				break
			}
		}
		if zeroMAC {
			dnd.getLog().Infof("跳过无 MAC 地址的网卡 " + name)
			continue
		}
		// Locally administered address bit.
		// https://standards.ieee.org/wp-content/uploads/import/documents/tutorials/macgrp.pdf
		if len(hw) > 0 && (hw[0]&0x02) != 0 {
			if strings.HasPrefix(name, "docker") || // Docker
				strings.HasPrefix(name, "llw") || // Low-Latency Wireless
				strings.HasPrefix(name, "awdl") { // Apple Wireless Direct Link (AWDL)
				dnd.getLog().Infof("跳过疑似虚拟网卡 " + name)
				continue
			}
		}

		if addrs, _ := face.Addrs(); dnd.withoutIPv4(addrs) {
			dnd.getLog().Infof("跳过无 IPv4 地址的网卡 " + name)
			continue
		}

		// 排除虚拟网卡
		if yes := virtuals[name]; yes {
			dnd.getLog().Infof("跳过虚拟网卡 " + name)
			continue
		}

		dnd.getLog().Infof("有效的网卡 " + name + " Flag=" + flags.String())
		mac := hw.String()
		if _, exists := uniq[mac]; !exists {
			uniq[mac] = struct{}{}
			macs = append(macs, mac)
		}
	}
	sort.Strings(macs)

	return macs
}

func (dnd *defaultNodeID) withoutIPv4(addrs []net.Addr) bool {
	for _, addr := range addrs {
		inet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		ip := inet.IP.To4()
		if ip != nil {
			return false
		}
	}

	return true
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
