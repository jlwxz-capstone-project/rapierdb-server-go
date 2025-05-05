package main

import (
	"context"
	_ "embed"
	"fmt"
	"testing"
	"time"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/db_conn"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/db_connector"
	network_server "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/server"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/synchronizer2"
	"github.com/stretchr/testify/assert"
)

//go:embed test_schema1.js
var testSchema1 string

//go:embed test_permission1.js
var testPermission1 string

const WAIT_TIMEOUT = 30 * time.Second

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

	<-time.After(WAIT_TIMEOUT) // run for WAIT_TIMEOUT

	cancel()

	<-synchronizer.WaitForStatus(synchronizer2.SynchronizerStatusStopped)
}

func setupDb(t *testing.T) string {
	dbPath := t.TempDir()
	fmt.Println("dbPath", dbPath)
	dbSchema, err := db_conn.NewDatabaseSchemaFromJs(testSchema1)
	assert.NoError(t, err)
	dbPermissionsJs := testPermission1
	err = db_conn.CreateNewPebbleDb(dbPath, dbSchema, dbPermissionsJs)
	assert.NoError(t, err)

	return dbPath
}
