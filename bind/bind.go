package bind

import (
	"net"
	"net/netip"
)

// BindToDeviceForTCP bind tcp socket to outbound interface
func BindToDeviceForTCP(ifaceName string, dialer *net.Dialer, network string, dst netip.Addr) error {
	return bindToDeviceForTCP(ifaceName, dialer, network, dst)
}

// BindToDeviceForUDP bind udp socket to local listening address
func BindToDeviceForUDP(ifaceName string, lc *net.ListenConfig, network, dst string) (string, error) {
	return bindToDeviceForUDP(ifaceName, lc, network, dst)
}
