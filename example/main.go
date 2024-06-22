package main

import (
	"log"
	"net"
	"net/netip"
	"os"
	"os/signal"
	"syscall"

	"github.com/josexy/netstackgo"
	"github.com/josexy/netstackgo/tun"
)

type myHandler struct{}

func (*myHandler) HandleTCPConn(info netstackgo.ConnTuple, conn net.Conn) {
	log.Printf("tcp, src: %s, dst: %s", info.Src(), info.Dst())
	// do something...
}
func (*myHandler) HandleUDPConn(info netstackgo.ConnTuple, conn net.PacketConn) {
	log.Printf("udp, src: %s, dst: %s", info.Src(), info.Dst())
	// do something...
}

func simple() {
	nt := netstackgo.New(tun.TunConfig{
		Name: "utun9",
		CIDR: []netip.Prefix{
			netip.MustParsePrefix("198.18.0.1/16"),
		},
		MTU: tun.DefaultMTU,
	})
	if err := nt.Start(); err != nil {
		log.Fatal(err)
	}
	defer nt.Close()
	nt.RegisterConnHandler(&myHandler{})
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT)
	<-interrupt
}
