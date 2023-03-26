# netstackgo
An easy-to-use netstack, wrapped by gvisor and wireguard-go, which supports macOS/Linux/Windows.

```shell
go get -u github.com/josexy/netstackgo
```

# usage

```go
type myHandler struct{}

func (*myHandler) HandleTCPConn(info *netstackgo.ConnTuple, conn net.Conn) {
	log.Printf("tcp, src: %s, dst: %s", info.Src(), info.Dst())
	// do something...
}
func (*myHandler) HandleUDPConn(info *netstackgo.ConnTuple, conn net.PacketConn) {
	log.Printf("udp, src: %s, dst: %s", info.Src(), info.Dst())
	// do something...
}

func main() {
	nt := netstackgo.New(tun.TunConfig{
		Name: "tun2",
		Addr: "192.18.0.1/16",
		MTU:  tun.DefaultMTU,
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
```

PS: Windows user requires downloading wintun.dll from https://www.wintun.net

# credits

- [clash](https://github.com/Dreamacro/clash)
- [tun2socks](https://github.com/xjasonlyu/tun2socks)
- [gvisor](https://github.com/google/gvisor)
- [wireguard-go](https://git.zx2c4.com/wireguard-go)
