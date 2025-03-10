package auth

import (
	"net/http"
)

type AuthenticationResult struct {
	ClientID string
	Err      error
}

// Authenticator 是一个身份认证器，用于从上下文中获取客户端 ID
// 注意返回的是一个通道（异步值），这使得它可以容易地与外部
// 系统，如单独的鉴权服务进行集成
type Authenticator[CTX any] interface {
	Authenticate(ctx CTX) <-chan AuthenticationResult
}

// HttpMockAuthProvider 实现了 AuthProvider 接口，专门用于 HTTP 请求认证
type HttpMockAuthProvider struct{}

var _ Authenticator[*http.Request] = (*HttpMockAuthProvider)(nil)

func (p *HttpMockAuthProvider) Authenticate(ctx *http.Request) <-chan AuthenticationResult {
	ch := make(chan AuthenticationResult)
	go func() {
		clientID := ctx.URL.Query().Get("client_id")
		ch <- AuthenticationResult{ClientID: clientID}
		close(ch)
	}()
	return ch
}
