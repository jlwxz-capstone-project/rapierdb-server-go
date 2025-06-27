package network_client

import (
	"context"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
	pe "github.com/pkg/errors"
)

type WebSocketNetworkOptions struct {
	ServerUrl        string
	WebSocketPath    string
	Headers          map[string][]string
	HandshakeTimeout time.Duration
	ReadBufferSize   int
	WriteBufferSize  int
	ReconnectDelay   time.Duration
}

type WebSocketNetwork struct {
	options    *WebSocketNetworkOptions
	conn       *websocket.Conn
	ctx        context.Context
	cancel     context.CancelFunc
	msgHandler func(msg []byte)
	status     atomic.Int32
	statusEb   *util.EventBus[NetworkStatus]
	sendCh     chan []byte
	connected  bool
}

var _ NetworkProvider = &WebSocketNetwork{}

func NewWebSocketNetwork(options *WebSocketNetworkOptions) *WebSocketNetwork {
	return NewWebSocketNetworkWithContext(options, context.Background())
}

func NewWebSocketNetworkWithContext(options *WebSocketNetworkOptions, ctx context.Context) *WebSocketNetwork {
	subCtx, cancel := context.WithCancel(ctx)

	// 设置默认值
	if options.HandshakeTimeout <= 0 {
		options.HandshakeTimeout = 10 * time.Second
	}
	if options.ReadBufferSize <= 0 {
		options.ReadBufferSize = 1024
	}
	if options.WriteBufferSize <= 0 {
		options.WriteBufferSize = 1024
	}
	if options.ReconnectDelay <= 0 {
		options.ReconnectDelay = 3 * time.Second
	}
	if options.Headers == nil {
		options.Headers = make(map[string][]string)
	}

	wn := &WebSocketNetwork{
		options:  options,
		ctx:      subCtx,
		cancel:   cancel,
		status:   atomic.Int32{},
		statusEb: util.NewEventBus[NetworkStatus](),
		sendCh:   make(chan []byte, 256),
	}

	return wn
}

func (wn *WebSocketNetwork) Connect() error {
	status := NetworkStatus(wn.status.Load())
	if status == NetworkClosed {
		return pe.Errorf("network is already closed")
	}
	if status == NetworkReady {
		return pe.Errorf("network is already connected")
	}

	wn.setStatus(NetworkNotReady)

	// 构造WebSocket URL
	u, err := url.Parse(wn.options.ServerUrl)
	if err != nil {
		return pe.Errorf("invalid server URL: %w", err)
	}

	if u.Scheme == "http" {
		u.Scheme = "ws"
	} else if u.Scheme == "https" {
		u.Scheme = "wss"
	}
	u.Path = wn.options.WebSocketPath

	// 创建WebSocket拨号器
	dialer := websocket.Dialer{
		HandshakeTimeout: wn.options.HandshakeTimeout,
		ReadBufferSize:   wn.options.ReadBufferSize,
		WriteBufferSize:  wn.options.WriteBufferSize,
	}

	// 连接WebSocket
	conn, _, err := dialer.Dial(u.String(), wn.options.Headers)
	if err != nil {
		wn.setStatus(NetworkNotReady)
		return pe.Errorf("failed to connect to WebSocket: %w", err)
	}

	wn.conn = conn
	wn.connected = true
	wn.setStatus(NetworkReady)

	// 启动读写协程
	go wn.readLoop()
	go wn.writeLoop()

	log.Infof("WebSocket client connected to %s", u.String())
	return nil
}

func (wn *WebSocketNetwork) Close() error {
	status := NetworkStatus(wn.status.Load())
	if status == NetworkClosed {
		return nil
	}

	wn.connected = false
	wn.cancel()

	if wn.conn != nil {
		wn.conn.Close()
	}

	wn.setStatus(NetworkClosed)
	log.Info("WebSocket client closed")
	return nil
}

func (wn *WebSocketNetwork) Send(msg []byte) error {
	status := NetworkStatus(wn.status.Load())
	if status != NetworkReady {
		return pe.Errorf("network is not ready, current status: %v", status)
	}

	select {
	case wn.sendCh <- msg:
		return nil
	case <-wn.ctx.Done():
		return pe.Errorf("network is closing")
	default:
		return pe.Errorf("send channel is full")
	}
}

func (wn *WebSocketNetwork) SetMsgHandler(handler func(msg []byte)) {
	wn.msgHandler = handler
}

func (wn *WebSocketNetwork) GetStatus() NetworkStatus {
	return NetworkStatus(wn.status.Load())
}

func (wn *WebSocketNetwork) SubscribeStatusChange() <-chan NetworkStatus {
	return wn.statusEb.Subscribe()
}

func (wn *WebSocketNetwork) UnsubscribeStatusChange(ch <-chan NetworkStatus) {
	wn.statusEb.Unsubscribe(ch)
}

func (wn *WebSocketNetwork) readLoop() {
	defer func() {
		wn.connected = false
		wn.setStatus(NetworkNotReady)
		if wn.conn != nil {
			wn.conn.Close()
		}
	}()

	for wn.connected {
		messageType, message, err := wn.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Errorf("WebSocket read error: %v", err)
			}
			break
		}

		if messageType == websocket.BinaryMessage || messageType == websocket.TextMessage {
			if wn.msgHandler != nil {
				wn.msgHandler(message)
			}
		}
	}
}

func (wn *WebSocketNetwork) writeLoop() {
	ticker := time.NewTicker(54 * time.Second) // Ping间隔
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-wn.sendCh:
			if !ok {
				wn.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			wn.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := wn.conn.WriteMessage(websocket.BinaryMessage, message)
			if err != nil {
				log.Errorf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			wn.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := wn.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				log.Errorf("WebSocket ping error: %v", err)
				return
			}

		case <-wn.ctx.Done():
			wn.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		}
	}
}

func (wn *WebSocketNetwork) setStatus(status NetworkStatus) {
	oldStatus := wn.status.Load()
	wn.status.Store(int32(status))

	if oldStatus != int32(status) {
		log.Debugf("WebSocket client status changed: %v -> %v", NetworkStatus(oldStatus), status)
		wn.statusEb.Publish(status)
	}
}
