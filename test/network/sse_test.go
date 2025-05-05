package main

// import (
// 	"context"
// 	_ "embed"
// 	"net/http"
// 	"testing"
// 	"time"

// 	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/auth"
// 	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/client"
// 	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
// 	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
// 	network_client "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/client"
// 	network_server "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/server"
// 	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
// 	qfe "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/query_filter_expr"
// 	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
// 	db "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/synchronizer"
// 	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
// 	"github.com/stretchr/testify/assert"
// )

// //go:embed test_permission1.js
// var testPermissionConditional string

// //go:embed test_schema1.js
// var testSchema1 string

// func TestClientCodeStart(t *testing.T) {
// 	dbSetup := func(t *testing.T, engine *storage_engine.StorageEngine) error {
// 		user1 := loro.NewLoroDoc()
// 		user1Map := user1.GetMap("data")
// 		user1Map.InsertValueCoerce("id", "user1")
// 		user1Map.InsertValueCoerce("username", "Alice")
// 		user1Map.InsertValueCoerce("role", "normal")

// 		user2 := loro.NewLoroDoc()
// 		user2Map := user2.GetMap("data")
// 		user2Map.InsertValueCoerce("id", "user2")
// 		user2Map.InsertValueCoerce("username", "Bob")
// 		user2Map.InsertValueCoerce("role", "admin")

// 		user3 := loro.NewLoroDoc()
// 		user3Map := user3.GetMap("data")
// 		user3Map.InsertValueCoerce("id", "user3")
// 		user3Map.InsertValueCoerce("username", "Charlie")
// 		user3Map.InsertValueCoerce("role", "normal")

// 		postMeta1 := loro.NewLoroDoc()
// 		postMeta1Map := postMeta1.GetMap("data")
// 		postMeta1Map.InsertValueCoerce("id", "post1")
// 		postMeta1Map.InsertValueCoerce("owner", "user1")
// 		postMeta1Map.InsertValueCoerce("title", "Learn Go In 5 Minutes")

// 		tr := &storage_engine.Transaction{
// 			TxID:           "123e4567-e89b-12d3-a456-426614174000",
// 			TargetDatabase: "testdb",
// 			Committer:      "user1",
// 			Operations: []storage_engine.TransactionOp{
// 				&storage_engine.InsertOp{
// 					Collection: "users",
// 					DocID:      "user1",
// 					Snapshot:   user1.ExportSnapshot().Bytes(),
// 				},
// 				&storage_engine.InsertOp{
// 					Collection: "users",
// 					DocID:      "user2",
// 					Snapshot:   user2.ExportSnapshot().Bytes(),
// 				},
// 				&storage_engine.InsertOp{
// 					Collection: "users",
// 					DocID:      "user3",
// 					Snapshot:   user3.ExportSnapshot().Bytes(),
// 				},
// 				&storage_engine.InsertOp{
// 					Collection: "postMetas",
// 					DocID:      "post1",
// 					Snapshot:   postMeta1.ExportSnapshot().Bytes(),
// 				},
// 			},
// 		}
// 		err := engine.Commit(tr)
// 		return err
// 	}

// 	// Start server
// 	ctx, cancel := context.WithCancel(context.Background())
// 	serverReadyCh := make(chan struct{})
// 	testServerOpts := &testServerOptions{
// 		SchemaJs:      testSchema1,
// 		Addr:          "localhost:8080",
// 		SseEndpoint:   "/sse",
// 		ApiEndpoint:   "/api",
// 		AuthProvider:  &auth.HttpMockAuthProvider{},
// 		DbSetup:       dbSetup,
// 		T:             t,
// 		ServerReadyCh: serverReadyCh,
// 		Ctx:           ctx,
// 	}
// 	go startServer(testServerOpts)
// 	<-serverReadyCh
// 	log.Info("Server started")

// 	// Start SSE client
// 	clientReadyCh := make(chan struct{})
// 	clientAction := func(c *client.TestClient) error {
// 		query1 := query.FindManyQuery{
// 			Collection: "users",
// 			Filter:     qfe.NewValueExpr(true),
// 		}
// 		rquery1, err := c.FindMany(&query1)
// 		assert.NoError(t, err)

// 		log.Debugf("Result before query1: %v", rquery1.Result.Get())
// 		time.Sleep(2 * time.Second)
// 		log.Debugf("Result after query1: %v", rquery1.Result.Get())

// 		user4 := loro.NewLoroDoc()
// 		user4Map := user4.GetMap("data")
// 		user4Map.InsertValueCoerce("id", "user4")
// 		user4Map.InsertValueCoerce("username", "David")
// 		user4Map.InsertValueCoerce("role", "normal")

// 		c.SubmitTransaction(&storage_engine.Transaction{
// 			TxID:           "123e4567-e89b-12d3-a456-426614174001",
// 			TargetDatabase: "testdb",
// 			Committer:      "user1",
// 			Operations: []storage_engine.TransactionOp{
// 				&storage_engine.InsertOp{
// 					Collection: "users",
// 					DocID:      "user4",
// 					Snapshot:   user4.ExportSnapshot().Bytes(),
// 				},
// 			},
// 		})

// 		time.Sleep(2 * time.Second)
// 		log.Debugf("Result after transaction commit: %v", rquery1.Result.Get())

// 		return nil
// 	}
// 	testClientOpts := &testClientOptions{
// 		ServerUrl:   testServerOpts.Addr,
// 		SseEndpoint: testServerOpts.SseEndpoint,
// 		ApiEndpoint: testServerOpts.ApiEndpoint,
// 		Headers: map[string]string{
// 			"X-Client-ID": "user1",
// 		},
// 		ClientReadyCh: clientReadyCh,
// 		T:             t,
// 		Ctx:           ctx,
// 		ClientAction:  clientAction,
// 	}
// 	go startTestClient(testClientOpts)
// 	<-clientReadyCh
// 	log.Info("Client started")

// 	time.Sleep(5 * time.Second)

// 	// Trigger graceful shutdown of all components by canceling the context
// 	cancel()

// 	// Give components some time to clean up
// 	time.Sleep(1 * time.Second)

// 	log.Info("Test completed successfully")
// }

// type testServerOptions struct {
// 	SchemaJs        string                                                         // Database schema (JS string)
// 	Addr            string                                                         // Server address, e.g., "localhost:8080"
// 	SseEndpoint     string                                                         // SSE endpoint, e.g., "/sse"
// 	ApiEndpoint     string                                                         // API endpoint, e.g., "/api"
// 	AuthProvider    auth.Authenticator[*http.Request]                              // Authenticator, can use HttpMockAuthProvider for testing
// 	DbSetup         func(t *testing.T, engine *storage_engine.StorageEngine) error // Database post-processing function, used to insert test data before test starts
// 	T               *testing.T                                                     //
// 	ServerReadyCh   chan<- struct{}                                                // Send a signal to this channel when ready
// 	Ctx             context.Context                                                // Context, used for graceful shutdown
// 	ShutdownTimeout time.Duration                                                  // New field for server shutdown timeout
// }

// func setupTempStorageEngine(opts *testServerOptions) *storage_engine.StorageEngine {
// 	dbPath := opts.T.TempDir()
// 	log.Infof("dbPath: %s", dbPath)

// 	dbSchema, err := storage_engine.NewDatabaseSchemaFromJs(opts.SchemaJs)
// 	assert.NoError(opts.T, err)

// 	dbPermissionsJs := testPermissionConditional
// 	err = storage_engine.CreateNewDatabase(dbPath, dbSchema, dbPermissionsJs)
// 	assert.NoError(opts.T, err)

// 	engineOpts, err := storage_engine.DefaultStorageEngineOptions(dbPath)
// 	assert.NoError(opts.T, err)

// 	engine, err := storage_engine.OpenStorageEngine(engineOpts)
// 	assert.NoError(opts.T, err)

// 	return engine
// }

// func startServer(opts *testServerOptions) {
// 	var err error

// 	// 1. Prepare storage engine
// 	tempStorageEngine := setupTempStorageEngine(opts)                         // Initialize temporary storage engine
// 	<-tempStorageEngine.WaitForStatus(storage_engine.StorageEngineStatusOpen) // Wait for storage engine to open successfully
// 	err = opts.DbSetup(opts.T, tempStorageEngine)                             // Post-processing, e.g., insert data
// 	if err != nil {
// 		log.Terrorf(opts.T, "Database post-processing function failed: %v", err)
// 		return
// 	}

// 	// 2. Start server
// 	serverOptions := &network_server.HttpNetworkOptions{
// 		BaseUrl:         opts.Addr,
// 		ReceiveEndpoint: opts.ApiEndpoint, // API is the receive endpoint for server (client sends here)
// 		SendEndpoint:    opts.SseEndpoint, // SSE is the send endpoint for server (server sends here)
// 		Authenticator:   opts.AuthProvider,
// 		ShutdownTimeout: opts.ShutdownTimeout,
// 	}
// 	server := network_server.NewHttpNetworkWithContext(serverOptions, opts.Ctx)

// 	log.Info("Server starting...")
// 	err = server.Start()
// 	if err != nil {
// 		log.Terrorf(opts.T, "Failed to start server: %v", err)
// 		return
// 	}

// 	// Wait for server to be running using the new status subscription
// 	serverStatusCh := server.SubscribeStatusChange()
// 	cleanupServerWait := func() {
// 		server.UnsubscribeStatusChange(serverStatusCh)
// 	}
// 	<-util.WaitForStatus(server.GetStatus, network_server.NetworkRunning, serverStatusCh, cleanupServerWait, 0)

// 	// 3. Start synchronizer
// 	synchronizerConfig := &db.SynchronizerConfig{}
// 	synchronizer := db.NewSynchronizerWithContext(opts.Ctx, tempStorageEngine, server, synchronizerConfig)

// 	err = synchronizer.Start()
// 	if err != nil {
// 		log.Terrorf(opts.T, "Failed to start synchronizer: %v", err)
// 		return
// 	}
// 	<-synchronizer.WaitForStatus(db.SynchronizerStatusRunning)

// 	// 4. Notify test function that server is ready
// 	opts.ServerReadyCh <- struct{}{}

// 	// 5. Graceful shutdown when shutdown signal is received
// 	<-opts.Ctx.Done()
// 	log.Info("Shutting down server...")

// 	// we don't need to stop these components manually, because we use
// 	// context to control the shutdown of all components

// 	log.Info("Server fully shut down")
// }

// type testClientOptions struct {
// 	ServerUrl     string
// 	SseEndpoint   string
// 	ApiEndpoint   string
// 	Headers       map[string]string
// 	ClientReadyCh chan<- struct{}
// 	T             *testing.T
// 	Ctx           context.Context
// 	ClientAction  func(c *client.TestClient) error
// }

// func startTestClient(opts *testClientOptions) {
// 	// Create the new HttpNetwork client provider
// 	networkOptions := &network_client.HttpNetworkOptions{
// 		BackendUrl:      "http://" + opts.ServerUrl, // Base URL of the server
// 		ReceiveEndpoint: opts.SseEndpoint,           // SSE endpoint for receiving messages from server
// 		SendEndpoint:    opts.ApiEndpoint,           // API endpoint for sending messages to server
// 		Headers:         opts.Headers,
// 	}
// 	networkProvider := network_client.NewHttpNetworkWithContext(networkOptions, opts.Ctx)

// 	// Create the test client with the new network provider
// 	c := client.NewTestClient(networkProvider)

// 	// Connect the client
// 	err := c.Connect()
// 	if err != nil {
// 		log.Terrorf(opts.T, "Client connection failed: %v", err)
// 		return
// 	}

// 	// Wait for the client to be ready
// 	<-c.WaitForStatus(client.TEST_CLIENT_STATUS_READY)
// 	opts.ClientReadyCh <- struct{}{}

// 	// Execute the client action
// 	err = opts.ClientAction(c)
// 	if err != nil {
// 		log.Terrorf(opts.T, "Client action failed: %v", err)
// 		return
// 	}

// 	// Wait for context cancellation to close the client
// 	<-opts.Ctx.Done()
// 	log.Info("Shutting down client...")
// 	c.Close()
// 	log.Info("Client shut down")
// }
