package tun

import (
	"fmt"
	"net"
)

func SetTunAddress(name string, addr string, mtu uint32) (err error) {
	ip, subnet, _ := net.ParseCIDR(addr)
	cmd := fmt.Sprintf("netsh interface ip set address \"%s\" static %s %s none", name, ip.String(), net.IP(subnet.Mask).String())
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
		cmd := fmt.Sprintf("netsh interface ipv4 add route %s \"%s\" %s metric=%d store=active",
			route.Dest.String(),
			name,
			route.Gateway.String(),
			10, // priority
		)
		if err := exeCmd(cmd); err != nil {
			return err
		}
	}
	return nil
}
