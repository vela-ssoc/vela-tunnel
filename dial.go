package tunnel

import (
	"context"
	"crypto/tls"
	"net"
	"time"
)

type dialer interface {
	iterDial(context.Context, time.Duration) (net.Conn, *Address, error)
	lookupMAC(net.IP) net.HardwareAddr
}

func newDialer(ethernet, internet Addresses) dialer {
	esz, isz := len(ethernet), len(internet)
	length := esz + isz
	addresses := make(Addresses, 0, length)
	for _, a := range ethernet {
		a.eth = true
		addresses = append(addresses, a)
	}
	addresses = append(addresses, internet...)

	return &iterDial{
		dial:   &tls.Dialer{NetDialer: new(net.Dialer)},
		macs:   make(map[string]net.HardwareAddr, 4),
		addrs:  addresses,
		length: length,
	}
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
