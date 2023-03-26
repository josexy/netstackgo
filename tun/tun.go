package tun

import (
	"fmt"
	"net/netip"
	"os/exec"
	"strings"
)

var DefaultMTU uint32 = 1350

type IPRoute struct {
	Dest    netip.Prefix
	Gateway netip.Addr
}

type TunConfig struct {
	Name string
	Addr string
	MTU  uint32
}

func exeCmd(cmd string) error {
	args := strings.Split(cmd, " ")
	if err := exec.Command(args[0], args[1:]...).Run(); err != nil {
		return fmt.Errorf("%q: %v", cmd, err)
	}
	return nil
}
