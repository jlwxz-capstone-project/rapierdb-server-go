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
var testSchema1 string

//go:embed test_permission1.js
var testPermission1 string

func TestSynchronizer(t *testing.T) {
	var err error

	dbPath := setupDb(t)

	ctx, cancel := context.WithCancel(context.Background())

	network := network_server.NewHttpNetworkWithContext(&network_server.HttpNetworkOptions{
		BaseUrl:         "localhost:8080",
		ReceiveEndpoint: "/api",
		SendEndpoint:    "/sse",
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

	// Create signal channel to listen for SIGINT (Ctrl-C) and SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Test started running, press Ctrl-C to manually terminate...")

	// Wait for signal
	sig := <-sigCh
	log.Infof("Received termination signal: %v, starting graceful shutdown...", sig)

	// Stop signal listening
	signal.Stop(sigCh)
	close(sigCh)

	cancel()

	<-synchronizer.WaitForStatus(synchronizer2.SynchronizerStatusStopped)
	log.Info("Test completed")
}

func setupDb(t *testing.T) string {
	dbPath := t.TempDir()
	fmt.Println("dbPath", dbPath)
	dbSchema, err := db_conn.NewDatabaseSchemaFromJs(testSchema1)
	assert.NoError(t, err)
	dbPermissionsJs := testPermission1
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
	log.Debug("Admin user created successfully")

	return dbPath
}
