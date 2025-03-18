package network_client

type ChannelStatus int

const (
	// 通道已准备好发送和接收消息
	CHANNEL_STATUS_READY ChannelStatus = 0
	// 通道未准备好发送和接收消息
	CHANNEL_STATUS_NOT_READY ChannelStatus = 1
)

// ClientChannel 定义了一个客户端通用的通信通道接口
type ClientChannel interface {
	// Send 发送消息
	Send(msg []byte) error

	// SetMsgHandler 设置消息处理函数
	SetMsgHandler(handler func(msg []byte))

	// GetStatus 获取通道状态
	GetStatus() ChannelStatus

	// SubscribeStatusChange 订阅通道状态变更事件
	SubscribeStatusChange() <-chan ChannelStatus

	// UnsubscribeStatusChange 取消订阅通道状态变更事件
	UnsubscribeStatusChange(ch <-chan ChannelStatus)
}

// ClientChannelManager 定义了一个客户端通用的通信通道管理器接口
type ClientChannelManager interface {
	// GetChannel 获取一个通信通道
	GetChannel() (ClientChannel, error)

	// Close 关闭一个通信通道
	Close(channel ClientChannel) error
}
