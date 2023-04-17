package bind

import (
	"encoding/binary"
	"net"
	"net/netip"
	"syscall"
	"unsafe"

	"github.com/josexy/netstackgo/iface"
	"golang.org/x/sys/windows"
)

const (
	IP_UNICAST_IF   = 31
	IPV6_UNICAST_IF = 31
)

type controlFn func(network, address string, c syscall.RawConn) error

func setupControl(ifaceIndex int, nextChain controlFn) controlFn {
	return func(network, address string, c syscall.RawConn) (err error) {
		defer func() {
			if err == nil && nextChain != nil {
				err = nextChain(network, address, c)
			}
		}()

		ipStr, _, err := net.SplitHostPort(address)
		if err == nil {
			if ip, err := netip.ParseAddr(ipStr); err == nil && !ip.IsGlobalUnicast() {
				return err
			}
		}

		var innerErr error
		err = c.Control(func(fd uintptr) {
			switch network {
			case "tcp4", "udp4":
				innerErr = setSocketopt4(windows.Handle(fd), uint32(ifaceIndex))
			case "tcp6", "udp6":
				innerErr = setSocketopt6(windows.Handle(fd), uint32(ifaceIndex))
			}
		})
		if innerErr != nil {
			err = innerErr
		}
		return
	}
}

func setSocketopt4(handle windows.Handle, index uint32) error {
	// For IPv4, this parameter must be an interface index in network byte order.
	// Ref: https://learn.microsoft.com/en-us/windows/win32/winsock/ipproto-ip-socket-options
	var bytes [4]byte
	binary.BigEndian.PutUint32(bytes[:], index)
	index = *(*uint32)(unsafe.Pointer(&bytes[0]))
	return windows.SetsockoptInt(handle, windows.IPPROTO_IP, IP_UNICAST_IF, int(index))
}

func setSocketopt6(handle windows.Handle, index uint32) error {
	return windows.SetsockoptInt(handle, windows.IPPROTO_IPV6, IPV6_UNICAST_IF, int(index))
}

func bindToDeviceForTCP(ifaceName string, dialer *net.Dialer, _ string, _ netip.Addr) error {
	iface, err := iface.GetInterfaceByName(ifaceName)
	if err != nil {
		return err
	}
	dialer.Control = setupControl(iface.Index, dialer.Control)
	return nil
}

func bindToDeviceForUDP(ifaceName string, lc *net.ListenConfig, _, address string) (string, error) {
	iface, err := iface.GetInterfaceByName(ifaceName)
	if err != nil {
		return "", err
	}
	lc.Control = setupControl(iface.Index, lc.Control)
	return address, nil
}
