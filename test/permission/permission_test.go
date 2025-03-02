package main

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
	"github.com/stretchr/testify/assert"
)

//go:embed test_permission1.js
var testPermissionConditional string

//go:embed test_schema1.js
var testSchema1 string

func TestPermissionFromJs(t *testing.T) {
	t.Run("test_permission_conditional", func(t *testing.T) {
		permission, err := query.NewPermissionFromJs(testPermissionConditional)
		assert.NoError(t, err)

		engine := setupEngine(t)

		post1, err := engine.LoadDoc("postMetas", "post1")
		assert.NoError(t, err)
		{
			params := query.CanViewParams{
				Collection: "postMetas",
				DocId:      "post1",
				Doc:        post1,
				ClientId:   "user1",
				Db:         &query.DbWrapper{QueryExecutor: &query.QueryExecutor{StorageEngine: engine}},
			}
			result := permission.CanView(params)
			assert.True(t, result)
		}

		{
			params := query.CanViewParams{
				Collection: "postMetas",
				DocId:      "post1",
				Doc:        post1,
				ClientId:   "userXXXX",
				Db:         &query.DbWrapper{QueryExecutor: &query.QueryExecutor{StorageEngine: engine}},
			}
			result := permission.CanView(params)
			assert.False(t, result)
		}

		cleanupEngine(t, engine)
	})
}

func setupEngine(t *testing.T) *storage_engine.StorageEngine {
	dbPath := t.TempDir()
	fmt.Println("dbPath", dbPath)
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
	user1Map.InsertString("role", "admin")

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
				Collection: "postMetas",
				DocID:      "post1",
				Snapshot:   postMeta1.ExportSnapshot().Bytes(),
			},
		},
	}
	err = engine.Commit(tr)
	assert.NoError(t, err)

	post1, err := engine.LoadDoc("postMetas", "post1")
	assert.NoError(t, err)
	fmt.Println("post1", post1)

	return engine
}

func cleanupEngine(t *testing.T, engine *storage_engine.StorageEngine) {
	assert.NoError(t, engine.Close())
}
