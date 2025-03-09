package sse

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v5"
)

var (
	headerID    = []byte("id:")
	headerData  = []byte("data:")
	headerEvent = []byte("event:")
	headerRetry = []byte("retry:")
)

func ClientMaxBufferSize(s int) func(c *SseClient) {
	return func(c *SseClient) {
		c.maxBufferSize = s
	}
}

// ConnCallback 定义了在特定连接事件上调用的函数
type ConnCallback func(c *SseClient)

// ResponseValidator 验证响应
type ResponseValidator func(c *SseClient, resp *http.Response) error

// SseClient 处理传入的服务器端发送事件(SSE)流
// 它提供了连接到 SSE 服务器、订阅事件流、处理重连以及管理多个订阅的功能。
// 使用方法:
// 1. 使用 NewClient 创建客户端实例
// 2. 可选择性配置 Headers、重连策略等
// 3. 调用 Subscribe 或 SubscribeChan 方法订阅事件流
// 4. 使用 Unsubscribe 方法取消订阅
type SseClient struct {
	Retry             time.Time                        // 重试连接的时间点
	ReconnectStrategy backoff.BackOff                  // 重连策略，定义了重连的间隔和次数
	disconnectcb      ConnCallback                     // 断开连接时的回调函数
	connectedcb       ConnCallback                     // 成功连接时的回调函数
	subscribed        map[chan *SseEvent]chan struct{} // 管理所有活跃的订阅通道
	Headers           map[string]string                // 发送请求时附加的HTTP头
	ReconnectNotify   backoff.Notify                   // 重连通知回调
	ResponseValidator ResponseValidator                // 自定义响应验证器，用于验证服务器响应
	Connection        *http.Client                     // 用于发送HTTP请求的客户端
	URL               string                           // 服务器端点URL
	LastEventID       atomic.Value                     // 最后接收的事件ID，用于断线重连时恢复
	maxBufferSize     int                              // 读取事件的最大缓冲区大小
	mu                sync.Mutex                       // 保护并发访问的互斥锁
	EncodingBase64    bool                             // 是否使用Base64编码解码事件数据
	Connected         bool                             // 当前连接状态
}

// NewSseClient 创建一个新的 Sse 客户端
func NewSseClient(url string, opts ...func(c *SseClient)) *SseClient {
	c := &SseClient{
		URL:           url,
		Connection:    &http.Client{},
		Headers:       make(map[string]string),
		subscribed:    make(map[chan *SseEvent]chan struct{}),
		maxBufferSize: 1 << 16, // 默认缓冲区大小为64KB
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Subscribe 订阅 SSE 端点
func (c *SseClient) Subscribe(handler func(msg *SseEvent)) error {
	return c.SubscribeWithContext(context.Background(), handler)
}

// SubscribeWithContext 使用上下文订阅 SSE 端点
func (c *SseClient) SubscribeWithContext(ctx context.Context, handler func(msg *SseEvent)) error {
	operation := func() (struct{}, error) {
		resp, err := c.request(ctx)
		if err != nil {
			return struct{}{}, err
		}
		if validator := c.ResponseValidator; validator != nil {
			err = validator(c, resp)
			if err != nil {
				return struct{}{}, err
			}
		} else if resp.StatusCode != 200 {
			resp.Body.Close()
			return struct{}{}, fmt.Errorf("could not connect to endpoint: %s", http.StatusText(resp.StatusCode))
		}
		defer resp.Body.Close()

		reader := NewEventStreamReader(resp.Body, c.maxBufferSize)
		eventChan, errorChan := c.startReadLoop(reader)

		for {
			select {
			case err = <-errorChan:
				return struct{}{}, err
			case msg := <-eventChan:
				handler(msg)
			}
		}
	}

	// 应用重连策略
	var err error
	if c.ReconnectStrategy != nil {
		_, err = backoff.Retry(
			ctx,
			operation,
			backoff.WithNotify(c.ReconnectNotify),
			backoff.WithBackOff(c.ReconnectStrategy),
		)
	} else {
		_, err = backoff.Retry(
			ctx,
			operation,
			backoff.WithNotify(c.ReconnectNotify),
			backoff.WithBackOff(backoff.NewExponentialBackOff()),
		)
	}
	return err
}

// SubscribeChan 将所有事件发送到提供的通道
func (c *SseClient) SubscribeChan(ch chan *SseEvent) error {
	return c.SubscribeChanWithContext(context.Background(), ch)
}

// SubscribeChanWithContext 使用上下文将所有事件发送到提供的通道
func (c *SseClient) SubscribeChanWithContext(ctx context.Context, ch chan *SseEvent) error {
	var connected bool
	errch := make(chan error)
	c.mu.Lock()
	c.subscribed[ch] = make(chan struct{})
	c.mu.Unlock()

	operation := func() (struct{}, error) {
		resp, err := c.request(ctx)
		if err != nil {
			return struct{}{}, err
		}
		if validator := c.ResponseValidator; validator != nil {
			err = validator(c, resp)
			if err != nil {
				return struct{}{}, err
			}
		} else if resp.StatusCode != 200 {
			resp.Body.Close()
			return struct{}{}, fmt.Errorf("could not connect to endpoint: %s", http.StatusText(resp.StatusCode))
		}
		defer resp.Body.Close()

		if !connected {
			errch <- nil
			connected = true
		}

		reader := NewEventStreamReader(resp.Body, c.maxBufferSize)
		eventChan, errorChan := c.startReadLoop(reader)

		for {
			var msg *SseEvent
			// 等待消息到达或退出
			select {
			case <-c.subscribed[ch]:
				return struct{}{}, nil
			case err = <-errorChan:
				return struct{}{}, err
			case msg = <-eventChan:
			}

			// 等待消息被发送或退出
			if msg != nil {
				select {
				case <-c.subscribed[ch]:
					return struct{}{}, nil
				case ch <- msg:
					// 消息已发送
				}
			}
		}
	}

	go func() {
		defer c.cleanup(ch)
		// 应用用户指定的重连策略或默认使用标准的NewExponentialBackOff()重连方法
		var err error
		if c.ReconnectStrategy != nil {
			_, err = backoff.Retry(
				ctx,
				operation,
				backoff.WithNotify(c.ReconnectNotify),
				backoff.WithBackOff(c.ReconnectStrategy),
			)
		} else {
			_, err = backoff.Retry(
				ctx,
				operation,
				backoff.WithNotify(c.ReconnectNotify),
				backoff.WithBackOff(backoff.NewExponentialBackOff()),
			)
		}

		// 一旦连接，通道关闭
		if err != nil && !connected {
			errch <- err
		}
	}()
	err := <-errch
	close(errch)
	return err
}

func (c *SseClient) startReadLoop(reader *EventStreamReader) (chan *SseEvent, chan error) {
	outCh := make(chan *SseEvent)
	erChan := make(chan error)
	go c.readLoop(reader, outCh, erChan)
	return outCh, erChan
}

func (c *SseClient) readLoop(reader *EventStreamReader, outCh chan *SseEvent, erChan chan error) {
	for {
		// 读取每个新行并处理事件类型
		event, err := reader.ReadEvent()
		if err != nil {
			if err == io.EOF {
				erChan <- nil
				return
			}
			// 运行用户指定的断开连接函数
			if c.disconnectcb != nil {
				c.Connected = false
				c.disconnectcb(c)
			}
			erChan <- err
			return
		}

		if !c.Connected && c.connectedcb != nil {
			c.Connected = true
			c.connectedcb(c)
		}

		// 如果我们得到错误，忽略它
		var msg *SseEvent
		if msg, err = c.processEvent(event); err == nil {
			if len(msg.ID) > 0 {
				c.LastEventID.Store(msg.ID)
			} else {
				msg.ID, _ = c.LastEventID.Load().([]byte)
			}

			// 如果事件有有用的内容，则向下游发送
			if msg.hasContent() {
				outCh <- msg
			}
		}
	}
}

// Unsubscribe 取消订阅通道
func (c *SseClient) Unsubscribe(ch chan *SseEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.subscribed[ch] != nil {
		c.subscribed[ch] <- struct{}{}
	}
}

// OnDisconnect 指定连接断开时运行的函数
func (c *SseClient) OnDisconnect(fn ConnCallback) {
	c.disconnectcb = fn
}

// OnConnect 指定连接成功时运行的函数
func (c *SseClient) OnConnect(fn ConnCallback) {
	c.connectedcb = fn
}

func (c *SseClient) request(ctx context.Context) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.URL, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	// 设置SSE必要的HTTP头
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Connection", "keep-alive")

	// 如果有上次事件ID，添加到请求头
	lastID, exists := c.LastEventID.Load().([]byte)
	if exists && lastID != nil {
		req.Header.Set("Last-Event-ID", string(lastID))
	}

	// 添加用户指定的头
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	return c.Connection.Do(req)
}

func (c *SseClient) processEvent(msg []byte) (event *SseEvent, err error) {
	var e SseEvent

	if len(msg) < 1 {
		return nil, errors.New("event message was empty")
	}

	// 将crlf规范化为lf，使分割行更容易
	// 按照规范，用"\n"或"\r"分割行
	for _, line := range bytes.FieldsFunc(msg, func(r rune) bool { return r == '\n' || r == '\r' }) {
		switch {
		case bytes.HasPrefix(line, headerID):
			e.ID = append([]byte(nil), trimHeader(len(headerID), line)...)
		case bytes.HasPrefix(line, headerData):
			// 规范允许每个事件有多个数据字段，用"\n"连接它们
			e.Data = append(e.Data[:], append(trimHeader(len(headerData), line), byte('\n'))...)
		// 规范说，仅包含字符串"data"的行应被视为具有空主体的数据字段
		case bytes.Equal(line, bytes.TrimSuffix(headerData, []byte(":"))):
			e.Data = append(e.Data, byte('\n'))
		case bytes.HasPrefix(line, headerEvent):
			e.Event = append([]byte(nil), trimHeader(len(headerEvent), line)...)
		case bytes.HasPrefix(line, headerRetry):
			e.Retry = append([]byte(nil), trimHeader(len(headerRetry), line)...)
		default:
			// 忽略任何不符合我们要查找的垃圾
		}
	}

	// 根据规范，修剪最后的"\n"
	e.Data = bytes.TrimSuffix(e.Data, []byte("\n"))

	// 如果启用了Base64编码，解码数据
	if c.EncodingBase64 {
		buf := make([]byte, base64.StdEncoding.DecodedLen(len(e.Data)))

		n, err := base64.StdEncoding.Decode(buf, e.Data)
		if err != nil {
			err = fmt.Errorf("failed to decode event message: %s", err)
		}
		e.Data = buf[:n]
	}
	return &e, err
}

func (c *SseClient) cleanup(ch chan *SseEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 清理订阅资源
	if c.subscribed[ch] != nil {
		close(c.subscribed[ch])
		delete(c.subscribed, ch)
	}
}

func trimHeader(size int, data []byte) []byte {
	if data == nil || len(data) < size {
		return data
	}

	data = data[size:]
	// 删除可选的前导空格
	if len(data) > 0 && data[0] == 32 {
		data = data[1:]
	}
	// 删除尾部的换行符
	if len(data) > 0 && data[len(data)-1] == 10 {
		data = data[:len(data)-1]
	}
	return data
}
