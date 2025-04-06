package network_server

import (
	"fmt"
	"net/http"
	"sync"

	pe "github.com/pkg/errors"
)

var (
	ErrClientNotFound = pe.New("client not found")
)

type HTTPChannel struct {
	clients map[string]http.ResponseWriter
	mu      sync.RWMutex
	handler func(clientId string, msg []byte)
}

func NewHTTPChannel() *HTTPChannel {
	return &HTTPChannel{
		clients: make(map[string]http.ResponseWriter),
		mu:      sync.RWMutex{},
		handler: nil,
	}
}

func (c *HTTPChannel) Setup() error {
	return nil
}

func (c *HTTPChannel) Accept(clientId string, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	c.mu.Lock()
	c.clients[clientId] = w
	c.mu.Unlock()

	w.(http.Flusher).Flush()

	return nil
}

func (c *HTTPChannel) Close(clientId string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.clients[clientId]; !exists {
		return pe.WithStack(fmt.Errorf("%w: 未知客户端 %s", ErrClientNotFound, clientId))
	}

	delete(c.clients, clientId)
	return nil
}

func (c *HTTPChannel) CloseAll() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.clients = make(map[string]http.ResponseWriter)
	return nil
}

func (c *HTTPChannel) Send(clientId string, msg []byte) error {
	// fmt.Printf("server send data=%v, clientId=%s\n", msg, clientId)
	c.mu.RLock()
	client, exists := c.clients[clientId]
	c.mu.RUnlock()

	if !exists {
		return pe.WithStack(fmt.Errorf("%w: 未知客户端 %s", ErrClientNotFound, clientId))
	}

	// 消息按 SSE 格式编码
	encoded := EncodeSSE(msg, nil)
	_, err := client.Write(encoded)
	if err != nil {
		return pe.WithStack(err)
	}

	client.(http.Flusher).Flush()
	return nil
}

func (c *HTTPChannel) Broadcast(msg []byte) error {
	// fmt.Printf("server broadcast data=%v\n", msg)
	c.mu.RLock()
	defer c.mu.RUnlock()

	for clientId, client := range c.clients {
		// 使用 EncodeSSE 函数来正确格式化 SSE 消息
		encoded := EncodeSSE(msg, nil)
		_, err := client.Write(encoded)
		if err != nil {
			return pe.WithStack(fmt.Errorf("发送消息给 %s 失败: %v", clientId, err))
		}
		client.(http.Flusher).Flush()
	}

	return nil
}

func (c *HTTPChannel) GetAllConnectedClientIds() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ids := make([]string, 0, len(c.clients))
	for id := range c.clients {
		ids = append(ids, id)
	}
	return ids
}

func (c *HTTPChannel) SetMsgHandler(handler func(clientId string, msg []byte)) {
	c.mu.Lock()
	// c.handler = func(clientId string, msg []byte) {
	// 	fmt.Printf("server receive data=%v, clientId=%s\n", msg, clientId)
	// 	handler(clientId, msg)
	// }
	c.handler = handler
	c.mu.Unlock()
}
