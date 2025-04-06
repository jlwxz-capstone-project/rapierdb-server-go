package main

import (
	"context"
	_ "embed"
	"net/http"
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

func TestClientCodeStart(t *testing.T) {
	dbSetup := func(t *testing.T, engine *storage_engine.StorageEngine) error {
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
			Operations: []storage_engine.TransactionOp{
				&storage_engine.InsertOp{
					Collection: "users",
					DocID:      "user1",
					Snapshot:   user1.ExportSnapshot().Bytes(),
				},
				&storage_engine.InsertOp{
					Collection: "users",
					DocID:      "user2",
					Snapshot:   user2.ExportSnapshot().Bytes(),
				},
				&storage_engine.InsertOp{
					Collection: "users",
					DocID:      "user3",
					Snapshot:   user3.ExportSnapshot().Bytes(),
				},
				&storage_engine.InsertOp{
					Collection: "postMetas",
					DocID:      "post1",
					Snapshot:   postMeta1.ExportSnapshot().Bytes(),
				},
			},
		}
		err := engine.Commit(tr)
		return err
	}

	// 启动服务器
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serverReadyCh := make(chan struct{})
	testServerOpts := &testServerOptions{
		SchemaJs:      testSchema1,
		Addr:          "localhost:8080",
		SseEndpoint:   "/sse",
		ApiEndpoint:   "/api",
		AuthProvider:  &auth.HttpMockAuthProvider{},
		DbSetup:       dbSetup,
		T:             t,
		ServerReadyCh: serverReadyCh,
		Ctx:           ctx,
	}
	go startServer(testServerOpts)
	<-serverReadyCh
	log.Info("服务器已启动")

	// 启动 SSE 客户端
	clientReadyCh := make(chan struct{})
	clientAction := func(c *client.TestClient) error {
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

		user4 := loro.NewLoroDoc()
		user4Map := user4.GetMap("data")
		user4Map.InsertString("id", "user4")
		user4Map.InsertString("username", "David")
		user4Map.InsertString("role", "normal")

		c.SubmitTransaction(&storage_engine.Transaction{
			TxID:           "123e4567-e89b-12d3-a456-426614174001",
			TargetDatabase: "testdb",
			Committer:      "user1",
			Operations: []storage_engine.TransactionOp{
				&storage_engine.InsertOp{
					Collection: "users",
					DocID:      "user4",
					Snapshot:   user4.ExportSnapshot().Bytes(),
				},
			},
		})

		time.Sleep(2 * time.Second)
		log.Debugf("事务提交之后结果：%v", rquery1.Result.Get())

		return nil
	}
	testClientOpts := &testClientOptions{
		ServerUrl:   testServerOpts.Addr,
		SseEndpoint: testServerOpts.SseEndpoint,
		ApiEndpoint: testServerOpts.ApiEndpoint,
		Headers: map[string]string{
			"X-Client-ID": "user1",
		},
		ClientReadyCh: clientReadyCh,
		T:             t,
		Ctx:           ctx,
		ClientAction:  clientAction,
	}
	go startTestClient(testClientOpts)
	<-clientReadyCh
	log.Info("客户端已启动")

	time.Sleep(5 * time.Second)

	// 通过取消上下文来触发所有组件的优雅关闭
	cancel()

	// 给组件一些时间进行清理
	time.Sleep(1 * time.Second)

	log.Info("测试成功完成")
}

type testServerOptions struct {
	SchemaJs      string                                                         // 数据库的 schema（Js 字符串）
	Addr          string                                                         // 服务器地址，如: "localhost:8080"
	SseEndpoint   string                                                         // SSE 端点，如: "/sse"
	ApiEndpoint   string                                                         // API 端点，如: "/api"
	AuthProvider  auth.Authenticator[*http.Request]                              // 认证器，测试时可以使用 HttpMockAuthProvider
	DbSetup       func(t *testing.T, engine *storage_engine.StorageEngine) error // 数据库后处理函数，用于在测试开始前插入测试数据
	T             *testing.T                                                     //
	ServerReadyCh chan<- struct{}                                                // 就绪后向这个通道发送一个信号
	Ctx           context.Context                                                // 上下文，用于优雅关闭
}

func setupTempStorageEngine(opts *testServerOptions) *storage_engine.StorageEngine {
	dbPath := opts.T.TempDir()
	log.Infof("dbPath: %s", dbPath)

	dbSchema, err := storage_engine.NewDatabaseSchemaFromJs(opts.SchemaJs)
	assert.NoError(opts.T, err)

	dbPermissionsJs := testPermissionConditional
	err = storage_engine.CreateNewDatabase(dbPath, dbSchema, dbPermissionsJs)
	assert.NoError(opts.T, err)

	engineOpts, err := storage_engine.DefaultStorageEngineOptions(dbPath)
	assert.NoError(opts.T, err)

	engine, err := storage_engine.OpenStorageEngine(engineOpts)
	assert.NoError(opts.T, err)

	return engine
}

func startServer(opts *testServerOptions) {
	var err error

	// 1. 准备存储引擎
	tempStorageEngine := setupTempStorageEngine(opts)                         // 初始化临时存储引擎
	<-tempStorageEngine.WaitForStatus(storage_engine.StorageEngineStatusOpen) // 等待存储引擎打开成功
	err = opts.DbSetup(opts.T, tempStorageEngine)                             // 后处理，比如插入数据
	if err != nil {
		log.Terrorf(opts.T, "数据库后处理函数失败: %v", err)
		return
	}

	// 2. 启动服务器
	server := network_server.NewRapierDbHTTPServer(&network_server.RapierDbHTTPServerOption{
		Addr:        opts.Addr,
		SseEndpoint: opts.SseEndpoint,
		ApiEndpoint: opts.ApiEndpoint,
	})
	server.SetAuthProvider(opts.AuthProvider) // 设置认证器

	log.Info("服务器启动中...")
	err = server.Start()
	if err != nil {
		log.Terrorf(opts.T, "启动服务器失败: %v", err)
		return
	}
	<-server.WaitForStatus(network_server.ServerStatusRunning)

	// 3. 启动同步器
	channel := server.GetChannel()
	synchronizerConfig := &db.SynchronizerConfig{}
	synchronizer := db.NewSynchronizer(tempStorageEngine, channel, synchronizerConfig)
	err = synchronizer.Start()
	if err != nil {
		log.Terrorf(opts.T, "启动同步器失败: %v", err)
		return
	}
	<-synchronizer.WaitForStatus(db.SynchronizerStatusRunning)

	// 4. 通知测试函数服务器已就绪
	opts.ServerReadyCh <- struct{}{}

	// 5. 接收到关闭信号时，优雅关闭
	<-opts.Ctx.Done()
	log.Info("开始关闭服务器...")

	synchronizer.Stop()
	server.Stop()
	tempStorageEngine.Close()

	log.Info("服务器已完全关闭")
}

type testClientOptions struct {
	ServerUrl     string
	SseEndpoint   string
	ApiEndpoint   string
	Headers       map[string]string
	ClientReadyCh chan<- struct{}
	T             *testing.T
	Ctx           context.Context
	ClientAction  func(c *client.TestClient) error
}

func startTestClient(opts *testClientOptions) {
	cm := network_client.NewHttpChannelManager(network_client.HttpChannelConfig{
		ReceiveURL: "http://" + opts.ServerUrl + opts.SseEndpoint,
		SendURL:    "http://" + opts.ServerUrl + opts.ApiEndpoint,
		Headers:    opts.Headers,
	})
	c := client.NewTestClient(cm)
	c.Connect()

	<-c.WaitForStatus(client.TEST_CLIENT_STATUS_READY)
	opts.ClientReadyCh <- struct{}{}

	err := opts.ClientAction(c)
	if err != nil {
		log.Terrorf(opts.T, "客户端操作失败: %v", err)
		return
	}
}
