package network_server

import "fmt"

type NetworkStatus int32

const (
	NetworkNotStarted NetworkStatus = 0
	NetworkStarting   NetworkStatus = 1
	NetworkRunning    NetworkStatus = 2
	NetworkStopping   NetworkStatus = 3
	NetworkStopped    NetworkStatus = 4
)

func (s NetworkStatus) String() string {
	switch s {
	case NetworkNotStarted:
		return "not started"
	case NetworkStarting:
		return "starting"
	case NetworkRunning:
		return "running"
	case NetworkStopping:
		return "stopping"
	case NetworkStopped:
		return "stopped"
	default:
		panic(fmt.Sprintf("Unknown server status: %d", s))
	}
}
