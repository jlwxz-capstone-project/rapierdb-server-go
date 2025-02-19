package synchronizer

import (
	"fmt"
	"net/http"
	"sync"
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
		return fmt.Errorf("client %s not found", clientId)
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
	c.mu.RLock()
	client, exists := c.clients[clientId]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("client %s not found", clientId)
	}

	// 消息按 SSE 格式编码
	encoded := EncodeSSE(msg, nil)
	_, err := client.Write(encoded)
	if err != nil {
		return err
	}

	client.(http.Flusher).Flush()
	return nil
}

func (c *HTTPChannel) Broadcast(msg []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for clientId, client := range c.clients {
		_, err := fmt.Fprintf(client, "data: %s\n\n", msg)
		if err != nil {
			return fmt.Errorf("failed to broadcast to client %s: %v", clientId, err)
		}
		client.(http.Flusher).Flush()
	}

	return nil
}

func (c *HTTPChannel) SetMsgHandler(handler func(clientId string, msg []byte)) {
	c.mu.Lock()
	c.handler = handler
	c.mu.Unlock()
}
