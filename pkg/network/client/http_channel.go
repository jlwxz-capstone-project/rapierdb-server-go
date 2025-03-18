package network_client

import (
	"context"
	"sync"
	"time"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/sse"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// HttpChannel 基于 HTTP 和 SSE 实现的客户端通信通道
// 它实现了 ClientChannel 接口
// HttpChannel 的生命周期完全由 HttpChannelManager 管理！
type HttpChannel struct {
	sendFunc   func(msg []byte) error // 发送消息的函数
	closeFunc  func() error           // 关闭通道的函数
	sseClient  *sse.SseClient         // SSE 客户端，用于接收服务器推送的消息
	ctx        context.Context        // 通道上下文
	cancel     context.CancelFunc     // 取消函数，用于停止通道
	msgHandler func(msg []byte)       // 消息处理函数，处理接收到的消息
	mutex      sync.RWMutex           // 保护消息处理函数的互斥锁

	// 状态相关字段
	status     ChannelStatus
	statusLock sync.RWMutex
	statusEb   *util.EventBus[ChannelStatus]
}

// 确保 HttpChannel 实现了 ClientChannel 接口
var _ ClientChannel = &HttpChannel{}

// Send 发送消息到服务器
// 参数 msg 是要发送的消息字节数组
// 返回发送过程中可能出现的错误
func (c *HttpChannel) Send(msg []byte) error {
	return c.sendFunc(msg)
}

// SetMsgHandler 设置消息处理函数
// 参数 handler 是用于处理接收到消息的回调函数
func (c *HttpChannel) SetMsgHandler(handler func(msg []byte)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.msgHandler = handler
}

// HandleSseEvent 处理 SSE 事件
// 参数 event 是接收到的 SSE 事件
func (c *HttpChannel) HandleSseEvent(event *sse.SseEvent) {
	c.mutex.RLock()
	handler := c.msgHandler
	c.mutex.RUnlock()

	if handler != nil && event.Data != nil {
		handler(event.Data)
	}
}

// GetStatus 获取通道当前状态
func (c *HttpChannel) GetStatus() ChannelStatus {
	c.statusLock.RLock()
	defer c.statusLock.RUnlock()
	return c.status
}

// setStatus 设置通道状态
func (c *HttpChannel) setStatus(status ChannelStatus) {
	c.statusLock.Lock()
	defer c.statusLock.Unlock()
	oldStatus := c.status
	c.status = status

	if oldStatus != status {
		c.statusEb.Publish(status)
	}
}

// BindSseClient 绑定 SSE 客户端
// 会监听 SSE 客户端的状态变更，并更新通道状态
func (c *HttpChannel) BindSseClient(sseClient *sse.SseClient) {
	c.sseClient = sseClient
	sseStatusCh := sseClient.SubscribeStatusChange()

	go func() {
		defer sseClient.UnsubscribeStatusChange(sseStatusCh)

		for {
			select {
			case <-c.ctx.Done():
				return
			case sseStatus := <-sseStatusCh:
				switch sseStatus {
				case sse.SSE_CLIENT_STATUS_CONNECTED:
					c.setStatus(CHANNEL_STATUS_READY)
				default:
					c.setStatus(CHANNEL_STATUS_NOT_READY)
				}
			}
		}
	}()
}

// SubscribeStatusChange 订阅通道状态变更事件
func (c *HttpChannel) SubscribeStatusChange() <-chan ChannelStatus {
	return c.statusEb.Subscribe()
}

// UnsubscribeStatusChange 取消订阅通道状态变更事件
func (c *HttpChannel) UnsubscribeStatusChange(ch <-chan ChannelStatus) {
	c.statusEb.Unsubscribe(ch)
}

// WaitForStatus 等待通道达到指定状态，返回一个通道，当达到目标状态时会收到通知
// timeout为等待超时时间，如果为0则永不超时
func (c *HttpChannel) WaitForStatus(targetStatus ChannelStatus, timeout time.Duration) <-chan struct{} {
	statusCh := c.SubscribeStatusChange()
	cleanup := func() {
		c.UnsubscribeStatusChange(statusCh)
	}
	return util.WaitForStatus(c.GetStatus, targetStatus, statusCh, cleanup, timeout)
}

func (c *HttpChannel) Close() error {
	return c.closeFunc()
}
