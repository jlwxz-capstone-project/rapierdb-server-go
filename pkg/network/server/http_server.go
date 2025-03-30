package network_server

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/auth"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
	pe "github.com/pkg/errors"
)

var (
	ErrAuthProviderNotSet = errors.New("auth provider not set")
)

// 服务器状态常量
type ServerStatus string

const (
	ServerStatusStopped  ServerStatus = "stopped"
	ServerStatusStarting ServerStatus = "starting"
	ServerStatusRunning  ServerStatus = "running"
	ServerStatusStopping ServerStatus = "stopping"
)

type RapierDbHTTPServer struct {
	channel      *HTTPChannel
	server       *http.Server
	authProvider auth.Authenticator[*http.Request]

	// 状态相关字段
	status     ServerStatus
	statusLock sync.RWMutex
	statusEb   *util.EventBus[ServerStatus]
}

type RapierDbHTTPServerOption struct {
	Addr        string
	SseEndpoint string
	ApiEndpoint string
}

func NewRapierDbHTTPServer(opt *RapierDbHTTPServerOption) *RapierDbHTTPServer {
	channel := NewHTTPChannel()
	server := &RapierDbHTTPServer{
		channel:  channel,
		status:   ServerStatusStopped,
		statusEb: util.NewEventBus[ServerStatus](),
	}

	// 创建路由
	mux := http.NewServeMux()
	mux.HandleFunc(opt.SseEndpoint, server.handleSSE)
	mux.HandleFunc(opt.ApiEndpoint, server.handleApi)

	// 初始化 HTTP 服务器
	server.server = &http.Server{
		Addr:    opt.Addr,
		Handler: mux,
	}

	return server
}

// GetStatus 获取服务器当前状态
func (s *RapierDbHTTPServer) GetStatus() ServerStatus {
	s.statusLock.RLock()
	defer s.statusLock.RUnlock()
	return s.status
}

// setStatus 设置服务器状态并通知订阅者
func (s *RapierDbHTTPServer) setStatus(status ServerStatus) {
	s.statusLock.Lock()
	oldStatus := s.status
	s.status = status
	s.statusLock.Unlock()

	// 只有状态发生变化时才发布事件
	if oldStatus != status {
		// 通过事件总线发布状态变更事件
		s.statusEb.Publish(status)
	}
}

// SubscribeStatusChange 订阅状态变更事件
func (s *RapierDbHTTPServer) SubscribeStatusChange() <-chan ServerStatus {
	return s.statusEb.Subscribe()
}

// UnsubscribeStatusChange 取消订阅状态变更事件
func (s *RapierDbHTTPServer) UnsubscribeStatusChange(ch <-chan ServerStatus) {
	s.statusEb.Unsubscribe(ch)
}

// WaitForStatus 等待服务器达到指定状态，返回一个通道，当达到目标状态时会收到通知
func (s *RapierDbHTTPServer) WaitForStatus(targetStatus ServerStatus) <-chan struct{} {
	statusCh := s.SubscribeStatusChange()
	cleanup := func() {
		s.UnsubscribeStatusChange(statusCh)
	}
	return util.WaitForStatus(s.GetStatus, targetStatus, statusCh, cleanup, 0)
}

func (s *RapierDbHTTPServer) Start() error {
	if s.authProvider == nil {
		return pe.WithStack(ErrAuthProviderNotSet)
	}

	s.setStatus(ServerStatusStarting)

	// 在新的 goroutine 中启动服务器
	go func() {
		err := s.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Errorf("HTTP服务器错误: %v", err)
			s.setStatus(ServerStatusStopped)
		}
	}()

	// 设置状态为运行中
	s.setStatus(ServerStatusRunning)
	return nil
}

func (s *RapierDbHTTPServer) Stop() error {
	s.setStatus(ServerStatusStopping)
	err := s.server.Close()
	s.setStatus(ServerStatusStopped)
	return err
}

func (s *RapierDbHTTPServer) SetAuthProvider(authProvider auth.Authenticator[*http.Request]) {
	s.authProvider = authProvider
}

// SetMessageHandler 设置消息处理函数
func (s *RapierDbHTTPServer) SetMessageHandler(handler func(clientId string, msg []byte)) {
	s.channel.SetMsgHandler(handler)
}

// handleSSE 处理 SSE 连接请求
func (s *RapierDbHTTPServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		msg := fmt.Sprintf("SSE 只接受 GET 请求，当前请求方法为 %s", r.Method)
		http.Error(w, msg, http.StatusMethodNotAllowed)
		log.Errorf("非法的 SSE 请求，方法为 %s", r.Method)
		return
	}

	// 从 AuthProvider 获取客户端 ID
	authResult := <-s.authProvider.Authenticate(r)
	if authResult.Err != nil {
		http.Error(w, "认证失败", http.StatusUnauthorized)
		log.Errorf("SSE 请求认证失败，错误为 %v", authResult.Err)
		return
	}
	clientId := authResult.ClientID

	// 委托给 HTTPChannel 处理
	err := s.channel.Accept(clientId, w)
	if err != nil {
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		log.Errorf("信道接受失败，错误为 %v", err)
		return
	}

	// 监听连接关闭
	<-r.Context().Done()
	s.channel.Close(clientId)
}

// handleApi 处理客户端发送来的请求
func (s *RapierDbHTTPServer) handleApi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "API 接口只接受 POST 请求", http.StatusMethodNotAllowed)
		log.Errorf("非法的 API 请求，方法为 %s", r.Method)
		return
	}

	// 从 AuthProvider 获取客户端 ID
	authResult := <-s.authProvider.Authenticate(r)
	if authResult.Err != nil {
		http.Error(w, "认证失败", http.StatusUnauthorized)
		log.Errorf("API 请求认证失败，错误为 %v", authResult.Err)
		return
	}
	clientId := authResult.ClientID

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		log.Errorf("读取请求体失败，错误为 %v", err)
		return
	}
	defer r.Body.Close()

	if s.channel.handler != nil {
		s.channel.handler(clientId, body)
	}

	w.WriteHeader(http.StatusOK)
}

// GetChannel
func (s *RapierDbHTTPServer) GetChannel() *HTTPChannel {
	return s.channel
}
