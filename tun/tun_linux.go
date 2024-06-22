package tun

import (
	"math/rand"
	"net"
	"net/netip"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

type TunLink struct {
	link           netlink.Link
	addrs          []netip.Prefix
	tableIndex     int
	ruleStartIndex int
	ruleEndIndex   int
}

func NewTunLink(name string, cidrs []netip.Prefix) (*TunLink, error) {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return nil, err
	}

	var tableIndex int = 10086
	var ruleIndex int = 10086
	for {
		tableIndex = int(rand.Uint32())
		routeList, fErr := netlink.RouteListFiltered(netlink.FAMILY_ALL, &netlink.Route{Table: tableIndex}, netlink.RT_FILTER_TABLE)
		if len(routeList) == 0 || fErr != nil {
			break
		}
	}

	return &TunLink{
		link:           link,
		addrs:          cidrs,
		tableIndex:     tableIndex,
		ruleStartIndex: ruleIndex,
		ruleEndIndex:   ruleIndex,
	}, nil
}

func (tr *TunLink) ConfigureTunAddrs() error {
	for _, cidr := range tr.addrs {
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
	for _, cidr := range tr.addrs {
		if err := tr.DelAddr(cidr); err != nil {
			return err
		}
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

// ip rule add not from all dport 53 lookup main suppress_prefixlength 0
// ip rule add not from all iif lo lookup 1970566510
// ip rule add from 0.0.0.0 iif lo uidrange 0-4294967294 lookup 1970566510
// ip rule add from 198.18.0.1 iif lo uidrange 0-4294967294 lookup 1970566510
func (tr *TunLink) AddRules() error {
	ruleIndex := tr.ruleStartIndex
	rule := netlink.NewRule()
	rule.Invert = true
	rule.Priority = ruleIndex
	rule.Dport = netlink.NewRulePortRange(53, 53)
	rule.Table = unix.RT_TABLE_MAIN
	rule.SuppressPrefixlen = 0
	rule.Family = netlink.FAMILY_V4
	netlink.RuleAdd(rule)
	ruleIndex++

	rule = netlink.NewRule()
	rule.Invert = true
	rule.Priority = ruleIndex
	rule.Dport = netlink.NewRulePortRange(53, 53)
	rule.Table = unix.RT_TABLE_MAIN
	rule.SuppressPrefixlen = 0
	rule.Family = netlink.FAMILY_V6
	netlink.RuleAdd(rule)
	ruleIndex++

	rule = netlink.NewRule()
	rule.Invert = true
	rule.Priority = ruleIndex
	rule.IifName = "lo"
	rule.Table = tr.tableIndex
	rule.Family = netlink.FAMILY_V4
	netlink.RuleAdd(rule)
	ruleIndex++

	const endUID = 0xFFFFFFFF - 1
	rule = netlink.NewRule()
	rule.Priority = ruleIndex
	rule.Src = netPrefix2IPNet(netip.PrefixFrom(netip.IPv4Unspecified(), 32))
	rule.IifName = "lo"
	rule.Table = tr.tableIndex
	rule.Family = netlink.FAMILY_V4
	rule.UIDRange = netlink.NewRuleUIDRange(0, endUID)
	netlink.RuleAdd(rule)
	ruleIndex++

	for _, addr := range tr.addrs {
		rule = netlink.NewRule()
		rule.Priority = ruleIndex
		rule.Src = netPrefix2IPNet(netip.PrefixFrom(addr.Addr(), 32))
		rule.IifName = "lo"
		rule.Table = tr.tableIndex
		rule.UIDRange = netlink.NewRuleUIDRange(0, endUID)
		if addr.Addr().Is6() {
			rule.Family = netlink.FAMILY_V6
		} else {
			rule.Family = netlink.FAMILY_V4
		}
		netlink.RuleAdd(rule)
		ruleIndex++
	}
	tr.ruleEndIndex = ruleIndex - 1
	return nil
}

func (tr *TunLink) DelRules() error {
	rules, err := netlink.RuleList(netlink.FAMILY_ALL)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		if rule.Priority >= tr.ruleStartIndex && rule.Priority <= tr.ruleEndIndex {
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
