package network_server

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/auth"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
	pe "github.com/pkg/errors"
)

type HttpNetworkOptions struct {
	BaseUrl         string
	ReceiveEndpoint string
	SendEndpoint    string
	ShutdownTimeout time.Duration
	Authenticator   auth.Authenticator[*http.Request]
}

type HttpConnection struct {
	clientId       string
	remoteAddr     string
	closeCh        chan struct{}
	responseWriter http.ResponseWriter
}

type HttpNetwork struct {
	server       *http.Server
	options      *HttpNetworkOptions
	status       atomic.Int32
	statusEb     *util.EventBus[NetworkStatus]
	connClosedEb *util.EventBus[ConnectionClosedEvent]
	ctx          context.Context
	cancel       context.CancelFunc
	connections  sync.Map //[string, HttpConnection]
	msgHandler   func(clientId string, msg []byte)
}

var _ NetworkProvider = &HttpNetwork{}

func (s *HttpNetwork) ensureOptionsValid() {
	if s.options.ShutdownTimeout <= 0 {
		log.Info("ShutdownTimeout is not set, using default value 10s")
		s.options.ShutdownTimeout = 10 * time.Second
	}
	if s.options.Authenticator == nil {
		log.Info("Authenticator is not set, using HttpMockAuthProvider")
		s.options.Authenticator = &auth.HttpMockAuthProvider{}
	}
}

func NewHttpNetworkWithContext(options *HttpNetworkOptions, ctx context.Context) *HttpNetwork {
	subCtx, cancel := context.WithCancel(ctx)

	s := &HttpNetwork{
		server:       nil, // init later
		options:      options,
		status:       atomic.Int32{},
		statusEb:     util.NewEventBus[NetworkStatus](),
		connClosedEb: util.NewEventBus[ConnectionClosedEvent](),
		ctx:          subCtx,
		cancel:       cancel,
		connections:  sync.Map{},
		msgHandler:   nil,
	}
	s.ensureOptionsValid()

	mux := http.NewServeMux()
	mux.HandleFunc(options.ReceiveEndpoint, s.handleReceive) // endpoint for client->server
	mux.HandleFunc(options.SendEndpoint, s.handleSend)       // endpoint for server->client (by SSE)
	s.server = &http.Server{
		Addr:    options.BaseUrl,
		Handler: mux,
	}

	return s
}

func (s *HttpNetwork) Start() error {
	status := NetworkStatus(s.status.Load())
	if status != NetworkNotStarted {
		return pe.Errorf("Can only start when status is ServerStatusNotStarted, current status: %s", status.String())
	}

	s.setStatus(NetworkStarting)

	listener, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		s.setStatus(NetworkStopped)
		return pe.Errorf("Unable to listen on %s: %w", s.server.Addr, err)
	}

	go s.stopServerWhenDone()
	go s.startServerNow(listener)

	return nil
}

func (s *HttpNetwork) Stop() error {
	status := NetworkStatus(s.status.Load())
	if status != NetworkRunning {
		return pe.Errorf("Can only stop when status is ServerStatusRunning, current status: %s", status.String())
	}
	s.cancel()
	return nil
}

func (s *HttpNetwork) CloseConnection(clientId string) error {
	conn, ok := s.connections.Load(clientId)
	if !ok {
		return pe.Errorf("Connection for client %s not found", clientId)
	}
	if conn, ok := conn.(HttpConnection); ok {
		s.closeHttpConnection(&conn)
	}
	return nil
}

func (s *HttpNetwork) CloseAllConnections() error {
	s.connections.Range(func(key, value any) bool {
		if conn, ok := value.(HttpConnection); ok {
			conn.closeCh <- struct{}{}
			// close(conn.closeCh)
		}
		return true
	})
	s.connections.Clear()
	return nil
}

func (s *HttpNetwork) Send(clientId string, msg []byte) error {
	conn, ok := s.connections.Load(clientId)
	if !ok {
		return pe.Errorf("Connection for client %s not found", clientId)
	}
	if conn, ok := conn.(HttpConnection); ok {
		base64Msg := base64.StdEncoding.EncodeToString(msg)
		encodedMsg := EncodeSSEBase64(base64Msg, nil)
		_, err := conn.responseWriter.Write(encodedMsg)
		if err != nil {
			return pe.Wrapf(err, "Failed to send message to client %s", clientId)
		}
		conn.responseWriter.(http.Flusher).Flush()
	}
	return nil
}

func (s *HttpNetwork) Broadcast(msg []byte) error {
	base64Msg := base64.StdEncoding.EncodeToString(msg)
	encodedMsg := EncodeSSEBase64(base64Msg, nil)
	s.connections.Range(func(key, value any) bool {
		if conn, ok := value.(HttpConnection); ok {
			_, err := conn.responseWriter.Write(encodedMsg)
			if err == nil {
				conn.responseWriter.(http.Flusher).Flush()
			}
		}
		return true
	})
	return nil
}

func (s *HttpNetwork) SetMsgHandler(handler func(clientId string, msg []byte)) {
	s.msgHandler = handler
}

func (s *HttpNetwork) GetAllClientIds() []string {
	ids := []string{}
	s.connections.Range(func(key, value any) bool {
		if _, ok := value.(HttpConnection); ok {
			ids = append(ids, key.(string))
		}
		return true
	})
	return ids
}

func (s *HttpNetwork) GetStatus() NetworkStatus {
	return NetworkStatus(s.status.Load())
}

func (s *HttpNetwork) SubscribeStatusChange() <-chan NetworkStatus {
	return s.statusEb.Subscribe()
}

func (s *HttpNetwork) UnsubscribeStatusChange(ch <-chan NetworkStatus) {
	s.statusEb.Unsubscribe(ch)
}

func (s *HttpNetwork) WaitForStatus(status NetworkStatus) <-chan struct{} {
	statusCh := s.SubscribeStatusChange()
	cleanup := func() {
		s.UnsubscribeStatusChange(statusCh)
	}
	return util.WaitForStatus(s.GetStatus, status, statusCh, cleanup, 0)
}

func (s *HttpNetwork) SubscribeConnectionClosed() <-chan ConnectionClosedEvent {
	return s.connClosedEb.Subscribe()
}

func (s *HttpNetwork) UnsubscribeConnectionClosed(ch <-chan ConnectionClosedEvent) {
	s.connClosedEb.Unsubscribe(ch)
}

func (s *HttpNetwork) stopServerWhenDone() {
	<-s.ctx.Done()
	s.setStatus(NetworkStopping)

	log.Info("Starting graceful shutdown of HTTP server...")
	s.CloseAllConnections()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.options.ShutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		log.Errorf("HTTP server graceful shutdown failed: %v", err)
	} else {
		log.Info("HTTP server gracefully shut down")
	}

	s.setStatus(NetworkStopped)
	log.Info("HTTP server shutdown complete")
}

func (s *HttpNetwork) startServerNow(listener net.Listener) {
	log.Info("HTTP server starting to accept connections...")
	s.setStatus(NetworkRunning)
	err := s.server.Serve(listener)

	if err != nil {
		s.setStatus(NetworkStopped)
		if !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("HTTP server Serve goroutine exited abnormally: %v", err)
		}
		return
	}

	log.Info("HTTP server Serve goroutine exited normally")
}

func (s *HttpNetwork) closeHttpConnection(conn *HttpConnection) {
	conn.closeCh <- struct{}{}
	// close(conn.closeCh)
	s.connections.Delete(conn.clientId)
}

func (s *HttpNetwork) handleReceive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		log.Warnf("invalid request to receive endpoint. expect POST, but got %s", r.Method)
		return
	}

	// get client id from authenticator
	authResult := <-s.options.Authenticator.Authenticate(r)
	if authResult.Err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Warnf("receive request from %s is unauthorized: %v", r.RemoteAddr, authResult.Err)
		return
	}
	clientId := authResult.ClientID

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		log.Errorf("failed to read request body: %v", err)
		return
	}
	defer r.Body.Close()

	log.Infof("new api call from client %s, remote addr: %s", clientId, r.RemoteAddr)

	if s.msgHandler != nil {
		// log.Debugf("server recv: %v", body)
		s.msgHandler(clientId, body)
	} else {
		log.Warn("no message handler set, a message is ignored")
	}

	w.WriteHeader(http.StatusOK)
}

func (s *HttpNetwork) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		log.Warnf("invalid request to sse endpoint. expect GET, but got %s", r.Method)
		return
	}

	// get client id from authenticator
	authResult := <-s.options.Authenticator.Authenticate(r)
	if authResult.Err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Warnf("sse request from %s is unauthorized: %v", r.RemoteAddr, authResult.Err)
		return
	}
	clientId := authResult.ClientID

	// if old connection exists, close it first
	oldConn, ok := s.connections.Load(clientId)
	if ok {
		if oldConn, ok := oldConn.(HttpConnection); ok {
			s.closeHttpConnection(&oldConn)
		}
	}

	log.Infof("new sse connection, client %s, remote addr: %s", clientId, r.RemoteAddr)

	// notice: don't put following code in new goroutine, because w http.ResponseWriter
	// and r *http.Request will be closed after this fuction returns.
	// some headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.(http.Flusher).Flush()

	conn := HttpConnection{
		clientId:       clientId,
		remoteAddr:     r.RemoteAddr,
		closeCh:        make(chan struct{}),
		responseWriter: w,
	}
	s.connections.Store(clientId, conn)

	// listen to request context done, and close connection
	// when client close the connection, the request context will be done
	go func() {
		<-r.Context().Done()
		s.closeHttpConnection(&conn)
	}()

	<-conn.closeCh

	// publish connection closed event
	s.connClosedEb.Publish(ConnectionClosedEvent{
		ClientId: clientId,
	})

	log.Debugf("Connection for client %s closed, remote addr: %s", clientId, conn.remoteAddr)
}

func (s *HttpNetwork) setStatus(status NetworkStatus) {
	oldStatus := s.status.Load()
	s.status.Store(int32(status))

	if oldStatus != int32(status) {
		log.Debugf("Server HttpNetwork status changed: %v -> %v", NetworkStatus(oldStatus), status)
		s.statusEb.Publish(status)
	}
}
