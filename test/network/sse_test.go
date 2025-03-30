package main

import (
	"context"
	_ "embed"
	"testing"
	"time"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/auth"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/client"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	network_client "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/client"
	network_server "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/server"
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

const serverUrl = "localhost:8080"
const sseEndpoint = "/sse"
const apiEndpoint = "/api"

func TestBasicCURD(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动服务器
	serverReadyCh := make(chan struct{})
	go startServer(t, ctx, serverReadyCh)
	<-serverReadyCh
	log.Info("服务器已启动")

	// 启动 SSE 客户端
	clientReadyCh := make(chan struct{})
	go startTestClient(t, ctx, clientReadyCh)
	<-clientReadyCh
	log.Info("客户端已启动")

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
	<-storageEngine.WaitForStatus(storage_engine.StorageEngineStatusOpen)

	opts := &network_server.RapierDbHTTPServerOption{
		Addr:        serverUrl,
		SseEndpoint: sseEndpoint,
		ApiEndpoint: apiEndpoint,
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
	<-server.WaitForStatus(network_server.ServerStatusRunning)

	// 启动同步器
	channel := server.GetChannel()
	synchronizerConfig := &db.SynchronizerConfig{}
	synchronizer := db.NewSynchronizer(storageEngine, channel, synchronizerConfig)
	err = synchronizer.Start()
	if err != nil {
		log.Terrorf(t, "启动同步器失败: %v", err)
		return
	}
	<-synchronizer.WaitForStatus(db.SynchronizerStatusRunning)

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

func startTestClient(t *testing.T, ctx context.Context, clientReadyCh chan<- struct{}) {
	cm := network_client.NewHttpChannelManager(network_client.HttpChannelConfig{
		ReceiveURL: "http://" + serverUrl + sseEndpoint,
		SendURL:    "http://" + serverUrl + apiEndpoint,
		Headers:    map[string]string{},
	})
	c := client.NewTestClient(cm)
	c.Connect()

	<-c.WaitForStatus(client.TEST_CLIENT_STATUS_READY)
	clientReadyCh <- struct{}{}

	query1 := query.FindManyQuery{
		Collection: "users",
		Filter: &query_filter_expr.ValueExpr{
			Value: true,
		},
	}
	rquery1, err := c.FindMany(&query1)
	assert.NoError(t, err)

	log.Debugf("query1 之前结果：%v", rquery1.Result.Get())
	time.Sleep(2 * time.Second)
	log.Debugf("query1 之后结果：%v", rquery1.Result.Get())
}

// func startSseClient(t *testing.T, ctx context.Context, clientReadyCh chan<- struct{}) {
// 	eventCh := make(chan *sse.SseEvent)
// 	sseClient := sse.NewSseClient("http://localhost:8080/sse?client_id=test_client")

// 	// 订阅事件
// 	go func() {
// 		err := sseClient.SubscribeChan(eventCh)
// 		if err != nil {
// 			log.Terrorf(t, "订阅失败: %v", err)
// 			return
// 		}
// 	}()

// 	// 等待客户端连接建立
// 	<-sseClient.WaitForStatus(sse.SSE_CLIENT_STATUS_CONNECTED)
// 	clientReadyCh <- struct{}{} // 通知测试函数客户端已就绪

// 	// 处理接收到的事件
// 	go func() {
// 		for event := range eventCh {
// 			buf := bytes.NewBuffer(event.Data)
// 			msg, err := message.DecodeMessage(buf)
// 			if err != nil {
// 				log.Terrorf(t, "解码失败: %v", err)
// 				return
// 			}
// 			log.Infof("客户端接收到消息: %s", msg.DebugPrint())
// 		}
// 	}()

// 	// 创建查询
// 	query1 := query.FindManyQuery{
// 		Collection: "users",
// 		Filter: &query_filter_expr.ValueExpr{
// 			Value: true,
// 		},
// 	}

// 	// 创建订阅更新消息
// 	apiUrl := "http://localhost:8080/api?client_id=test_client"
// 	msg1 := message.SubscriptionUpdateMessageV1{
// 		Added:   []query.Query{&query1},
// 		Removed: []query.Query{},
// 	}
// 	msg1Bytes, err := msg1.Encode()
// 	if err != nil {
// 		log.Terrorf(t, "编码失败: %v", err)
// 		return
// 	}

// 	// 发送订阅请求
// 	log.Info("客户端发送消息要求更新订阅")
// 	resp, err := http.Post(apiUrl, "application/octet-stream", bytes.NewReader(msg1Bytes))
// 	if err != nil {
// 		log.Terrorf(t, "发送订阅请求失败: %v", err)
// 		return
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		log.Terrorf(t, "订阅请求返回非200状态码: %d", resp.StatusCode)
// 		return
// 	}

// 	// 接收到关闭信号时，优雅关闭
// 	<-ctx.Done()
// 	log.Info("开始关闭SSE客户端...")

// 	sseClient.Close()

// 	log.Info("SSE客户端已完全关闭")
// }
