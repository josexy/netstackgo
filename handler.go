package netstackgo

import (
	"net"
	"net/netip"
	"strconv"

	"github.com/josexy/netstackgo/tun/core/adapter"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type ConnHandler interface {
	HandleTCPConn(*ConnTuple, net.Conn)
	HandleUDPConn(*ConnTuple, net.PacketConn)
}

type ConnTuple struct {
	SrcIP   netip.Addr
	SrcPort uint16
	DstIP   netip.Addr
	DstPort uint16
}

func newConnTuple(id *stack.TransportEndpointID) *ConnTuple {
	srcIP, _ := netip.AddrFromSlice([]byte(id.RemoteAddress))
	dstIP, _ := netip.AddrFromSlice([]byte(id.LocalAddress))

	return &ConnTuple{
		SrcIP:   srcIP,
		SrcPort: id.RemotePort,
		DstIP:   dstIP,
		DstPort: id.LocalPort,
	}
}

func (t *ConnTuple) Src() string {
	return net.JoinHostPort(t.SrcIP.String(), strconv.FormatUint(uint64(t.SrcPort), 10))
}

func (t *ConnTuple) Dst() string {
	return net.JoinHostPort(t.DstIP.String(), strconv.FormatUint(uint64(t.DstPort), 10))
}

func (t *ConnTuple) DstAddrPort() netip.AddrPort {
	return netip.AddrPortFrom(t.DstIP, uint16(t.DstPort))
}

type tunTransportHandler struct {
	tcpQueue chan adapter.TCPConn
	udpQueue chan adapter.UDPConn
	closeCh  chan struct{}
	adapter.TransportHandler
	connHandler ConnHandler
}

func newTunTransportHandler() *tunTransportHandler {
	handler := &tunTransportHandler{
		tcpQueue: make(chan adapter.TCPConn, 128),
		udpQueue: make(chan adapter.UDPConn, 128),
		closeCh:  make(chan struct{}, 1),
	}
	handler.TransportHandler = handler
	return handler
}

func (h *tunTransportHandler) registerConnHandler(handler ConnHandler) {
	h.connHandler = handler
}

func (h *tunTransportHandler) run() {
	go func() {
		defer func() { recover() }()
		for {
			select {
			case conn := <-h.tcpQueue:
				go h.handleTCPConn(conn)
			case conn := <-h.udpQueue:
				go h.handleUDPConn(conn)
			case <-h.closeCh:
				return
			}
		}
	}()
}

func (h *tunTransportHandler) finish() {
	h.closeCh <- struct{}{}
}

func (h *tunTransportHandler) HandleTCP(conn adapter.TCPConn) { h.tcpQueue <- conn }

func (h *tunTransportHandler) HandleUDP(conn adapter.UDPConn) { h.udpQueue <- conn }

func (h *tunTransportHandler) handleTCPConn(conn adapter.TCPConn) {
	defer conn.Close()
	connTuple := newConnTuple(conn.ID())
	if h.connHandler != nil {
		h.connHandler.HandleTCPConn(connTuple, conn)
	}
}

func (h *tunTransportHandler) handleUDPConn(conn adapter.UDPConn) {
	defer conn.Close()

	connTuple := newConnTuple(conn.ID())
	if h.connHandler != nil {
		h.connHandler.HandleUDPConn(connTuple, conn)
	}
}
