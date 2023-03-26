package tun

import (
	"fmt"
	"net"
)

func SetTunAddress(name string, addr string, mtu uint32) error {
	ip, _, _ := net.ParseCIDR(addr)
	if mtu <= 0 {
		mtu = DefaultMTU
	}
	cmd := fmt.Sprintf("ifconfig %s inet %s %s mtu %d up", name, addr, ip.String(), mtu)
	if err := exeCmd(cmd); err != nil {
		return err
	}
	return nil
}

func AddTunRoutes(name string, routes []IPRoute) error {
	for _, route := range routes {
		if !route.Dest.IsValid() && !route.Gateway.IsValid() {
			continue
		}
		cmd := fmt.Sprintf("route add -net %s %s", route.Dest.String(), route.Gateway)
		if err := exeCmd(cmd); err != nil {
			return err
		}
	}
	return nil
}
