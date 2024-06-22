package tun

import (
	"math/rand"
	"net"
	"net/netip"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

type TunLink struct {
	link                               netlink.Link
	addrv4, addrv6                     []netip.Prefix
	tableIndex                         int
	ruleStartIndexV4, ruleStartIndexV6 int
	ruleEndIndexV4, ruleEndIndexV6     int
}

func NewTunLink(tunCfg TunConfig) (*TunLink, error) {
	link, err := netlink.LinkByName(tunCfg.Name)
	if err != nil {
		return nil, err
	}

	tableIndex := tunCfg.IPRoute2TableIndex
	ruleIndex := tunCfg.IPRoute2RuleIndex
	if tableIndex <= 0 {
		tableIndex = 10086
	}
	if ruleIndex <= 0 {
		ruleIndex = 10086
	}
	for {
		routeList, err := netlink.RouteListFiltered(netlink.FAMILY_ALL, &netlink.Route{Table: tableIndex}, netlink.RT_FILTER_TABLE)
		if len(routeList) == 0 || err != nil {
			break
		}
		tableIndex = int(rand.Uint32())
	}

	var addrv4, addrv6 []netip.Prefix
	for _, cidr := range tunCfg.CIDR {
		if cidr.Addr().Is4() {
			addrv4 = append(addrv4, cidr)
		} else if cidr.Addr().Is6() {
			addrv6 = append(addrv6, cidr)
		}
	}
	return &TunLink{
		link:             link,
		addrv4:           addrv4,
		addrv6:           addrv6,
		tableIndex:       tableIndex,
		ruleStartIndexV4: ruleIndex,
		ruleStartIndexV6: ruleIndex,
		ruleEndIndexV4:   ruleIndex,
		ruleEndIndexV6:   ruleIndex,
	}, nil
}

func (tr *TunLink) ConfigureTunAddrs() error {
	for _, cidr := range tr.addrv4 {
		if err := tr.AddAddr(cidr); err != nil {
			return err
		}
	}
	for _, cidr := range tr.addrv6 {
		if err := tr.AddAddr(cidr); err != nil {
			return err
		}
	}
	return tr.Setup()
}

func (tr *TunLink) ConfigureTunRoutes() error {
	if err := tr.AddRoute(netip.PrefixFrom(netip.IPv4Unspecified(), 0)); err != nil {
		return err
	}
	if err := tr.AddRoute(netip.PrefixFrom(netip.IPv6Unspecified(), 0)); err != nil {
		return err
	}
	return tr.AddRules()
}

func (tr *TunLink) DeleteConfiguration() error {
	tr.DelRoute(netip.PrefixFrom(netip.IPv4Unspecified(), 0))
	tr.DelRoute(netip.PrefixFrom(netip.IPv6Unspecified(), 0))
	for _, cidr := range tr.addrv4 {
		tr.DelAddr(cidr)
	}
	for _, cidr := range tr.addrv6 {
		tr.DelAddr(cidr)
	}
	tr.Setdown()
	return tr.DelRules()
}

func (tr *TunLink) AddAddr(cidr netip.Prefix) error {
	addr, err := netlink.ParseAddr(cidr.String())
	if err != nil {
		return err
	}
	return netlink.AddrAdd(tr.link, addr)
}

func (tr *TunLink) DelAddr(cidr netip.Prefix) error {
	addr, err := netlink.ParseAddr(cidr.String())
	if err != nil {
		return err
	}
	return netlink.AddrDel(tr.link, addr)
}

// 9000:	from all to 192.18.0.0/16 lookup 2022
// 9001:	from all lookup 2022 suppress_prefixlength 0
// 9002:	not from all dport 53 lookup main suppress_prefixlength 0
// 9002:	from all iif tun0 goto 9010
// 9003:	not from all iif lo lookup 2022
// 9003:	from 0.0.0.0 iif lo lookup 2022
// 9003:	from 192.18.0.0/16 iif lo lookup 2022
// 9010:	from all nop
func (tr *TunLink) AddRules() error {
	addrules := func(ruleIndex int, cidrs []netip.Prefix, family int) int {
		var rule *netlink.Rule

		for _, cidr := range cidrs {
			rule = netlink.NewRule()
			rule.Priority = ruleIndex
			rule.Dst = netPrefix2IPNet(cidr.Masked())
			rule.Table = tr.tableIndex
			rule.Family = family
			netlink.RuleAdd(rule)
			ruleIndex++
		}

		rule = netlink.NewRule()
		rule.Priority = ruleIndex
		rule.Table = tr.tableIndex
		rule.SuppressPrefixlen = 0
		rule.Family = family
		netlink.RuleAdd(rule)
		ruleIndex++

		rule = netlink.NewRule()
		rule.Invert = true
		rule.Priority = ruleIndex
		rule.Dport = netlink.NewRulePortRange(53, 53)
		rule.Table = unix.RT_TABLE_MAIN
		rule.SuppressPrefixlen = 0
		rule.Family = family
		netlink.RuleAdd(rule)
		ruleIndex++

		if family == netlink.FAMILY_V4 {
			rule = netlink.NewRule()
			rule.Invert = true
			rule.Priority = ruleIndex
			rule.IifName = "lo"
			rule.Table = tr.tableIndex
			rule.Family = family
			netlink.RuleAdd(rule)

			rule = netlink.NewRule()
			rule.Priority = ruleIndex
			rule.Src = netPrefix2IPNet(netip.PrefixFrom(netip.IPv4Unspecified(), 32))
			rule.IifName = "lo"
			rule.Table = tr.tableIndex
			rule.Family = family
			netlink.RuleAdd(rule)
		}

		for _, cidr := range cidrs {
			rule = netlink.NewRule()
			rule.Priority = ruleIndex
			rule.Src = netPrefix2IPNet(cidr.Masked())
			rule.IifName = "lo"
			rule.Table = tr.tableIndex
			rule.Family = family
			netlink.RuleAdd(rule)
		}
		return ruleIndex
	}

	ruleIndexV4 := tr.ruleStartIndexV4
	ruleIndexV6 := tr.ruleStartIndexV6
	if len(tr.addrv4) > 0 {
		tr.ruleEndIndexV4 = addrules(ruleIndexV4, tr.addrv4, netlink.FAMILY_V4)
	}
	if len(tr.addrv6) > 0 {
		tr.ruleEndIndexV6 = addrules(ruleIndexV6, tr.addrv6, netlink.FAMILY_V6)
	}

	return nil
}

func (tr *TunLink) DelRules() error {
	rules, err := netlink.RuleList(netlink.FAMILY_ALL)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		if (rule.Family == netlink.FAMILY_V4 && rule.Priority >= tr.ruleStartIndexV4 && rule.Priority <= tr.ruleEndIndexV4) ||
			rule.Family == netlink.FAMILY_V6 && rule.Priority >= tr.ruleStartIndexV6 && rule.Priority <= tr.ruleEndIndexV6 {
			delRule := netlink.NewRule()
			delRule.Family = rule.Family
			delRule.Priority = rule.Priority
			netlink.RuleDel(delRule)
		}
	}
	return nil
}

func (tr *TunLink) AddRoute(cidr netip.Prefix) error {
	return netlink.RouteAdd(&netlink.Route{
		LinkIndex: tr.link.Attrs().Index,
		Table:     tr.tableIndex,
		Dst:       netPrefix2IPNet(cidr),
	})
}

func (tr *TunLink) DelRoute(cidr netip.Prefix) error {
	return netlink.RouteDel(&netlink.Route{
		LinkIndex: tr.link.Attrs().Index,
		Table:     tr.tableIndex,
		Dst:       netPrefix2IPNet(cidr),
	})
}

func (tr *TunLink) Setup() error { return netlink.LinkSetUp(tr.link) }

func (tr *TunLink) Setdown() error { return netlink.LinkSetDown(tr.link) }

func netPrefix2IPNet(cidr netip.Prefix) *net.IPNet {
	return &net.IPNet{
		IP:   cidr.Addr().AsSlice(),
		Mask: net.CIDRMask(cidr.Bits(), cidr.Addr().BitLen()),
	}
}
