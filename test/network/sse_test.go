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
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
	db "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/synchronizer"
	"github.com/stretchr/testify/assert"
)

//go:embed test_permission1.js
var testPermissionConditional string

//go:embed test_schema1.js
var testSchema1 string

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
		log.Info("正在启动服务器...")
		err := server.Start()
		if err != nil && err != http.ErrServerClosed {
			log.Terrorf(t, "启动服务器失败: %v", err)
			return
		}
	}()

	// 等待服务器启动
	time.Sleep(1 * time.Second)
	log.Info("服务器已启动")

	// 创建客户端，并订阅事件
	eventCh := make(chan *sse.SseEvent)
	receivedCh := make(chan bool)
	client := sse.NewSseClient("http://localhost:8080/sse?client_id=test_client")

	log.Info("正在连接到 SSE 服务...")
	go func() {
		err := client.SubscribeChanWithContext(ctx, eventCh)
		if err != nil {
			log.Terrorf(t, "订阅失败: %v", err)
			return
		}
	}()

	// 等待连接建立
	time.Sleep(1 * time.Second)
	log.Info("SSE 连接已建立")

	// 处理接收到的事件
	go func() {
		for event := range eventCh {
			log.Infof("客户端接收到事件: %s\n", event.Data)
			receivedCh <- true
		}
	}()

	// 发送消息
	log.Info("发送 hello 到客户端")
	channel := server.GetChannel()
	channel.Broadcast([]byte("hello"))

	// 等待接收事件或超时
	select {
	case <-receivedCh:
	case <-time.After(3 * time.Second):
		log.Terrorf(t, "超时未收到事件")
	}
	time.Sleep(1 * time.Second)
	log.Info("测试结束，关闭服务器")
	server.Stop()
}

func TestBasicCURD(t *testing.T) {
	storageEngine := setupEngine(t)

	opts := &network_server.RapierDbHTTPServerOption{
		Addr:        "localhost:8080",
		SseEndpoint: "/sse",
		ApiEndpoint: "/api",
	}
	server := network_server.NewRapierDbHTTPServer(opts)
	server.SetAuthProvider(&auth.HttpMockAuthProvider{})

	// 启动服务器
	go func() {
		log.Info("服务器启动中...")
		err := server.Start()
		if err != nil && err != http.ErrServerClosed {
			log.Terrorf(t, "启动服务器失败: %v", err)
			return
		}
	}()

	// 启动同步器
	channel := server.GetChannel()
	synchronizerConfig := &db.SynchronizerConfig{}
	synchronizer := db.NewSynchronizer(storageEngine, channel, synchronizerConfig)
	synchronizer.Start()

	eventCh := make(chan *sse.SseEvent)
	sseClient := sse.NewSseClient("http://localhost:8080/sse?client_id=test_client")
	go func() {
		err := sseClient.SubscribeChan(eventCh)
		if err != nil {
			log.Terrorf(t, "订阅失败: %v", err)
			return
		}
	}()

	apiUrl := "http://localhost:8080/api?client_id=test_client"
	msg1 := message.SubscriptionUpdateMessageV1{
		Added:   []string{"users"},
		Removed: []string{},
	}
	msg1Bytes, err := msg1.Encode()
	if err != nil {
		log.Terrorf(t, "编码失败: %v", err)
		return
	}
	log.Info("客户端发送消息要求更新订阅")
	http.Post(apiUrl, "", bytes.NewReader(msg1Bytes))

	// 处理接收到的事件
	go func() {
		for event := range eventCh {
			log.Infof("客户端接收到事件: %s\n", event.Data)
		}
	}()

	time.Sleep(10 * time.Second)
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
