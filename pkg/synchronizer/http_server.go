package synchronizer

import (
	"io"
	"net/http"
)

type RapierDbHTTPServer struct {
	channel *HTTPChannel
	server  *http.Server
}

type RapierDbHTTPServerOption struct {
	addr string
}

func NewRapierDbHTTPServer(opt *RapierDbHTTPServerOption) *RapierDbHTTPServer {
	channel := NewHTTPChannel()
	server := &RapierDbHTTPServer{
		channel: channel,
	}

	// 创建路由
	mux := http.NewServeMux()
	mux.HandleFunc("/events", server.handleSSE)
	mux.HandleFunc("/api", server.handleApi)

	// 初始化 HTTP 服务器
	server.server = &http.Server{
		Addr:    opt.addr,
		Handler: mux,
	}

	return server
}

func (s *RapierDbHTTPServer) Start() error {
	return s.server.ListenAndServe()
}

func (s *RapierDbHTTPServer) Stop() error {
	return s.server.Close()
}

// SetMessageHandler 设置消息处理函数
func (s *RapierDbHTTPServer) SetMessageHandler(handler func(clientId string, msg []byte)) {
	s.channel.SetMsgHandler(handler)
}

// handleSSE 处理 SSE 连接请求
func (s *RapierDbHTTPServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "仅支持 GET 方法", http.StatusMethodNotAllowed)
		return
	}

	clientId := r.URL.Query().Get("client_id")
	if clientId == "" {
		http.Error(w, "缺少 client_id", http.StatusBadRequest)
		return
	}

	err := s.channel.Accept(clientId, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 监听连接关闭
	<-r.Context().Done()
	s.channel.Close(clientId)
}

// handleApi 处理客户端发送来的请求
func (s *RapierDbHTTPServer) handleApi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "仅支持 POST 方法", http.StatusMethodNotAllowed)
		return
	}

	clientId := r.URL.Query().Get("client_id")
	if clientId == "" {
		http.Error(w, "缺少 client_id", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "读取请求体失败", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if s.channel.handler == nil {
		s.channel.handler(clientId, body)
	}

	w.WriteHeader(http.StatusOK)
}

// GetChannel
func (s *RapierDbHTTPServer) GetChannel() *HTTPChannel {
	return s.channel
}
