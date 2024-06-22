package tun

import (
	"net/netip"
)

var DefaultMTU uint32 = 1350

type TunConfig struct {
	Name string
	CIDR []netip.Prefix
	MTU  uint32
}
