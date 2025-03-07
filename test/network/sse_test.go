package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/auth"
	network_server "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/server"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/sse"
)

func TestBasicSse(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := &network_server.RapierDbHTTPServerOption{
		Addr:        "localhost:8080",
		SseEndpoint: "/sse",
		ApiEndpoint: "/api",
	}
	server := network_server.NewRapierDbHTTPServer(opts)
	server.SetAuthProvider(&auth.HttpMockAuthProvider{})

	// 启动服务器
	go func() {
		fmt.Println("正在启动服务器...")
		err := server.Start()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("启动服务器失败: %v", err)
			return
		}
	}()

	// 等待服务器启动
	time.Sleep(1 * time.Second)
	fmt.Println("服务器已启动")

	// 创建客户端，并订阅事件
	eventCh := make(chan *sse.SseEvent)
	receivedCh := make(chan bool)
	client := sse.NewClient("http://localhost:8080/sse?client_id=test_client")

	fmt.Println("正在连接到 SSE 服务...")
	go func() {
		err := client.SubscribeChanWithContext(ctx, eventCh)
		if err != nil {
			t.Errorf("订阅失败: %v", err)
			return
		}
	}()

	// 等待连接建立
	time.Sleep(1 * time.Second)
	fmt.Println("SSE 连接已建立")

	// 处理接收到的事件
	go func() {
		for event := range eventCh {
			fmt.Printf("客户端接收到事件: %s\n", event.Data)
			receivedCh <- true
		}
	}()

	// 发送消息
	fmt.Println("发送 hello 到客户端")
	channel := server.GetChannel()
	channel.Broadcast([]byte("hello"))

	// 等待接收事件或超时
	select {
	case <-receivedCh:
	case <-time.After(3 * time.Second):
		t.Error("超时未收到事件")
	}

	time.Sleep(1 * time.Second)
	fmt.Println("测试结束，关闭服务器")
	server.Stop()
}
