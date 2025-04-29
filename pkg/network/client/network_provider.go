package network_client

type NetworkProvider interface {
	Connect() error
	Close() error
	Send(msg []byte) error
	SetMsgHandler(handler func(msg []byte))
	GetStatus() NetworkStatus
	SubscribeStatusChange() <-chan NetworkStatus
	UnsubscribeStatusChange(ch <-chan NetworkStatus)
}
