package tun

import "net/netip"

type TunLink struct {
	addrs []netip.Prefix
}

func NewTunLink(tunCfg TunConfig) (*TunLink, error) {
	return &TunLink{}, nil
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
