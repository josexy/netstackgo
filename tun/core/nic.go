package core

import (
	"fmt"

	"github.com/josexy/netstackgo/tun/core/option"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

const (
	// nicPromiscuousModeEnabled is the value used by stack to enable
	// or disable NIC's promiscuous mode.
	NicPromiscuousModeEnabled = true

	// nicSpoofingEnabled is the value used by stack to enable or disable
	// NIC's spoofing.
	NicSpoofingEnabled = true
)

// withCreatingNIC creates NIC for stack.
func WithCreatingNIC(nicID tcpip.NICID, ep stack.LinkEndpoint) option.Option {
	return func(s *stack.Stack) error {
		if err := s.CreateNICWithOptions(nicID, ep,
			stack.NICOptions{
				Disabled: false,
				// If no queueing discipline was specified
				// provide a stub implementation that just
				// delegates to the lower link endpoint.
				QDisc: nil,
			}); err != nil {
			return fmt.Errorf("create NIC: %s", err)
		}
		return nil
	}
}

// withPromiscuousMode sets promiscuous mode in the given NICs.
func WithPromiscuousMode(nicID tcpip.NICID, v bool) option.Option {
	return func(s *stack.Stack) error {
		if err := s.SetPromiscuousMode(nicID, v); err != nil {
			return fmt.Errorf("set promiscuous mode: %s", err)
		}
		return nil
	}
}

// withSpoofing sets address spoofing in the given NICs, allowing
// endpoints to bind to any address in the NIC.
func WithSpoofing(nicID tcpip.NICID, v bool) option.Option {
	return func(s *stack.Stack) error {
		if err := s.SetSpoofing(nicID, v); err != nil {
			return fmt.Errorf("set spoofing: %s", err)
		}
		return nil
	}
}
