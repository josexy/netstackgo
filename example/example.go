package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net"
	"net/netip"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/josexy/netstackgo"
	"github.com/josexy/netstackgo/bind"
	"github.com/josexy/netstackgo/iface"
	"github.com/josexy/netstackgo/tun"
)

var (
	tunName    string = "utun5"
	tunCIDR    string = "192.168.0.10/16"
	remoteAddr string = ""
)

func tunnel(dst, src io.ReadWriteCloser) {
	ch := make(chan error, 2)
	defer dst.Close()
	defer src.Close()
	fn := func(dest, src io.ReadWriteCloser) {
		_, err := io.Copy(dest, src)
		ch <- err
	}
	go fn(dst, src)
	go fn(src, dst)
	<-ch
}

type handler struct{}

func (h *handler) HandleTCPConn(info netstackgo.ConnTuple, conn net.Conn) {
	log.Printf("tcp, src: %s, dst: %s", info.Src(), info.Dst())
	dialer := net.Dialer{Timeout: time.Second * 10}
	name, err := iface.DefaultRouteInterface()
	if err != nil {
		log.Println(err)
		return
	}
	// bind an outbound interface to avoid routing loops
	if err := bind.BindToDeviceForConn(name, &dialer, "tcp4", info.DstAddr.Addr()); err != nil {
		log.Println(err)
		return
	}
	target, err := dialer.DialContext(context.Background(), "tcp", remoteAddr)
	if err != nil {
		log.Println(err)
		return
	}
	defer target.Close()
	tunnel(target, conn)
}

func (h *handler) HandleUDPConn(info netstackgo.ConnTuple, conn net.PacketConn) {
	log.Printf("udp, src: %s, dst: %s", info.Src(), info.Dst())
	// do something...
}

func main() {
	flag.StringVar(&tunName, "name", tunName, "tun device name")
	flag.StringVar(&tunCIDR, "addr", tunCIDR, "tun device cidr address")
	flag.StringVar(&remoteAddr, "remote", remoteAddr, "test remote address")
	flag.Parse()

	log.Println(tunName, tunCIDR, remoteAddr)

	// creating a tun device requires root permissions
	nt := netstackgo.New(tun.TunConfig{
		Name: tunName,
		CIDR: []netip.Prefix{netip.MustParsePrefix(tunCIDR)},
		MTU:  tun.DefaultMTU,
	})
	if err := nt.Start(); err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := nt.Close(); err != nil {
			log.Println(err)
		}
	}()

	nt.RegisterConnHandler(&handler{})

	// wait
	inter := make(chan os.Signal, 1)
	signal.Notify(inter, syscall.SIGINT)
	<-inter
}
