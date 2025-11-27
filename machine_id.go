package tunnel

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"net"
	"os"
	"os/exec"
	"slices"
	"sort"
	"strings"
)

func NewIdent(file string, log Logger) Identifier {
	return &defaultNodeID{
		file: file,
		log:  log,
	}
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

	return new(stdLog)
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
		zero := true
		for _, b := range hw {
			if b != 0 {
				zero = false
				break
			}
		}
		if zero {
			dnd.getLog().Infof("跳过无 MAC 地址的网卡 " + name)
			continue
		}
		if len(hw) > 0 && (hw[0]&0x02) != 0 {
			// Locally administered address bit.
			// https://standards.ieee.org/wp-content/uploads/import/documents/tutorials/macgrp.pdf
			dnd.getLog().Infof("跳过疑似虚拟网卡 " + name)
			continue
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

func NewMachineID(cachefile string) Identifier {
	return machineIDGenerate{
		cachefile: cachefile,
	}
}

type machineIDGenerate struct {
	cachefile string
}

func (g machineIDGenerate) MachineID(rebuild bool) string {
	if f := g.cachefile; f != "" && !rebuild {
		if raw, _ := os.ReadFile(f); len(raw) != 0 {
			return string(raw)
		}
	}

	return g.generateAndSaveCache()
}

func (g machineIDGenerate) generateAndSaveCache() string {
	mid := g.generate()
	if f := g.cachefile; f != "" {
		_ = os.WriteFile(f, []byte(mid), 0o644)
	}

	return mid
}

func (g machineIDGenerate) generate() string {
	// sha1(hostid-hostname-macs)

	mid, _ := machineID()
	hostname, _ := os.Hostname()
	card := g.networks()
	str := strings.Join([]string{mid, hostname, card}, ",")
	sum := sha1.Sum([]byte(str))

	return hex.EncodeToString(sum[:])
}

func (g machineIDGenerate) networks() string {
	virtuals := virtualNetworks()
	if virtuals == nil {
		virtuals = make(map[string]bool)
	}
	faces, _ := net.Interfaces()
	cards := make(nics, 0, len(faces))
	for _, face := range faces {
		// 跳过换回网卡和未启用的网卡
		if face.Flags&net.FlagUp == 0 ||
			face.Flags&net.FlagLoopback != 0 {
			continue
		}

		// 虚拟网卡经常变化，不纳入计算因子。
		if virtuals[face.Name] {
			continue
		}

		// 一些虚拟网卡是没有 MAC 地址的
		if len(face.HardwareAddr) == 0 {
			continue
		}

		zeroMAC := true
		for _, b := range face.HardwareAddr {
			if b != 0 {
				zeroMAC = false
				break
			}
		}
		if zeroMAC {
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

			// 仅统计 IPv4 不统计 IPv6。
			// 因为现实使用中往往 IPv6 地址变化性更强。
			if ip4 := ip.To4(); ip4 != nil {
				ips = append(ips, ip.String())
			}
		}
		if len(ips) != 0 {
			cards = append(cards, &nic{
				MAC:   face.HardwareAddr.String(),
				Inets: ips,
			})
		}
	}
	cards.sort()

	return cards.join()
}

type nic struct {
	MAC   string
	Inets []string
}

type nics []*nic

func (ns nics) sort() {
	slices.SortFunc(ns, func(a, b *nic) int {
		return strings.Compare(a.MAC, b.MAC)
	})
	for _, n := range ns {
		slices.Sort(n.Inets)
	}
}

func (ns nics) join() string {
	strs := make([]string, 0, len(ns))
	for _, n := range ns {
		ele := make([]string, 0, len(n.Inets)+1)
		ele = append(ele, n.MAC)
		ele = append(ele, n.Inets...)
		line := strings.Join(ele, ",")
		strs = append(strs, line)
	}

	return strings.Join(strs, ",")
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
