package main

import (
	_ "embed"
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	"github.com/stretchr/testify/assert"
)

//go:embed test_permission_conditional.js
var testPermissionConditional string

func TestPermissionFromJs(t *testing.T) {
	t.Run("test_permission_conditional", func(t *testing.T) {
		_, err := query.NewPermissionFromJs(testPermissionConditional)
		assert.Nil(t, err)

		// 测试 CanView 方法
		// 1. 当 doc.owner === clientId 时，应该返回 true
		// mockDoc1 := loro.NewLoroDoc()
		// docMap1 := mockDoc1.GetMap("data")
		// err = docMap1.InsertString("owner", "user1")
		// assert.Nil(t, err)
		// result := permission.CanView("users", "doc1", mockDoc1, "user1")
		// assert.True(t, result)

		// // 2. 当 doc.owner !== clientId 时，应该返回 false
		// result = permission.CanView("users", "doc1", mockDoc1, "user2")
		// assert.False(t, result)
	})
}
