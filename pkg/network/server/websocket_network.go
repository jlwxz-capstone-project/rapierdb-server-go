package network_server

import (
	"context"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/auth"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
	pe "github.com/pkg/errors"
)

type WebSocketNetworkOptions struct {
	BaseUrl          string
	WebSocketPath    string
	ShutdownTimeout  time.Duration
	Authenticator    auth.Authenticator[*http.Request]
	AllowOrigin      string
	CheckOrigin      func(*http.Request) bool
	ReadBufferSize   int
	WriteBufferSize  int
	HandshakeTimeout time.Duration
}

type WebSocketConnection struct {
	clientId   string
	remoteAddr string
	conn       *websocket.Conn
	closeCh    chan struct{}
	sendCh     chan []byte
	mu         sync.Mutex
}

type WebSocketNetwork struct {
	server       *http.Server
	options      *WebSocketNetworkOptions
	upgrader     websocket.Upgrader
	status       atomic.Int32
	statusEb     *util.EventBus[NetworkStatus]
	connClosedEb *util.EventBus[ConnectionClosedEvent]
	ctx          context.Context
	cancel       context.CancelFunc
	connections  sync.Map // [string, *WebSocketConnection]
	msgHandler   func(clientId string, msg []byte)
}

var _ NetworkProvider = &WebSocketNetwork{}

func NewWebSocketNetworkWithContext(options *WebSocketNetworkOptions, ctx context.Context) *WebSocketNetwork {
	subCtx, cancel := context.WithCancel(ctx)

	// 设置默认值
	if options.ShutdownTimeout <= 0 {
		options.ShutdownTimeout = 10 * time.Second
	}
	if options.Authenticator == nil {
		options.Authenticator = &auth.HttpMockAuthProvider{}
	}
	if options.AllowOrigin == "" {
		options.AllowOrigin = "*"
	}
	if options.ReadBufferSize <= 0 {
		options.ReadBufferSize = 1024
	}
	if options.WriteBufferSize <= 0 {
		options.WriteBufferSize = 1024
	}
	if options.HandshakeTimeout <= 0 {
		options.HandshakeTimeout = 10 * time.Second
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:   options.ReadBufferSize,
		WriteBufferSize:  options.WriteBufferSize,
		HandshakeTimeout: options.HandshakeTimeout,
		CheckOrigin: func(r *http.Request) bool {
			if options.CheckOrigin != nil {
				return options.CheckOrigin(r)
			}
			// 默认允许所有跨域请求
			return true
		},
	}

	ws := &WebSocketNetwork{
		server:       nil, // 稍后初始化
		options:      options,
		upgrader:     upgrader,
		status:       atomic.Int32{},
		statusEb:     util.NewEventBus[NetworkStatus](),
		connClosedEb: util.NewEventBus[ConnectionClosedEvent](),
		ctx:          subCtx,
		cancel:       cancel,
		connections:  sync.Map{},
		msgHandler:   nil,
	}

	mux := http.NewServeMux()
	mux.HandleFunc(options.WebSocketPath, ws.handleWebSocket)

	// CORS中间件
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", options.AllowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	ws.server = &http.Server{
		Addr:    options.BaseUrl,
		Handler: corsMiddleware(mux),
	}

	return ws
}

func (ws *WebSocketNetwork) Start() error {
	status := NetworkStatus(ws.status.Load())
	if status != NetworkNotStarted {
		return pe.Errorf("Can only start when status is NetworkNotStarted, current status: %s", status.String())
	}

	ws.setStatus(NetworkStarting)

	listener, err := net.Listen("tcp", ws.server.Addr)
	if err != nil {
		ws.setStatus(NetworkStopped)
		return pe.Errorf("Unable to listen on %s: %w", ws.server.Addr, err)
	}

	go ws.stopServerWhenDone()
	go ws.startServerNow(listener)

	return nil
}

func (ws *WebSocketNetwork) Stop() error {
	status := NetworkStatus(ws.status.Load())
	if status != NetworkRunning {
		return pe.Errorf("Can only stop when status is NetworkRunning, current status: %s", status.String())
	}
	ws.cancel()
	return nil
}

func (ws *WebSocketNetwork) CloseConnection(clientId string) error {
	conn, ok := ws.connections.Load(clientId)
	if !ok {
		return pe.Errorf("Connection for client %s not found", clientId)
	}
	if wsConn, ok := conn.(*WebSocketConnection); ok {
		ws.closeWebSocketConnection(wsConn)
	}
	return nil
}

func (ws *WebSocketNetwork) CloseAllConnections() error {
	ws.connections.Range(func(key, value any) bool {
		if wsConn, ok := value.(*WebSocketConnection); ok {
			ws.closeWebSocketConnection(wsConn)
		}
		return true
	})
	return nil
}

func (ws *WebSocketNetwork) Send(clientId string, msg []byte) error {
	conn, ok := ws.connections.Load(clientId)
	if !ok {
		return pe.Errorf("Connection for client %s not found", clientId)
	}
	if wsConn, ok := conn.(*WebSocketConnection); ok {
		select {
		case wsConn.sendCh <- msg:
			return nil
		default:
			return pe.Errorf("Send channel full for client %s", clientId)
		}
	}
	return pe.Errorf("Invalid connection type for client %s", clientId)
}

func (ws *WebSocketNetwork) Broadcast(msg []byte) error {
	ws.connections.Range(func(key, value any) bool {
		if wsConn, ok := value.(*WebSocketConnection); ok {
			select {
			case wsConn.sendCh <- msg:
			default:
				log.Warnf("Failed to send message to client %s: send channel full", wsConn.clientId)
			}
		}
		return true
	})
	return nil
}

func (ws *WebSocketNetwork) SetMsgHandler(handler func(clientId string, msg []byte)) {
	ws.msgHandler = handler
}

func (ws *WebSocketNetwork) GetAllClientIds() []string {
	ids := []string{}
	ws.connections.Range(func(key, value any) bool {
		if _, ok := value.(*WebSocketConnection); ok {
			ids = append(ids, key.(string))
		}
		return true
	})
	return ids
}

func (ws *WebSocketNetwork) GetStatus() NetworkStatus {
	return NetworkStatus(ws.status.Load())
}

func (ws *WebSocketNetwork) SubscribeStatusChange() <-chan NetworkStatus {
	return ws.statusEb.Subscribe()
}

func (ws *WebSocketNetwork) UnsubscribeStatusChange(ch <-chan NetworkStatus) {
	ws.statusEb.Unsubscribe(ch)
}

func (ws *WebSocketNetwork) WaitForStatus(targetStatus NetworkStatus) <-chan struct{} {
	statusCh := ws.SubscribeStatusChange()
	cleanup := func() {
		ws.UnsubscribeStatusChange(statusCh)
	}
	return util.WaitForStatus(ws.GetStatus, targetStatus, statusCh, cleanup, 0)
}

func (ws *WebSocketNetwork) SubscribeConnectionClosed() <-chan ConnectionClosedEvent {
	return ws.connClosedEb.Subscribe()
}

func (ws *WebSocketNetwork) UnsubscribeConnectionClosed(ch <-chan ConnectionClosedEvent) {
	ws.connClosedEb.Unsubscribe(ch)
}

func (ws *WebSocketNetwork) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 认证
	authResult := <-ws.options.Authenticator.Authenticate(r)
	if authResult.Err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Warnf("WebSocket request from %s is unauthorized: %v", r.RemoteAddr, authResult.Err)
		return
	}
	clientId := authResult.ClientID

	// 升级到WebSocket
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Failed to upgrade to WebSocket: %v", err)
		return
	}

	// 如果已有连接，关闭旧连接
	if oldConn, ok := ws.connections.Load(clientId); ok {
		if oldWsConn, ok := oldConn.(*WebSocketConnection); ok {
			ws.closeWebSocketConnection(oldWsConn)
		}
	}

	log.Infof("New WebSocket connection, client %s, remote addr: %s", clientId, r.RemoteAddr)

	// 创建WebSocket连接对象
	wsConn := &WebSocketConnection{
		clientId:   clientId,
		remoteAddr: r.RemoteAddr,
		conn:       conn,
		closeCh:    make(chan struct{}),
		sendCh:     make(chan []byte, 256), // 缓冲队列
	}

	ws.connections.Store(clientId, wsConn)

	// 启动读写协程
	go ws.handleConnectionRead(wsConn)
	go ws.handleConnectionWrite(wsConn)

	// 等待连接关闭
	<-wsConn.closeCh

	// 发布连接关闭事件
	ws.connClosedEb.Publish(ConnectionClosedEvent{
		ClientId: clientId,
	})

	log.Debugf("WebSocket connection for client %s closed, remote addr: %s", clientId, wsConn.remoteAddr)
}

func (ws *WebSocketNetwork) handleConnectionRead(wsConn *WebSocketConnection) {
	defer ws.closeWebSocketConnection(wsConn)

	for {
		messageType, message, err := wsConn.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Errorf("WebSocket read error for client %s: %v", wsConn.clientId, err)
			}
			break
		}

		if messageType == websocket.BinaryMessage || messageType == websocket.TextMessage {
			if ws.msgHandler != nil {
				ws.msgHandler(wsConn.clientId, message)
			}
		}
	}
}

func (ws *WebSocketNetwork) handleConnectionWrite(wsConn *WebSocketConnection) {
	ticker := time.NewTicker(54 * time.Second) // Ping间隔
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-wsConn.sendCh:
			wsConn.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				wsConn.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := wsConn.conn.WriteMessage(websocket.BinaryMessage, message)
			if err != nil {
				log.Errorf("WebSocket write error for client %s: %v", wsConn.clientId, err)
				return
			}

		case <-ticker.C:
			wsConn.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := wsConn.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				log.Errorf("WebSocket ping error for client %s: %v", wsConn.clientId, err)
				return
			}

		case <-wsConn.closeCh:
			return
		}
	}
}

func (ws *WebSocketNetwork) closeWebSocketConnection(wsConn *WebSocketConnection) {
	wsConn.mu.Lock()
	defer wsConn.mu.Unlock()

	select {
	case <-wsConn.closeCh:
		return // 已经关闭
	default:
		close(wsConn.closeCh)
	}

	// 关闭发送通道
	close(wsConn.sendCh)

	// 关闭WebSocket连接
	wsConn.conn.Close()

	// 从连接映射中删除
	ws.connections.Delete(wsConn.clientId)
}

func (ws *WebSocketNetwork) stopServerWhenDone() {
	<-ws.ctx.Done()
	ws.setStatus(NetworkStopping)

	log.Info("Starting graceful shutdown of WebSocket server...")
	ws.CloseAllConnections()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), ws.options.ShutdownTimeout)
	defer cancel()

	if err := ws.server.Shutdown(shutdownCtx); err != nil {
		log.Errorf("WebSocket server graceful shutdown failed: %v", err)
	} else {
		log.Info("WebSocket server gracefully shut down")
	}

	ws.setStatus(NetworkStopped)
	log.Info("WebSocket server shutdown complete")
}

func (ws *WebSocketNetwork) startServerNow(listener net.Listener) {
	log.Info("WebSocket server starting to accept connections...")
	ws.setStatus(NetworkRunning)
	err := ws.server.Serve(listener)

	if err != nil {
		ws.setStatus(NetworkStopped)
		if err != http.ErrServerClosed {
			log.Errorf("WebSocket server Serve goroutine exited abnormally: %v", err)
		}
		return
	}

	log.Info("WebSocket server Serve goroutine exited normally")
}

func (ws *WebSocketNetwork) setStatus(status NetworkStatus) {
	oldStatus := ws.status.Load()
	ws.status.Store(int32(status))

	if oldStatus != int32(status) {
		log.Debugf("WebSocket Network status changed: %v -> %v", NetworkStatus(oldStatus), status)
		ws.statusEb.Publish(status)
	}
}
