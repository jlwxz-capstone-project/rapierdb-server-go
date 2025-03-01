package network

// Channel 定义了一个通用的通信通道接口
// CONN 是一个泛型参数,表示连接的具体类型
type Channel interface {
	// Setup 初始化通道
	Setup() error

	// Close 关闭指定客户端的连接
	// clientId: 要关闭的客户端ID
	Close(clientId string) error

	// CloseAll 关闭所有客户端连接
	CloseAll() error

	// Send 向指定客户端发送消息
	// clientId: 目标客户端ID
	// msg: 要发送的消息内容
	Send(clientId string, msg []byte) error

	// Broadcast 向所有已连接的客户端广播消息
	// msg: 要广播的消息内容
	Broadcast(msg []byte) error

	// SetMsgHandler 设置消息处理函数
	// handler: 用于处理接收到的消息的回调函数
	SetMsgHandler(handler func(clientId string, msg []byte))

	// GetAllConnectedClientIds 获取所有已连接的客户端ID
	GetAllConnectedClientIds() []string
}
