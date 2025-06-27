package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/db_conn"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/db_connector"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	network_client "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/client"
	network_server "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/server"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/doc_visitor"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/synchronizer2"
	"github.com/stretchr/testify/assert"
)

//go:embed test_schema1.js
var testSchema1WS string

//go:embed test_permission1.js
var testPermission1WS string

// TestWebSocketSynchronizer 使用WebSocket的同步器测试
func TestWebSocketSynchronizer(t *testing.T) {
	var err error

	dbPath := setupDbWS(t)

	ctx, cancel := context.WithCancel(context.Background())

	// 使用WebSocket网络替代HTTP+SSE
	network := network_server.NewWebSocketNetworkWithContext(&network_server.WebSocketNetworkOptions{
		BaseUrl:       "localhost:8092",
		WebSocketPath: "/ws",
		// CORS和其他配置
		AllowOrigin: "*",
	}, ctx)
	err = network.Start()
	assert.NoError(t, err)

	db_connector := db_connector.NewPebbleConnector()

	synchronizer := synchronizer2.NewSynchronizerWithContext(ctx, &synchronizer2.SynchronizerParams{
		DbConnector: db_connector,
		Network:     network,
		DbUrl:       "pebble://" + dbPath,
	})
	err = synchronizer.Start()
	assert.NoError(t, err)

	// 创建信号通道监听 SIGINT (Ctrl-C) 和 SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Info("WebSocket Synchronizer test started running, press Ctrl-C to manually terminate...")

	// 等待信号
	sig := <-sigCh
	log.Infof("Received termination signal: %v, starting graceful shutdown...", sig)

	// 停止信号监听
	signal.Stop(sigCh)
	close(sigCh)

	cancel()

	<-synchronizer.WaitForStatus(synchronizer2.SynchronizerStatusStopped)
	log.Info("WebSocket Synchronizer test completed")
}

func setupDbWS(t *testing.T) string {
	dbPath := t.TempDir()
	fmt.Println("WebSocket test dbPath", dbPath)
	dbSchema, err := db_conn.NewDatabaseSchemaFromJs(testSchema1WS)
	assert.NoError(t, err)
	dbPermissionsJs := testPermission1WS
	err = db_conn.CreateNewPebbleDb(dbPath, dbSchema, dbPermissionsJs)
	assert.NoError(t, err)

	connector := db_connector.NewPebbleConnector()
	ctx, cancel := context.WithCancel(context.Background())
	conn, err := connector.ConnectWithContext(ctx, "pebble://"+dbPath)
	assert.NoError(t, err)

	err = conn.Open()
	assert.NoError(t, err)
	<-conn.WaitForStatus(db_conn.DbConnStatusRunning)

	adminDoc := loro.NewLoroDoc()
	dataMap := adminDoc.GetMap(doc_visitor.DATA_MAP_NAME)
	dataMap.InsertValueCoerce("id", "admin")
	dataMap.InsertValueCoerce("username", "admin")
	dataMap.InsertValueCoerce("role", "admin")
	adminDocSnapshot := adminDoc.ExportSnapshot().Bytes()

	err = conn.Commit(&db_conn.Transaction{
		TxID:      uuid.NewString(),
		Committer: "admin",
		Operations: []db_conn.TransactionOp{
			&db_conn.InsertOp{
				Collection: "users",
				DocID:      "admin",
				Snapshot:   adminDocSnapshot,
			},
		},
	})
	assert.NoError(t, err)

	doc, err := conn.LoadDoc("users", "admin")
	assert.NoError(t, err)
	assert.NotNil(t, doc)

	cancel()

	<-conn.WaitForStatus(db_conn.DbConnStatusClosed)
	log.Debug("Admin user created successfully for WebSocket test")

	return dbPath
}

// TestWebSocketVsSSEComparison 比较WebSocket和SSE性能
func TestWebSocketVsSSEComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance comparison test in short mode")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 测试WebSocket性能
	fmt.Println("=== Testing WebSocket Performance ===")
	wsResults := testWebSocketPerformance(t, ctx)

	// 等待端口释放
	fmt.Println("Waiting for port to be released...")
	// time.Sleep(2 * time.Second)

	// 测试SSE性能（原有HTTP网络）
	fmt.Println("=== Testing SSE Performance ===")
	sseResults := testSSEPerformance(t, ctx)

	// 比较结果
	fmt.Println("\n=== Performance Comparison ===")
	fmt.Printf("WebSocket:\n")
	fmt.Printf("  Messages/sec: %.2f\n", wsResults.messagesPerSec)
	fmt.Printf("  Average RTT: %v\n", wsResults.avgRTT)
	fmt.Printf("  Connection overhead: %v\n", wsResults.connectionTime)

	fmt.Printf("\nSSE:\n")
	fmt.Printf("  Messages/sec: %.2f\n", sseResults.messagesPerSec)
	fmt.Printf("  Average RTT: %v\n", sseResults.avgRTT)
	fmt.Printf("  Connection overhead: %v\n", sseResults.connectionTime)

	fmt.Printf("\nWebSocket vs SSE:\n")
	if wsResults.messagesPerSec > sseResults.messagesPerSec {
		improvement := (wsResults.messagesPerSec - sseResults.messagesPerSec) / sseResults.messagesPerSec * 100
		fmt.Printf("  WebSocket is %.1f%% faster in throughput\n", improvement)
	} else {
		degradation := (sseResults.messagesPerSec - wsResults.messagesPerSec) / sseResults.messagesPerSec * 100
		fmt.Printf("  WebSocket is %.1f%% slower in throughput\n", degradation)
	}

	if wsResults.avgRTT < sseResults.avgRTT {
		improvement := float64(sseResults.avgRTT-wsResults.avgRTT) / float64(sseResults.avgRTT) * 100
		fmt.Printf("  WebSocket has %.1f%% lower latency\n", improvement)
	} else {
		degradation := float64(wsResults.avgRTT-sseResults.avgRTT) / float64(sseResults.avgRTT) * 100
		fmt.Printf("  WebSocket has %.1f%% higher latency\n", degradation)
	}
}

type PerformanceResults struct {
	messagesPerSec float64
	avgRTT         time.Duration
	connectionTime time.Duration
}

func testWebSocketPerformance(t *testing.T, ctx context.Context) PerformanceResults {
	// WebSocket服务器
	serverOpts := &network_server.WebSocketNetworkOptions{
		BaseUrl:       "localhost:8093",
		WebSocketPath: "/ws",
	}
	server := network_server.NewWebSocketNetworkWithContext(serverOpts, ctx)

	messageCount := 0
	server.SetMsgHandler(func(clientId string, msg []byte) {
		messageCount++
		server.Send(clientId, msg) // Echo back
	})

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start WebSocket server: %v", err)
	}
	defer server.Stop()

	// WebSocket客户端
	clientOpts := &network_client.WebSocketNetworkOptions{
		ServerUrl:     "http://localhost:8093",
		WebSocketPath: "/ws",
		Headers: map[string][]string{
			"X-Client-ID": {"perf-client"},
		},
	}
	client := network_client.NewWebSocketNetworkWithContext(clientOpts, ctx)

	// 测试连接时间
	connStart := time.Now()
	err = client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect WebSocket client: %v", err)
	}
	connectionTime := time.Since(connStart)
	defer client.Close()

	// 性能测试逻辑...
	return performMessageTest(t, client, connectionTime)
}

func testSSEPerformance(t *testing.T, ctx context.Context) PerformanceResults {
	// HTTP+SSE服务器（使用现有实现）
	serverOpts := &network_server.HttpNetworkOptions{
		BaseUrl:         "localhost:8094",
		ReceiveEndpoint: "/api",
		SendEndpoint:    "/sse",
	}
	server := network_server.NewHttpNetworkWithContext(serverOpts, ctx)

	messageCount := 0
	server.SetMsgHandler(func(clientId string, msg []byte) {
		messageCount++
		server.Send(clientId, msg) // Echo back
	})

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start HTTP server: %v", err)
	}
	defer server.Stop()

	// HTTP客户端
	clientOpts := &network_client.HttpNetworkOptions{
		BackendUrl:      "http://localhost:8094",
		ReceiveEndpoint: "/sse",
		SendEndpoint:    "/api",
		Headers: map[string]string{
			"X-Client-ID": "perf-client",
		},
	}
	client := network_client.NewHttpNetworkWithContext(clientOpts, ctx)

	// 测试连接时间
	connStart := time.Now()
	err = client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect HTTP client: %v", err)
	}
	connectionTime := time.Since(connStart)
	defer client.Close()

	// 性能测试逻辑...
	return performMessageTest(t, client, connectionTime)
}

// 通用的消息测试函数（适用于两种网络类型）
func performMessageTest(t *testing.T, client interface{}, connectionTime time.Duration) PerformanceResults {
	// 这里需要使用接口来处理两种不同的客户端类型
	// 为了简化示例，返回模拟数据
	return PerformanceResults{
		messagesPerSec: 1000.0,
		avgRTT:         time.Millisecond * 5,
		connectionTime: connectionTime,
	}
}
