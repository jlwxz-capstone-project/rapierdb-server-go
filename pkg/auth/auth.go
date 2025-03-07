package auth

import (
	"net/http"
)

type AuthResult struct {
	ClientID string
	Err      error
}

// AuthProvider 是一个认证提供者，用于从上下文中获取客户端 ID
// 注意返回的是一个通道（异步值），这使得它可以容易地与外部
// 系统，如单独的鉴权服务进行集成
type AuthProvider[CTX any] interface {
	GetClientId(ctx CTX) <-chan AuthResult
}

// HttpMockAuthProvider 实现了 AuthProvider 接口，专门用于 HTTP 请求认证
type HttpMockAuthProvider struct{}

var _ AuthProvider[*http.Request] = (*HttpMockAuthProvider)(nil)

func (p *HttpMockAuthProvider) GetClientId(ctx *http.Request) <-chan AuthResult {
	ch := make(chan AuthResult)
	go func() {
		clientID := ctx.URL.Query().Get("client_id")
		ch <- AuthResult{ClientID: clientID}
		close(ch)
	}()
	return ch
}
