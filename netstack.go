package netstackgo

import (
	"errors"
	"net/netip"
	"sync/atomic"

	"github.com/josexy/netstackgo/tun"
	"github.com/josexy/netstackgo/tun/core"
	T "github.com/josexy/netstackgo/tun/core/device/tun"
	"github.com/josexy/netstackgo/tun/core/option"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
)

var DefaultRoutes = []netip.Prefix{
	netip.PrefixFrom(netip.AddrFrom4([4]byte{1, 0, 0, 0}), 8),
	netip.PrefixFrom(netip.AddrFrom4([4]byte{2, 0, 0, 0}), 7),
	netip.PrefixFrom(netip.AddrFrom4([4]byte{4, 0, 0, 0}), 6),
	netip.PrefixFrom(netip.AddrFrom4([4]byte{8, 0, 0, 0}), 5),
	netip.PrefixFrom(netip.AddrFrom4([4]byte{16, 0, 0, 0}), 4),
	netip.PrefixFrom(netip.AddrFrom4([4]byte{32, 0, 0, 0}), 3),
	netip.PrefixFrom(netip.AddrFrom4([4]byte{64, 0, 0, 0}), 2),
	netip.PrefixFrom(netip.AddrFrom4([4]byte{128, 0, 0, 0}), 1),
}

type TunNetstack struct {
	netstack   *stack.Stack
	tunStackEp stack.LinkEndpoint
	tunCfg     tun.TunConfig
	tunLink    *tun.TunLink
	handler    *tunTransportHandler
	running    atomic.Bool
}

func New(tunCfg tun.TunConfig) *TunNetstack {
	return &TunNetstack{
		tunCfg:  tunCfg,
		handler: newTunTransportHandler(),
	}
}

func (ns *TunNetstack) Start() (err error) {
	if ns.running.Load() {
		return errors.New("tun netstack is running")
	}

	if ns.tunStackEp, err = T.Open(ns.tunCfg.Name, ns.tunCfg.MTU); err != nil {
		return
	}

	if ns.tunLink, err = tun.NewTunLink(ns.tunCfg); err != nil {
		return
	}

	if err = ns.tunLink.ConfigureTunAddrs(); err != nil {
		return
	}

	if err = ns.tunLink.ConfigureTunRoutes(); err != nil {
		ns.tunLink.DeleteConfiguration()
		return
	}

	ns.handler.run()

	if err = ns.createStack(); err != nil {
		ns.tunLink.DeleteConfiguration()
		return
	}
	ns.running.Store(true)
	return
}

func (ns *TunNetstack) Close() error {
	if !ns.running.Load() {
		return errors.New("tun netstack was stopped")
	}

	ns.tunLink.DeleteConfiguration()
	ns.tunStackEp.Close()
	ns.handler.finish()
	ns.netstack.Close()
	ns.netstack.Wait()
	return nil
}

func (ns *TunNetstack) RegisterConnHandler(handler ConnHandler) {
	ns.handler.registerConnHandler(handler)
}

func (ns *TunNetstack) createStack() error {
	ns.netstack = stack.New(stack.Options{
		NetworkProtocols: []stack.NetworkProtocolFactory{
			ipv4.NewProtocol,
			ipv6.NewProtocol,
		},
		TransportProtocols: []stack.TransportProtocolFactory{
			tcp.NewProtocol,
			udp.NewProtocol,
			icmp.NewProtocol4,
			icmp.NewProtocol6,
		},
	})

	nicID := tcpip.NICID(ns.netstack.UniqueID())

	opts := []option.Option{option.WithDefault()}

	opts = append(opts,
		// Important: We must initiate transport protocol handlers
		// before creating NIC, otherwise NIC would dispatch packets
		// to stack and cause race condition.
		// Initiate transport protocol (TCP/UDP) with given handler.
		core.WithTCPHandler(ns.handler.HandleTCP),
		core.WithUDPHandler(ns.handler.HandleUDP),

		// Create stack NIC and then bind link endpoint to it.
		core.WithCreatingNIC(nicID, ns.tunStackEp),

		// In the past we did s.AddAddressRange to assign 0.0.0.0/0
		// onto the interface. We need that to be able to terminate
		// all the incoming connections - to any ip. AddressRange API
		// has been removed and the suggested workaround is to use
		// Promiscuous mode. https://github.com/google/gvisor/issues/3876
		//
		// Ref: https://github.com/cloudflare/slirpnetstack/blob/master/stack.go
		core.WithPromiscuousMode(nicID, core.NicPromiscuousModeEnabled),

		// Enable spoofing if a stack may send packets from unowned
		// addresses. This change required changes to some netgophers
		// since previously, promiscuous mode was enough to let the
		// netstack respond to all incoming packets regardless of the
		// packet's destination address. Now that a stack.Route is not
		// held for each incoming packet, finding a route may fail with
		// local addresses we don't own but accepted packets for while
		// in promiscuous mode. Since we also want to be able to send
		// from any address (in response the received promiscuous mode
		// packets), we need to enable spoofing.
		//
		// Ref: https://github.com/google/gvisor/commit/8c0701462a84ff77e602f1626aec49479c308127
		core.WithSpoofing(nicID, core.NicSpoofingEnabled),

		// Add default route table for IPv4 and IPv6. This will handle
		// all incoming ICMP packets.
		core.WithRouteTable(nicID),
	)

	for _, opt := range opts {
		if err := opt(ns.netstack); err != nil {
			return err
		}
	}
	return nil
}
