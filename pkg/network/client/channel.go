package network_client

// ClientChannel 定义了一个客户端通用的通信通道接口
type ClientChannel interface {
	// Setup 初始化通道
	Setup() error

	// Close 关闭通道
	Close() error

	// Send 发送消息
	Send(msg []byte) error

	// SetMsgHandler 设置消息处理函数
	SetMsgHandler(handler func(msg []byte))
}
