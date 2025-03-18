package network_client

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	pe "github.com/pkg/errors"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/sse"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// HttpChannelConfig 配置 HTTP 通道的参数
type HttpChannelConfig struct {
	ReceiveURL string            // 接收消息的URL（SSE端点）
	SendURL    string            // 发送消息的URL
	Headers    map[string]string // HTTP请求头
}

// HttpChannelManager 实现 ClientChannelManager 接口，管理 HttpChannel
type HttpChannelManager struct {
	config       HttpChannelConfig     // 通道配置
	channels     map[*HttpChannel]bool // 活跃的通道列表
	channelMutex sync.Mutex            // 保护通道列表的互斥锁
	httpClient   *http.Client          // 共享的HTTP客户端
	ctx          context.Context       // 上下文
	cancel       context.CancelFunc    // 取消函数
}

// NewHttpChannelManager 创建一个新的HTTP通道管理器
func NewHttpChannelManager(config HttpChannelConfig) *HttpChannelManager {
	return &HttpChannelManager{
		config:     config,
		channels:   make(map[*HttpChannel]bool),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetChannel 获取一个新的 HttpChannel
func (m *HttpChannelManager) GetChannel() (ClientChannel, error) {
	// Channel 的 context 继承自 HttpChannelManager 的 context
	ctx, cancel := context.WithCancel(m.ctx)
	sseClient := sse.NewSseClient(m.config.ReceiveURL)

	// 设置请求头
	for k, v := range m.config.Headers {
		sseClient.Headers[k] = v
	}

	// 创建HTTP通道
	channel := &HttpChannel{
		sendFunc:   nil, // 下面设置
		closeFunc:  nil, // 下面设置
		sseClient:  nil, // 由 BindSseClient 设置
		ctx:        ctx,
		cancel:     cancel,
		msgHandler: nil, // 由 Channel 的使用者使用 SetMsgHandler 设置
		mutex:      sync.RWMutex{},

		status:     CHANNEL_STATUS_NOT_READY,
		statusLock: sync.RWMutex{},
		statusEb:   util.NewEventBus[ChannelStatus](),
	}
	channel.BindSseClient(sseClient)
	channel.sendFunc = func(msg []byte) error {
		return m.sendMessage(ctx, msg)
	}
	channel.closeFunc = func() error {
		return m.Close(channel)
	}

	// 启动SSE接收
	err := sseClient.Subscribe(channel.HandleSseEvent)
	if err != nil {
		cancel()
		return nil, pe.WithStack(fmt.Errorf("启动SSE接收失败: %w", err))
	}

	// 注册通道到管理器
	m.channelMutex.Lock()
	m.channels[channel] = true
	m.channelMutex.Unlock()

	return channel, nil
}

// sendMessage 发送消息到服务器
// HttpChannelManager 管理的所有 HttpChannel 共享同一个 httpClient，都是用这个方法发送消息
func (m *HttpChannelManager) sendMessage(ctx context.Context, msg []byte) error {
	req, err := http.NewRequestWithContext(ctx, "POST", m.config.SendURL, bytes.NewReader(msg))
	if err != nil {
		return pe.WithStack(fmt.Errorf("创建HTTP请求失败: %w", err))
	}

	// 设置内容类型
	req.Header.Set("Content-Type", "application/json")

	// 设置其他头信息
	for k, v := range m.config.Headers {
		req.Header.Set(k, v)
	}

	// 发送请求
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return pe.WithStack(fmt.Errorf("发送HTTP请求失败: %w", err))
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return pe.WithStack(fmt.Errorf("HTTP请求返回非成功状态码: %d", resp.StatusCode))
	}

	return nil
}

// Close 关闭并销毁一个通信通道
func (m *HttpChannelManager) Close(channel ClientChannel) error {
	httpChannel, ok := channel.(*HttpChannel)
	if !ok {
		return pe.WithStack(fmt.Errorf("无效的通道类型，期望HttpChannel"))
	}

	// 从管理器中移除通道
	m.channelMutex.Lock()
	exists := m.channels[httpChannel]
	if exists {
		delete(m.channels, httpChannel)
	}
	m.channelMutex.Unlock()

	if !exists {
		return pe.WithStack(fmt.Errorf("通道未找到或已关闭"))
	}

	// 取消上下文
	httpChannel.cancel()

	// 关闭SSE客户端
	httpChannel.sseClient.Close()

	return nil
}

// CloseAll 关闭所有活跃的通道
func (m *HttpChannelManager) CloseAll() {
	m.channelMutex.Lock()
	channels := make([]*HttpChannel, 0, len(m.channels))
	for ch := range m.channels {
		channels = append(channels, ch)
	}
	m.channelMutex.Unlock()

	// 关闭所有通道
	for _, ch := range channels {
		m.Close(ch)
	}
}

// GetActiveChannelsCount 获取活跃通道数量
func (m *HttpChannelManager) GetActiveChannelsCount() int {
	m.channelMutex.Lock()
	defer m.channelMutex.Unlock()
	return len(m.channels)
}
