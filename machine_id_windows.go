//go:build windows

package tunnel

import (
	"os"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// machineID returns the key MachineGuid in registry `HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography`.
// If there is an error running the commad an empty string is returned.
func machineID() (string, error) {
	pth := `SOFTWARE\Microsoft\Cryptography`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, pth, registry.QUERY_VALUE|registry.WOW64_64KEY)
	if err != nil {
		return "", err
	}
	defer k.Close()

	s, _, err := k.GetStringValue("MachineGuid")
	if err != nil {
		return "", err
	}

	return s, nil
}

func virtualNetworks() map[string]bool {
	ignores := []string{
		"VMware Virtual Ethernet Adapter",       // VMware
		"TAP-Windows Adapter",                   // OpenVPN/WireGuard
		"VirtualBox Host-Only Ethernet Adapter", // VirtualBox
		//"Hyper-V Virtual Ethernet Adapter",      // Hyper-V
		// FIXME 在安装 Hyper-V 虚拟机或启用了 WSL 环境时，会出现 Hyper-V 类型的虚拟网卡，
		//  我自测 Windows Server 安装了 Hyper-V 虚拟机，会创建一个 vSwitch，然后其他本机
		//  网卡会虚拟出来接入到 vSwitch 中。如果把 Hyper-V 虚拟网卡排除，就会导致无有效的网
		//  卡了，所以暂且不排除 Hyper-V 虚拟网卡。
	}

	addrs, _ := adapterAddresses()
	hm := make(map[string]bool, len(addrs))
	for _, addr := range addrs {
		friendlyName := windows.UTF16PtrToString(addr.FriendlyName)
		description := windows.UTF16PtrToString(addr.Description)
		var ignore bool
		for _, s := range ignores {
			if strings.HasPrefix(description, s) {
				ignore = true
				break
			}
		}
		hm[friendlyName] = ignore
	}

	return hm
}

// adapterAddresses returns a list of IP adapter and address
// structures. The structure contains an IP adapter and flattened
// multiple IP addresses including unicast, anycast and multicast
// addresses.
//
// 参考：https://github.com/golang/go/blob/go1.25.4/src/net/interface_windows.go#L14-L43
func adapterAddresses() ([]*windows.IpAdapterAddresses, error) {
	var b []byte
	l := uint32(15000) // recommended initial size
	for {
		b = make([]byte, l)
		const flags = windows.GAA_FLAG_INCLUDE_PREFIX | windows.GAA_FLAG_INCLUDE_GATEWAYS
		err := windows.GetAdaptersAddresses(syscall.AF_UNSPEC, flags, 0, (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])), &l)
		if err == nil {
			if l == 0 {
				return nil, nil
			}
			break
		}
		if err.(syscall.Errno) != syscall.ERROR_BUFFER_OVERFLOW {
			return nil, os.NewSyscallError("getadaptersaddresses", err)
		}
		if l <= uint32(len(b)) {
			return nil, os.NewSyscallError("getadaptersaddresses", err)
		}
	}
	var aas []*windows.IpAdapterAddresses
	for aa := (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])); aa != nil; aa = aa.Next {
		aas = append(aas, aa)
	}
	return aas, nil
}
