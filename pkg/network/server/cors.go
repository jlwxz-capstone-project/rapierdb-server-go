package network_server

import (
	"net/http"
	"strconv"
)

type CorsMiddlewareBuilder struct {
	AllowOrigin      *string
	AllowMethods     *string
	AllowHeaders     *string
	AllowCredentials *bool
}

func (b *CorsMiddlewareBuilder) WithAllowOrigin(allowOrigin string) *CorsMiddlewareBuilder {
	b.AllowOrigin = &allowOrigin
	return b
}

func (b *CorsMiddlewareBuilder) WithAllowMethods(allowMethods string) *CorsMiddlewareBuilder {
	b.AllowMethods = &allowMethods
	return b
}

func (b *CorsMiddlewareBuilder) WithAllowHeaders(allowHeaders string) *CorsMiddlewareBuilder {
	b.AllowHeaders = &allowHeaders
	return b
}

func (b *CorsMiddlewareBuilder) WithAllowCredentials(allowCredentials bool) *CorsMiddlewareBuilder {
	b.AllowCredentials = &allowCredentials
	return b
}

type CorsMiddleware = func(next http.Handler) http.Handler

func (b *CorsMiddlewareBuilder) Build() CorsMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if b.AllowOrigin != nil {
				w.Header().Set("Access-Control-Allow-Origin", *b.AllowOrigin)
			}
			if b.AllowMethods != nil {
				w.Header().Set("Access-Control-Allow-Methods", *b.AllowMethods)
			}
			if b.AllowHeaders != nil {
				w.Header().Set("Access-Control-Allow-Headers", *b.AllowHeaders)
			}
			if b.AllowCredentials != nil {
				w.Header().Set("Access-Control-Allow-Credentials", strconv.FormatBool(*b.AllowCredentials))
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
