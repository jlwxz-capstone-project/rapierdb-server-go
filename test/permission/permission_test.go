package main

import (
	"context"
	_ "embed"
	"fmt"
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/db_conn"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/permission_proxy"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/doc_visitor"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query_executor"
	"github.com/stretchr/testify/assert"
)

//go:embed test_permission1.js
var testPermissionConditional string

//go:embed test_schema1.js
var testSchema1 string

func TestPermission(t *testing.T) {
	engine := setupConn(t)
	defer cleanupEngine(t, engine)

	canViewTests := []struct {
		name     string
		clientId string
		want     bool
	}{
		{
			name:     "文档所有者有权限",
			clientId: "user1",
			want:     true,
		},
		{
			name:     "管理员有权限",
			clientId: "user2",
			want:     true,
		},
		{
			name:     "不是文档所有者，也不是 admin，没有权限",
			clientId: "user3",
			want:     false,
		},
		{
			name:     "不存在的用户没有权限",
			clientId: "userXXXX",
			want:     false,
		},
	}

	canCreateTests := []struct {
		name     string
		clientId string
		want     bool
	}{
		{
			name:     "任何用户都有权限创建 1",
			clientId: "user1",
			want:     true,
		},
		{
			name:     "任何用户都有权限创建 2",
			clientId: "user2",
			want:     true,
		},
		{
			name:     "任何用户都有权限创建 3",
			clientId: "user3",
			want:     true,
		},
	}

	post1, err := engine.LoadDoc("postMetas", "post1")
	assert.NoError(t, err)
	post1new := loro.NewLoroDoc()
	post1new.Import(post1.ExportSnapshot().Bytes())
	datamap := post1new.GetMap(doc_visitor.DATA_MAP_NAME)
	datamap.InsertValueCoerce("owner", "user2")

	post1new2 := loro.NewLoroDoc()
	post1new2.Import(post1.ExportSnapshot().Bytes())
	datamap2 := post1new2.GetMap(doc_visitor.DATA_MAP_NAME)
	datamap2.InsertValueCoerce("title", "Another Post")

	canUpdateTests := []struct {
		name     string
		clientId string
		newDoc   *loro.LoroDoc
		want     bool
	}{
		{
			name:     "不允许修改文档所有者，即使是文档所有者 / 管理员",
			clientId: "user1",
			newDoc:   post1new,
			want:     false,
		},
		{
			name:     "允许文档所有者修改文档标题",
			clientId: "user1",
			newDoc:   post1new2,
			want:     true,
		},
		{
			name:     "允许管理员修改文档标题",
			clientId: "user2",
			newDoc:   post1new2,
			want:     true,
		},
	}

	permission, err := permission_proxy.NewPermissionFromJs(testPermissionConditional)
	assert.NoError(t, err)

	t.Run("测试 canView 权限", func(t *testing.T) {
		for _, tt := range canViewTests {
			t.Run(tt.name, func(t *testing.T) {
				params := permission_proxy.CanViewParams{
					Collection: "postMetas",
					DocId:      "post1",
					Doc:        post1,
					ClientId:   tt.clientId,
					Db: &permission_proxy.DbWrapper{
						QueryExecutor: query_executor.NewQueryExecutor(engine),
					},
				}
				result := permission.CanView(params)
				assert.Equal(t, tt.want, result)
			})
		}
	})

	t.Run("测试 canCreate 权限", func(t *testing.T) {
		for _, tt := range canCreateTests {
			t.Run(tt.name, func(t *testing.T) {
				params := permission_proxy.CanCreateParams{
					Collection: "postMetas",
					DocId:      "post1",
					NewDoc:     post1,
					ClientId:   tt.clientId,
					Db: &permission_proxy.DbWrapper{
						QueryExecutor: query_executor.NewQueryExecutor(engine),
					},
				}
				result := permission.CanCreate(params)
				assert.Equal(t, tt.want, result)
			})
		}
	})

	t.Run("测试 canUpdate 权限", func(t *testing.T) {
		for _, tt := range canUpdateTests {
			t.Run(tt.name, func(t *testing.T) {
				params := permission_proxy.CanUpdateParams{
					Collection: "postMetas",
					DocId:      "post1",
					NewDoc:     tt.newDoc,
					OldDoc:     post1,
					ClientId:   tt.clientId,
					Db: &permission_proxy.DbWrapper{
						QueryExecutor: query_executor.NewQueryExecutor(engine),
					},
				}
				result := permission.CanUpdate(params)
				assert.Equal(t, tt.want, result)
			})
		}
	})
}

func setupConn(t *testing.T) db_conn.DbConnection {
	dbPath := t.TempDir()
	fmt.Println("dbPath", dbPath)
	dbSchema, err := db_conn.NewDatabaseSchemaFromJs(testSchema1)
	assert.NoError(t, err)
	dbPermissionsJs := testPermissionConditional
	err = db_conn.CreateNewPebbleDb(dbPath, dbSchema, dbPermissionsJs)
	assert.NoError(t, err)

	opts := db_conn.PebbleDbConnParams{
		Path: dbPath,
	}
	opts.EnsureDefaults()
	ctx := context.Background()
	conn, err := db_conn.NewPebbleDbConnWithContext(ctx, &opts)
	assert.NoError(t, err)
	err = conn.Open()
	assert.NoError(t, err)

	user1 := loro.NewLoroDoc()
	user1Map := user1.GetMap(doc_visitor.DATA_MAP_NAME)
	user1Map.InsertValueCoerce("id", "user1")
	user1Map.InsertValueCoerce("username", "Alice")
	user1Map.InsertValueCoerce("role", "normal")

	user2 := loro.NewLoroDoc()
	user2Map := user2.GetMap(doc_visitor.DATA_MAP_NAME)
	user2Map.InsertValueCoerce("id", "user2")
	user2Map.InsertValueCoerce("username", "Bob")
	user2Map.InsertValueCoerce("role", "admin")

	user3 := loro.NewLoroDoc()
	user3Map := user3.GetMap(doc_visitor.DATA_MAP_NAME)
	user3Map.InsertValueCoerce("id", "user3")
	user3Map.InsertValueCoerce("username", "Charlie")
	user3Map.InsertValueCoerce("role", "normal")

	postMeta1 := loro.NewLoroDoc()
	postMeta1Map := postMeta1.GetMap(doc_visitor.DATA_MAP_NAME)
	postMeta1Map.InsertValueCoerce("id", "post1")
	postMeta1Map.InsertValueCoerce("owner", "user1")
	postMeta1Map.InsertValueCoerce("title", "Learn Go In 5 Minutes")

	tr := &db_conn.Transaction{
		TxID:           "123e4567-e89b-12d3-a456-426614174000",
		TargetDatabase: "testdb",
		Committer:      "user1",
		Operations: []db_conn.TransactionOp{
			&db_conn.InsertOp{
				Collection: "users",
				DocID:      "user1",
				Snapshot:   user1.ExportSnapshot().Bytes(),
			},
			&db_conn.InsertOp{
				Collection: "users",
				DocID:      "user2",
				Snapshot:   user2.ExportSnapshot().Bytes(),
			},
			&db_conn.InsertOp{
				Collection: "users",
				DocID:      "user3",
				Snapshot:   user3.ExportSnapshot().Bytes(),
			},
			&db_conn.InsertOp{
				Collection: "postMetas",
				DocID:      "post1",
				Snapshot:   postMeta1.ExportSnapshot().Bytes(),
			},
		},
	}
	err = conn.Commit(tr)
	assert.NoError(t, err)
	return conn
}

func cleanupEngine(t *testing.T, conn db_conn.DbConnection) {
	err := conn.Close()
	assert.NoError(t, err)
}
