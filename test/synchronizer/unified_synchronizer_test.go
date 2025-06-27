package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/google/uuid"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/db_conn"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/db_connector"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	network_server "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/server"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/doc_visitor"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/synchronizer2"
	"github.com/stretchr/testify/assert"
)

//go:embed test_schema1.js
var testSchemaUnified string

//go:embed test_permission1.js
var testPermissionUnified string

// TestUnifiedSynchronizerWebSocket 使用统一工厂创建WebSocket同步器
func TestUnifiedSynchronizerWebSocket(t *testing.T) {
	testUnifiedSynchronizer(t, network_server.NetworkTypeWebSocket)
}

// TestUnifiedSynchronizerHTTP 使用统一工厂创建HTTP+SSE同步器
func TestUnifiedSynchronizerHTTP(t *testing.T) {
	testUnifiedSynchronizer(t, network_server.NetworkTypeHTTP)
}

func testUnifiedSynchronizer(t *testing.T, networkType network_server.NetworkType) {
	var err error

	dbPath := setupDbUnified(t)

	ctx, cancel := context.WithCancel(context.Background())

	// 使用统一网络工厂创建网络提供者
	var networkOptions *network_server.UnifiedNetworkOptions

	switch networkType {
	case network_server.NetworkTypeWebSocket:
		networkOptions = network_server.DefaultWebSocketOptions("localhost:8095")
		log.Info("Testing with WebSocket network")
	case network_server.NetworkTypeHTTP:
		networkOptions = network_server.DefaultHttpOptions("localhost:8096")
		log.Info("Testing with HTTP+SSE network")
	}

	network := network_server.CreateNetworkProvider(networkOptions, ctx)
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

	log.Infof("Unified Synchronizer (%s) test started running, press Ctrl-C to manually terminate...", networkType)

	// 等待信号
	sig := <-sigCh
	log.Infof("Received termination signal: %v, starting graceful shutdown...", sig)

	// 停止信号监听
	signal.Stop(sigCh)
	close(sigCh)

	cancel()

	<-synchronizer.WaitForStatus(synchronizer2.SynchronizerStatusStopped)
	log.Infof("Unified Synchronizer (%s) test completed", networkType)
}

func setupDbUnified(t *testing.T) string {
	dbPath := t.TempDir()
	fmt.Println("Unified test dbPath", dbPath)
	dbSchema, err := db_conn.NewDatabaseSchemaFromJs(testSchemaUnified)
	assert.NoError(t, err)
	dbPermissionsJs := testPermissionUnified
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
	log.Debug("Admin user created successfully for unified test")

	return dbPath
}
