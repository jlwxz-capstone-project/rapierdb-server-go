package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	network_client "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/client"
	network_server "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/server"
)

func TestWebSocketServerNetwork(t *testing.T) {
	opts := &network_server.WebSocketNetworkOptions{
		BaseUrl:       "localhost:8089",
		WebSocketPath: "/ws",
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := network_server.NewWebSocketNetworkWithContext(opts, ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	server.SetMsgHandler(func(clientId string, msg []byte) {
		fmt.Printf("WebSocket server received from %s: %s\n", clientId, string(msg))
		// 回显消息
		server.Send(clientId, []byte("Echo: "+string(msg)))
	})

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start WebSocket server: %v", err)
	}
	fmt.Println("WebSocket server started on localhost:8089/ws")

	<-sigChan

	err = server.Stop()
	if err != nil {
		t.Fatalf("Failed to stop WebSocket server: %v", err)
	}
}

func TestWebSocketClientServer(t *testing.T) {
	// 启动WebSocket服务器
	serverOpts := &network_server.WebSocketNetworkOptions{
		BaseUrl:       "localhost:8090",
		WebSocketPath: "/ws",
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := network_server.NewWebSocketNetworkWithContext(serverOpts, ctx)

	messageCount := 0
	server.SetMsgHandler(func(clientId string, msg []byte) {
		messageCount++
		fmt.Printf("Server received message %d from %s: %s\n", messageCount, clientId, string(msg))
		// 广播消息给所有客户端
		server.Broadcast([]byte(fmt.Sprintf("Broadcast %d: %s", messageCount, string(msg))))
	})

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start WebSocket server: %v", err)
	}
	fmt.Println("WebSocket server started for client-server test")

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 创建WebSocket客户端
	clientOpts := &network_client.WebSocketNetworkOptions{
		ServerUrl:     "http://localhost:8090",
		WebSocketPath: "/ws",
		Headers: map[string][]string{
			"X-Client-ID": {"test-client"},
		},
	}
	client := network_client.NewWebSocketNetworkWithContext(clientOpts, ctx)

	clientMsgCount := 0
	client.SetMsgHandler(func(msg []byte) {
		clientMsgCount++
		fmt.Printf("Client received message %d: %s\n", clientMsgCount, string(msg))
	})

	// 连接客户端
	err = client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect WebSocket client: %v", err)
	}
	fmt.Println("WebSocket client connected")

	// 发送测试消息
	for i := 1; i <= 3; i++ {
		msg := fmt.Sprintf("Test message %d", i)
		err = client.Send([]byte(msg))
		if err != nil {
			t.Errorf("Failed to send message %d: %v", i, err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	// 等待消息处理
	time.Sleep(2 * time.Second)

	// 关闭客户端
	client.Close()

	// 关闭服务器
	cancel()
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("Test completed. Server processed %d messages, client received %d messages\n",
		messageCount, clientMsgCount)
}

func TestWebSocketPerformance(t *testing.T) {
	// 性能测试：比较WebSocket和HTTP+SSE的性能
	serverOpts := &network_server.WebSocketNetworkOptions{
		BaseUrl:       "localhost:8091",
		WebSocketPath: "/ws",
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := network_server.NewWebSocketNetworkWithContext(serverOpts, ctx)

	receivedCount := 0
	server.SetMsgHandler(func(clientId string, msg []byte) {
		receivedCount++
		// 立即回复，测试往返时间
		server.Send(clientId, msg)
	})

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start WebSocket server: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// 创建客户端
	clientOpts := &network_client.WebSocketNetworkOptions{
		ServerUrl:     "http://localhost:8091",
		WebSocketPath: "/ws",
		Headers: map[string][]string{
			"X-Client-ID": {"perf-test-client"},
		},
	}
	client := network_client.NewWebSocketNetworkWithContext(clientOpts, ctx)

	responseCount := 0
	responseTimes := make([]time.Duration, 0)
	startTimes := make(map[string]time.Time)

	client.SetMsgHandler(func(msg []byte) {
		responseCount++
		msgStr := string(msg)
		if startTime, exists := startTimes[msgStr]; exists {
			rtt := time.Since(startTime)
			responseTimes = append(responseTimes, rtt)
			delete(startTimes, msgStr)
		}
	})

	err = client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect WebSocket client: %v", err)
	}

	// 发送100条消息测试性能
	messageCount := 100
	fmt.Printf("Starting performance test with %d messages...\n", messageCount)

	startTest := time.Now()
	for i := 0; i < messageCount; i++ {
		msg := fmt.Sprintf("perf-test-%d", i)
		startTimes[msg] = time.Now()
		err = client.Send([]byte(msg))
		if err != nil {
			t.Errorf("Failed to send message %d: %v", i, err)
		}
	}

	// 等待所有响应
	timeout := time.After(10 * time.Second)
	for responseCount < messageCount {
		select {
		case <-timeout:
			t.Errorf("Timeout waiting for responses. Received %d/%d", responseCount, messageCount)
			break
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	testDuration := time.Since(startTest)

	// 计算统计数据
	if len(responseTimes) > 0 {
		var totalRTT time.Duration
		var maxRTT time.Duration
		var minRTT time.Duration = responseTimes[0]

		for _, rtt := range responseTimes {
			totalRTT += rtt
			if rtt > maxRTT {
				maxRTT = rtt
			}
			if rtt < minRTT {
				minRTT = rtt
			}
		}

		avgRTT := totalRTT / time.Duration(len(responseTimes))

		fmt.Printf("WebSocket Performance Results:\n")
		fmt.Printf("  Total messages: %d\n", messageCount)
		fmt.Printf("  Successful responses: %d\n", responseCount)
		fmt.Printf("  Test duration: %v\n", testDuration)
		fmt.Printf("  Messages per second: %.2f\n", float64(messageCount)/testDuration.Seconds())
		fmt.Printf("  Average RTT: %v\n", avgRTT)
		fmt.Printf("  Min RTT: %v\n", minRTT)
		fmt.Printf("  Max RTT: %v\n", maxRTT)
	}

	client.Close()
	cancel()
}
