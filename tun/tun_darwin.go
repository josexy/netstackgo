package tun

import "net/netip"

type TunLink struct {
	addrs []netip.Prefix
}

func NewTunLink(name string, cidrs []netip.Prefix) (*TunLink, error) {
	return &TunLink{
		addrs: cidrs,
	}, nil
}

func (tr *TunLink) ConfigureTunAddrs() error {
	return nil
}

func (tr *TunLink) ConfigureTunRoutes() error {
	return nil
}

func (tr *TunLink) DeleteConfiguration() error {
	return nil
}
