package network_server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/auth"
	pe "github.com/pkg/errors"
)

var (
	ErrAuthProviderNotSet = errors.New("auth provider not set")
)

type RapierDbHTTPServer struct {
	channel      *HTTPChannel
	server       *http.Server
	authProvider auth.AuthProvider[*http.Request]
}

type RapierDbHTTPServerOption struct {
	Addr        string
	SseEndpoint string
	ApiEndpoint string
}

func NewRapierDbHTTPServer(opt *RapierDbHTTPServerOption) *RapierDbHTTPServer {
	channel := NewHTTPChannel()
	server := &RapierDbHTTPServer{
		channel: channel,
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

func (s *RapierDbHTTPServer) Start() error {
	if s.authProvider == nil {
		return pe.WithStack(ErrAuthProviderNotSet)
	}
	return s.server.ListenAndServe()
}

func (s *RapierDbHTTPServer) Stop() error {
	return s.server.Close()
}

func (s *RapierDbHTTPServer) SetAuthProvider(authProvider auth.AuthProvider[*http.Request]) {
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
		log.Printf("非法的 SSE 请求，方法为 %s", r.Method)
		return
	}

	// 从 AuthProvider 获取客户端 ID
	authResult := <-s.authProvider.GetClientId(r)
	if authResult.Err != nil {
		http.Error(w, "认证失败", http.StatusUnauthorized)
		log.Printf("SSE 请求认证失败，错误为 %v", authResult.Err)
		return
	}
	clientId := authResult.ClientID

	// 委托给 HTTPChannel 处理
	err := s.channel.Accept(clientId, w)
	if err != nil {
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		log.Printf("信道接受失败，错误为 %v", err)
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
		log.Printf("非法的 API 请求，方法为 %s", r.Method)
		return
	}

	// 从 AuthProvider 获取客户端 ID
	authResult := <-s.authProvider.GetClientId(r)
	if authResult.Err != nil {
		http.Error(w, "认证失败", http.StatusUnauthorized)
		log.Printf("API 请求认证失败，错误为 %v", authResult.Err)
		return
	}
	clientId := authResult.ClientID

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "服务器错误", http.StatusInternalServerError)
		log.Printf("读取请求体失败，错误为 %v", err)
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
