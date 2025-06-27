package network_server

import (
	"context"
	"net/http"
	"time"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/auth"
)

// NetworkType 网络类型枚举
type NetworkType string

const (
	NetworkTypeHTTP      NetworkType = "http"      // HTTP + SSE
	NetworkTypeWebSocket NetworkType = "websocket" // WebSocket
)

// UnifiedNetworkOptions 统一的网络配置选项
type UnifiedNetworkOptions struct {
	NetworkType NetworkType // 网络类型: "http" 或 "websocket"

	// 通用配置
	BaseUrl         string
	ShutdownTimeout time.Duration
	Authenticator   auth.Authenticator[*http.Request]
	AllowOrigin     string

	// HTTP特定配置 (用于SSE)
	ReceiveEndpoint  string // API端点，如 "/api"
	SendEndpoint     string // SSE端点，如 "/sse"
	AllowMethods     string
	AllowHeaders     string
	AllowCredentials bool

	// WebSocket特定配置
	WebSocketPath    string                   // WebSocket端点，如 "/ws"
	CheckOrigin      func(*http.Request) bool // 自定义跨域检查
	ReadBufferSize   int
	WriteBufferSize  int
	HandshakeTimeout time.Duration
}

// CreateNetworkProvider 根据配置创建网络提供者
func CreateNetworkProvider(options *UnifiedNetworkOptions, ctx context.Context) NetworkProvider {
	switch options.NetworkType {
	case NetworkTypeWebSocket:
		return createWebSocketNetwork(options, ctx)
	case NetworkTypeHTTP:
		fallthrough
	default:
		return createHttpNetwork(options, ctx)
	}
}

// createWebSocketNetwork 创建WebSocket网络提供者
func createWebSocketNetwork(options *UnifiedNetworkOptions, ctx context.Context) NetworkProvider {
	wsOptions := &WebSocketNetworkOptions{
		BaseUrl:          options.BaseUrl,
		WebSocketPath:    options.WebSocketPath,
		ShutdownTimeout:  options.ShutdownTimeout,
		Authenticator:    options.Authenticator,
		AllowOrigin:      options.AllowOrigin,
		CheckOrigin:      options.CheckOrigin,
		ReadBufferSize:   options.ReadBufferSize,
		WriteBufferSize:  options.WriteBufferSize,
		HandshakeTimeout: options.HandshakeTimeout,
	}

	// 设置默认值
	if wsOptions.WebSocketPath == "" {
		wsOptions.WebSocketPath = "/ws"
	}

	return NewWebSocketNetworkWithContext(wsOptions, ctx)
}

// createHttpNetwork 创建HTTP+SSE网络提供者
func createHttpNetwork(options *UnifiedNetworkOptions, ctx context.Context) NetworkProvider {
	httpOptions := &HttpNetworkOptions{
		BaseUrl:          options.BaseUrl,
		ReceiveEndpoint:  options.ReceiveEndpoint,
		SendEndpoint:     options.SendEndpoint,
		ShutdownTimeout:  options.ShutdownTimeout,
		Authenticator:    options.Authenticator,
		AllowOrigin:      options.AllowOrigin,
		AllowMethods:     options.AllowMethods,
		AllowHeaders:     options.AllowHeaders,
		AllowCredentials: options.AllowCredentials,
	}

	// 设置默认值
	if httpOptions.ReceiveEndpoint == "" {
		httpOptions.ReceiveEndpoint = "/api"
	}
	if httpOptions.SendEndpoint == "" {
		httpOptions.SendEndpoint = "/sse"
	}

	return NewHttpNetworkWithContext(httpOptions, ctx)
}

// DefaultWebSocketOptions 默认WebSocket配置
func DefaultWebSocketOptions(baseUrl string) *UnifiedNetworkOptions {
	return &UnifiedNetworkOptions{
		NetworkType:   NetworkTypeWebSocket,
		BaseUrl:       baseUrl,
		WebSocketPath: "/ws",
		AllowOrigin:   "*",
	}
}

// DefaultHttpOptions 默认HTTP+SSE配置
func DefaultHttpOptions(baseUrl string) *UnifiedNetworkOptions {
	return &UnifiedNetworkOptions{
		NetworkType:      NetworkTypeHTTP,
		BaseUrl:          baseUrl,
		ReceiveEndpoint:  "/api",
		SendEndpoint:     "/sse",
		AllowOrigin:      "*",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowHeaders:     "*",
		AllowCredentials: true,
	}
}
