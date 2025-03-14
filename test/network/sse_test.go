package main

import (
	"bytes"
	"context"
	_ "embed"
	"net/http"
	"testing"
	"time"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/auth"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/message/v1"
	network_server "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/server"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/sse"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/query_filter_expr"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
	db "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/synchronizer"
	"github.com/stretchr/testify/assert"
)

//go:embed test_permission1.js
var testPermissionConditional string

//go:embed test_schema1.js
var testSchema1 string

func TestBasicSse(t *testing.T) {
	// 创建带取消的上下文，用于控制所有组件
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建通道用于同步
	serverReadyCh := make(chan struct{})
	clientReadyCh := make(chan struct{})
	eventReceivedCh := make(chan bool)

	// 启动服务器
	go func() {
		opts := &network_server.RapierDbHTTPServerOption{
			Addr:        "localhost:8080",
			SseEndpoint: "/sse",
			ApiEndpoint: "/api",
		}
		server := network_server.NewRapierDbHTTPServer(opts)
		server.SetAuthProvider(&auth.HttpMockAuthProvider{})

		// 启动服务器
		log.Info("正在启动服务器...")
		err := server.Start()
		if err != nil {
			log.Terrorf(t, "启动服务器失败: %v", err)
			return
		}

		// 等待服务器启动完成
		select {
		case <-server.WaitForStatus(network_server.ServerStatusRunning, 5*time.Second):
			log.Info("服务器已成功启动")
		case <-time.After(5 * time.Second):
			log.Terrorf(t, "服务器未能在超时时间内成功启动")
			return
		}

		// 通知测试函数服务器已就绪
		serverReadyCh <- struct{}{}

		// 接收关闭信号
		<-ctx.Done()
		log.Info("开始关闭服务器...")
		server.Stop()
		log.Info("服务器已完全关闭")
	}()

	// 启动客户端
	go func() {
		// 等待服务器准备就绪
		<-serverReadyCh

		// 创建客户端
		eventCh := make(chan *sse.SseEvent)
		client := sse.NewSseClient("http://localhost:8080/sse?client_id=test_client")

		log.Info("正在连接到 SSE 服务...")
		go func() {
			err := client.SubscribeChanWithContext(ctx, eventCh)
			if err != nil {
				log.Terrorf(t, "订阅失败: %v", err)
				return
			}
		}()

		// 等待客户端连接建立
		select {
		case <-client.WaitForStatus(sse.ClientStatusConnected, 5*time.Second):
			log.Info("SSE客户端已成功连接")
		case <-time.After(5 * time.Second):
			log.Terrorf(t, "SSE客户端未能在超时时间内成功连接")
			return
		}

		// 通知测试函数客户端已就绪
		clientReadyCh <- struct{}{}

		// 处理接收到的事件
		go func() {
			for event := range eventCh {
				log.Infof("客户端接收到事件: %s\n", event.Data)
				eventReceivedCh <- true
			}
		}()

		// 获取服务器的通道用于广播
		// 这里假设服务器已经共享了channel对象
		// 在测试中，我们可以直接使用endpoint API来代替

		// 接收关闭信号
		<-ctx.Done()
		log.Info("开始关闭客户端...")
		client.Close()
		log.Info("客户端已完全关闭")
	}()

	// 等待服务器和客户端准备就绪
	<-serverReadyCh
	<-clientReadyCh
	log.Info("服务器和客户端已准备就绪")

	// 发送消息
	log.Info("发送 hello 到客户端")
	// 使用API发送消息
	apiUrl := "http://localhost:8080/api?client_id=test_client"
	resp, err := http.Post(apiUrl, "application/octet-stream", bytes.NewReader([]byte("hello")))
	if err != nil {
		log.Terrorf(t, "发送消息失败: %v", err)
		return
	}
	defer resp.Body.Close()

	// 等待接收事件或超时
	select {
	case <-eventReceivedCh:
		log.Info("成功接收到事件")
	case <-time.After(3 * time.Second):
		log.Terrorf(t, "超时未收到事件")
	}

	log.Info("测试完成，开始清理资源")

	// 通过取消上下文来触发所有组件的优雅关闭
	cancel()

	// 给组件一些时间进行清理
	time.Sleep(1 * time.Second)

	log.Info("测试成功完成")
}

func TestBasicCURD(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动服务器
	serverReadyCh := make(chan struct{})
	go startServer(t, ctx, serverReadyCh)

	select {
	case <-serverReadyCh:
	case <-time.After(10 * time.Second):
		log.Terrorf(t, "服务器未在超时时间内成功启动")
		return
	}

	// 启动客户端
	clientReadyCh := make(chan struct{})
	go startSseClient(t, ctx, clientReadyCh)

	select {
	case <-clientReadyCh:
	case <-time.After(10 * time.Second):
		log.Terrorf(t, "SSE客户端未在超时时间内成功连接")
		return
	}

	log.Info("服务器和客户端已准备就绪")

	time.Sleep(5 * time.Second)

	// 通过取消上下文来触发所有组件的优雅关闭
	cancel()

	// 给组件一些时间进行清理
	time.Sleep(1 * time.Second)

	log.Info("测试成功完成")
}

func setupEngine(t *testing.T) *storage_engine.StorageEngine {
	dbPath := t.TempDir()
	log.Infof("dbPath: %s", dbPath)
	dbSchema, err := storage_engine.NewDatabaseSchemaFromJs(testSchema1)
	assert.NoError(t, err)
	dbPermissionsJs := testPermissionConditional
	err = storage_engine.CreateNewDatabase(dbPath, dbSchema, dbPermissionsJs)
	assert.NoError(t, err)
	opts, err := storage_engine.DefaultStorageEngineOptions(dbPath)
	assert.NoError(t, err)
	engine, err := storage_engine.OpenStorageEngine(opts)
	assert.NoError(t, err)

	user1 := loro.NewLoroDoc()
	user1Map := user1.GetMap("data")
	user1Map.InsertString("id", "user1")
	user1Map.InsertString("username", "Alice")
	user1Map.InsertString("role", "normal")

	user2 := loro.NewLoroDoc()
	user2Map := user2.GetMap("data")
	user2Map.InsertString("id", "user2")
	user2Map.InsertString("username", "Bob")
	user2Map.InsertString("role", "admin")

	user3 := loro.NewLoroDoc()
	user3Map := user3.GetMap("data")
	user3Map.InsertString("id", "user3")
	user3Map.InsertString("username", "Charlie")
	user3Map.InsertString("role", "normal")

	postMeta1 := loro.NewLoroDoc()
	postMeta1Map := postMeta1.GetMap("data")
	postMeta1Map.InsertString("id", "post1")
	postMeta1Map.InsertString("owner", "user1")
	postMeta1Map.InsertString("title", "Learn Go In 5 Minutes")

	tr := &storage_engine.Transaction{
		TxID:           "123e4567-e89b-12d3-a456-426614174000",
		TargetDatabase: "testdb",
		Committer:      "user1",
		Operations: []any{
			storage_engine.InsertOp{
				Collection: "users",
				DocID:      "user1",
				Snapshot:   user1.ExportSnapshot().Bytes(),
			},
			storage_engine.InsertOp{
				Collection: "users",
				DocID:      "user2",
				Snapshot:   user2.ExportSnapshot().Bytes(),
			},
			storage_engine.InsertOp{
				Collection: "users",
				DocID:      "user3",
				Snapshot:   user3.ExportSnapshot().Bytes(),
			},
			storage_engine.InsertOp{
				Collection: "postMetas",
				DocID:      "post1",
				Snapshot:   postMeta1.ExportSnapshot().Bytes(),
			},
		},
	}
	err = engine.Commit(tr)
	assert.NoError(t, err)
	return engine
}

func startServer(t *testing.T, ctx context.Context, serverReadyCh chan<- struct{}) {
	// 初始化临时数据库
	storageEngine := setupEngine(t)

	// 等待存储引擎完全打开
	select {
	case <-storageEngine.WaitForStatus(storage_engine.StorageEngineStatusOpen, 5*time.Second):
		log.Info("存储引擎已成功打开")
	case <-time.After(5 * time.Second):
		log.Terrorf(t, "存储引擎未能在超时时间内成功打开")
		return
	}

	opts := &network_server.RapierDbHTTPServerOption{
		Addr:        "localhost:8080",
		SseEndpoint: "/sse",
		ApiEndpoint: "/api",
	}
	server := network_server.NewRapierDbHTTPServer(opts)
	server.SetAuthProvider(&auth.HttpMockAuthProvider{})

	// 启动服务器
	log.Info("服务器启动中...")
	err := server.Start()
	if err != nil {
		log.Terrorf(t, "启动服务器失败: %v", err)
		return
	}

	// 等待服务器启动完成
	select {
	case <-server.WaitForStatus(network_server.ServerStatusRunning, 5*time.Second):
		log.Info("服务器已成功启动")
	case <-time.After(5 * time.Second):
		log.Terrorf(t, "服务器未能在超时时间内成功启动")
		return
	}

	// 启动同步器
	channel := server.GetChannel()
	synchronizerConfig := &db.SynchronizerConfig{}
	synchronizer := db.NewSynchronizer(storageEngine, channel, synchronizerConfig)
	err = synchronizer.Start()
	if err != nil {
		log.Terrorf(t, "启动同步器失败: %v", err)
		return
	}

	// 等待同步器完全启动
	select {
	case <-synchronizer.WaitForStatus(db.SynchronizerStatusRunning, 5*time.Second):
		log.Info("同步器已成功启动")
	case <-time.After(5 * time.Second):
		log.Terrorf(t, "同步器未能在超时时间内成功启动")
		return
	}

	// 通知测试函数服务器已就绪
	serverReadyCh <- struct{}{}

	// 接收到关闭信号时，优雅关闭
	<-ctx.Done()
	log.Info("开始关闭服务器...")

	synchronizer.Stop()
	server.Stop()
	storageEngine.Close()

	log.Info("服务器已完全关闭")
}

func startSseClient(t *testing.T, ctx context.Context, clientReadyCh chan<- struct{}) {
	eventCh := make(chan *sse.SseEvent)
	sseClient := sse.NewSseClient("http://localhost:8080/sse?client_id=test_client")

	// 订阅事件
	go func() {
		err := sseClient.SubscribeChan(eventCh)
		if err != nil {
			log.Terrorf(t, "订阅失败: %v", err)
			return
		}
	}()

	// 等待客户端连接建立
	select {
	case <-sseClient.WaitForStatus(sse.ClientStatusConnected, 5*time.Second):
		log.Info("SSE客户端已成功连接")
	case <-time.After(5 * time.Second):
		log.Terrorf(t, "SSE客户端未能在超时时间内成功连接")
		return
	}

	// 通知测试函数客户端已就绪
	clientReadyCh <- struct{}{}

	// 处理接收到的事件
	go func() {
		for event := range eventCh {
			buf := bytes.NewBuffer(event.Data)
			msg, err := message.DecodeMessage(buf)
			if err != nil {
				log.Terrorf(t, "解码失败: %v", err)
				return
			}
			log.Infof("客户端接收到消息: %s", msg.DebugPrint())
		}
	}()

	// 创建查询
	query1 := query.FindManyQuery{
		Collection: "users",
		Filter: &query_filter_expr.ValueExpr{
			Value: true,
		},
	}

	// 创建订阅更新消息
	apiUrl := "http://localhost:8080/api?client_id=test_client"
	msg1 := message.SubscriptionUpdateMessageV1{
		Added:   []query.Query{&query1},
		Removed: []query.Query{},
	}
	msg1Bytes, err := msg1.Encode()
	if err != nil {
		log.Terrorf(t, "编码失败: %v", err)
		return
	}

	// 发送订阅请求
	log.Info("客户端发送消息要求更新订阅")
	resp, err := http.Post(apiUrl, "application/octet-stream", bytes.NewReader(msg1Bytes))
	if err != nil {
		log.Terrorf(t, "发送订阅请求失败: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Terrorf(t, "订阅请求返回非200状态码: %d", resp.StatusCode)
		return
	}

	// 接收到关闭信号时，优雅关闭
	<-ctx.Done()
	log.Info("开始关闭SSE客户端...")

	sseClient.Close()

	log.Info("SSE客户端已完全关闭")
}
