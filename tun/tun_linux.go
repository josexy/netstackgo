package tun

import "fmt"

func SetTunAddress(name string, addr string, mtu uint32) (err error) {
	if mtu <= 0 {
		mtu = DefaultMTU
	}
	cmds := []string{
		fmt.Sprintf("ip link set dev %s mtu %d", name, mtu),
		fmt.Sprintf("ip address add %s dev %s", addr, name),
		fmt.Sprintf("ip link set dev %s up", name),
	}
	for _, cmd := range cmds {
		if err := exeCmd(cmd); err != nil {
			return err
		}
	}
	return nil
}

func AddTunRoutes(name string, routes []IPRoute) error {
	for _, route := range routes {
		if !route.Dest.IsValid() && !route.Gateway.IsValid() {
			continue
		}
		cmd := fmt.Sprintf("ip route add %s via %s dev %s", route.Dest.String(), route.Gateway.String(), name)
		if err := exeCmd(cmd); err != nil {
			return err
		}
	}
	return nil
}
