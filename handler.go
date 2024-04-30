package netstackgo

import (
	"net"
	"net/netip"

	"github.com/josexy/netstackgo/tun/core/adapter"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type ConnHandler interface {
	HandleTCPConn(ConnTuple, net.Conn)
	HandleUDPConn(ConnTuple, net.PacketConn)
}

type ConnTuple struct {
	SrcAddr netip.AddrPort
	DstAddr netip.AddrPort
}

func newConnTuple(id *stack.TransportEndpointID) ConnTuple {
	srcIP, _ := netip.AddrFromSlice(id.RemoteAddress.AsSlice())
	dstIP, _ := netip.AddrFromSlice(id.LocalAddress.AsSlice())
	return ConnTuple{
		SrcAddr: netip.AddrPortFrom(srcIP, id.RemotePort),
		DstAddr: netip.AddrPortFrom(dstIP, id.LocalPort),
	}
}

func (t *ConnTuple) Src() string {
	return t.SrcAddr.String()
}

func (t *ConnTuple) Dst() string {
	return t.DstAddr.String()
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
