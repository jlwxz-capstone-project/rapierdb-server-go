package network_client

import "fmt"

type NetworkStatus int32

const (
	NetworkNotReady NetworkStatus = 0
	NetworkReady    NetworkStatus = 1
	NetworkClosed   NetworkStatus = 2
)

func (s NetworkStatus) String() string {
	switch s {
	case NetworkNotReady:
		return "not ready"
	case NetworkReady:
		return "ready"
	case NetworkClosed:
		return "closed"
	default:
		return fmt.Sprintf("unknown status: %d", s)
	}
}
