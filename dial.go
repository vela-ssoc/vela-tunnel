package tunnel

import (
	"context"
	"crypto/tls"
	"net"
	"net/url"
	"time"
)

type dialer interface {
	iterDial(context.Context, time.Duration) (net.Conn, *Address, error)
	lookupMAC(net.IP) net.HardwareAddr
}

func newDialer(addrs []string, servername string) dialer {
	dl := &iterDial{
		dial: &tls.Dialer{NetDialer: new(net.Dialer)},
		macs: make(map[string]net.HardwareAddr, 4),
	}

	ads := dl.toAddrs(addrs, servername)
	dl.addrs = ads
	dl.length = len(ads)

	return dl
}

type iterDial struct {
	dial   *tls.Dialer
	macs   map[string]net.HardwareAddr
	addrs  Addresses
	length int
	index  int
}

func (dl *iterDial) iterDial(parent context.Context, timeout time.Duration) (net.Conn, *Address, error) {
	idx := dl.index
	addr := dl.addrs[idx]
	dl.index = (idx + 1) % dl.length

	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	if addr.TLS {
		dial := dl.dial
		dial.Config = &tls.Config{ServerName: addr.Name}
		conn, err := dial.DialContext(ctx, "tcp", addr.Addr)
		return conn, addr, err
	} else {
		conn, err := dl.dial.NetDialer.DialContext(ctx, "tcp", addr.Addr)
		return conn, addr, err
	}
}

func (dl *iterDial) lookupMAC(ip net.IP) net.HardwareAddr {
	sip := ip.String()
	if hw, ok := dl.macs[sip]; ok {
		return hw
	}

	var mac net.HardwareAddr
	ifs, _ := net.Interfaces()
	for _, face := range ifs {
		var match bool
		addrs, _ := face.Addrs()
		for _, addr := range addrs {
			inet, ok := addr.(*net.IPNet)
			if match = ok && inet.IP.Equal(ip); match {
				break
			}
		}
		if match {
			mac = face.HardwareAddr
			break
		}
	}

	dl.macs[sip] = mac

	return mac
}

func (dl *iterDial) toAddrs(addrs []string, servername string) Addresses {
	size := len(addrs)
	ret := make(Addresses, 0, size)
	tcps := make(map[string]struct{}, size)
	ssls := make(map[string]struct{}, size)

	for _, addr := range addrs {
		host, port := dl.splitHP(addr)
		sport, tport := port, port
		if port == "" {
			sport, tport = "443", "80"
		}
		shost := net.JoinHostPort(host, sport)
		thost := net.JoinHostPort(host, tport)

		if _, ok := ssls[shost]; !ok {
			ssls[shost] = struct{}{}
			ret = append(ret, &Address{TLS: true, Addr: shost, Name: servername})
		}
		if _, ok := tcps[thost]; !ok {
			tcps[thost] = struct{}{}
			ret = append(ret, &Address{Addr: thost, Name: servername})
		}
	}

	return ret
}

// splitHP 分割出主机和端口号
func (*iterDial) splitHP(hp string) (string, string) {
	if u, _ := url.Parse(hp); u != nil {
		hp = u.Host
	}

	if host, port, err := net.SplitHostPort(hp); err == nil {
		return host, port
	}

	return hp, ""
}
