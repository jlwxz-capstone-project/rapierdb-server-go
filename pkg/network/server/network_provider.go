package network_server

type NetworkProvider interface {
	Start() error
	Stop() error
	CloseConnection(clientId string) error
	CloseAllConnections() error
	Send(clientId string, msg []byte) error
	Broadcast(msg []byte) error
	SetMsgHandler(handler func(clientId string, msg []byte))
	GetAllClientIds() []string
	GetStatus() NetworkStatus
	SubscribeStatusChange() <-chan NetworkStatus
	UnsubscribeStatusChange(ch <-chan NetworkStatus)
	WaitForStatus(targetStatus NetworkStatus) <-chan struct{}
}
