package tunnel

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func NewMachineID(file string, log ...*slog.Logger) Identifier {
	dnd := &defaultNodeID{file: file}
	if len(log) > 0 {
		dnd.log = log[0]
	}

	return dnd
}

type defaultNodeID struct {
	file string
	log  *slog.Logger
}

func (dnd *defaultNodeID) MachineID(rebuild bool) string {
	if !rebuild {
		dnd.getLog().Info("准备从缓存中加载机器码")
		if mid := dnd.readFile(); mid != "" {
			dnd.getLog().Info("从缓存中读取到了机器码", "machine_id", mid)
			return mid
		}
		dnd.getLog().Warn("从缓存中未找到机器码")
	}

	dnd.getLog().Warn("准备计算机器码")
	hostid, _ := machineID()
	macs := dnd.hardwareAddrs()
	smac := strings.Join(macs, ",")

	args := []any{"mac", smac, "hostid", hostid}
	dnd.getLog().Info("得到设备信息", args...)

	input := strings.Join([]string{hostid, smac}, "-")
	sum := sha1.Sum([]byte(input))
	mid := hex.EncodeToString(sum[:])
	dnd.writeFile(mid) // 缓存机器码

	args = append(args, "machine_id", mid)
	dnd.getLog().Info("计算得到的机器码", args...)

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

func (dnd *defaultNodeID) getLog() *slog.Logger {
	if dnd.log != nil {
		return dnd.log
	}

	return slog.Default()
}

func (dnd *defaultNodeID) hardwareAddrs() []string {
	uniq := make(map[string]struct{}, 8)
	macs := make([]string, 0, 8)
	faces, _ := net.Interfaces()
	virtuals := virtualNetworks()
	for _, face := range faces {
		name := face.Name
		flags := face.Flags
		args := []any{"name", name}
		if (flags & net.FlagUp) == 0 {
			dnd.getLog().Info("跳过未启用的网卡", args...)
			continue
		}
		if (flags & net.FlagLoopback) != 0 {
			dnd.getLog().Info("跳过环回网卡", args...)
			continue
		}
		if (flags & net.FlagPointToPoint) != 0 {
			dnd.getLog().Info("跳过点对点网卡", args...)
			continue
		}
		hw := face.HardwareAddr
		if len(hw) == 0 {
			dnd.getLog().Info("跳过无 MAC 地址的网卡", args...)
			continue
		}
		zeroMAC := true
		for _, b := range hw {
			if b != 0 {
				zeroMAC = false
				break
			}
		}
		if zeroMAC {
			args = append(args, "mac", hw.String())
			dnd.getLog().Info("跳过 MAC 地址无效的网卡", args...)
			continue
		}
		if addrs, _ := face.Addrs(); dnd.withoutIPv4(addrs) {
			dnd.getLog().Info("跳过没有 IPv4 地址的网卡", args...)
			continue
		}

		// Locally administered address bit.
		// https://standards.ieee.org/wp-content/uploads/import/documents/tutorials/macgrp.pdf
		if (hw[0] & 0x02) != 0 {
			if dnd.isVirtualName(name) {
				dnd.getLog().Warn("网卡判定为虚拟网卡设备，不采用", args...)
				continue
			}

			dnd.getLog().Info("网卡判定为可用设备", args...)
		}
		// 排除虚拟网卡
		if yes := virtuals[name]; yes {
			dnd.getLog().Info("跳过虚拟网卡", args...)
			continue
		}

		args = append(args, "flag", flags.String())
		dnd.getLog().Info("有效的网卡", args...)
		mac := hw.String()
		if _, exists := uniq[mac]; !exists {
			uniq[mac] = struct{}{}
			macs = append(macs, mac)
		}
	}
	if len(macs) == 0 {
		dnd.getLog().Warn("没有找到有效的网卡")
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

func (dnd *defaultNodeID) isVirtualName(name string) bool {
	return strings.HasPrefix(name, "docker") ||
		strings.HasPrefix(name, "llw") ||
		strings.HasPrefix(name, "awdl")
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
