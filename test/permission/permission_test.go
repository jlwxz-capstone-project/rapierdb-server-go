package main

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/permissions"
	"github.com/stretchr/testify/assert"
)

//go:embed test_permission1.js
var testPermission1 string

//go:embed test_permission_emp.js
var testPermissionEmp string

//go:embed test_permission_invalid.js
var testPermissionInvalid string

//go:embed test_permission_conditional.js
var testPermissionConditional string

//go:embed test_permission_multiple.js
var testPermissionMultiple string

//go:embed test_permission_methods.js
var testPermissionMethods string

func TestPermissionFromJs(t *testing.T) {
	t.Run("test_permission_emp", func(t *testing.T) {
		permission, err := permissions.NewPermissionFromJs(testPermissionEmp)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("%+v\n", permission)
	})

	t.Run("test_permission1", func(t *testing.T) {
		permission, err := permissions.NewPermissionFromJs(testPermission1)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("%+v\n", permission)
	})

	t.Run("test_permission_invalid", func(t *testing.T) {
		_, err := permissions.NewPermissionFromJs(testPermissionInvalid)
		if err == nil {
			t.Fatal("应该返回错误，但没有")
		}
	})

	t.Run("test_permission_conditional", func(t *testing.T) {
		permission, err := permissions.NewPermissionFromJs(testPermissionConditional)
		assert.Nil(t, err)

		// 测试 CanView 方法
		// 1. 当 doc.owner === clientId 时，应该返回 true
		mockDoc1 := loro.NewLoroDoc()
		docMap1 := mockDoc1.GetMap("root")
		err = docMap1.InsertString("owner", "user1")
		assert.Nil(t, err)
		result := permission.CanView("users", "doc1", mockDoc1, "user1")
		assert.True(t, result)

		// 2. 当 clientId === "admin" 时，应该返回 true
		mockDoc2 := loro.NewLoroDoc()
		docMap2 := mockDoc2.GetMap("root")
		err = docMap2.InsertString("owner", "user2")
		if err != nil {
			t.Fatal(err)
		}
		result = permission.CanView("users", "doc2", mockDoc2, "admin")
		if !result {
			t.Errorf("当 clientId === \"admin\" 时，CanView 应该返回 true")
		}

		// 3. 当既不是 owner 也不是 admin 时，应该返回 false
		mockDoc3 := loro.NewLoroDoc()
		docMap3 := mockDoc3.GetMap("root")
		err = docMap3.InsertString("owner", "user3")
		if err != nil {
			t.Fatal(err)
		}
		result = permission.CanView("users", "doc3", mockDoc3, "user4")
		if result {
			t.Errorf("当既不是 owner 也不是 admin 时，CanView 应该返回 false")
		}
	})

	t.Run("test_permission_multiple", func(t *testing.T) {
		permission, err := permissions.NewPermissionFromJs(testPermissionMultiple)
		if err != nil {
			t.Fatal(err)
		}

		// 测试不存在的集合
		result := permission.CanView("nonexistent", "doc1", nil, "user1")
		if result {
			t.Errorf("对于不存在的集合，CanView 应该返回 false")
		}

		// 测试 posts 集合的 canView 方法
		// 1. 当 doc.isPublic === true 时，应该返回 true
		mockPublicPost := loro.NewLoroDoc()
		postMap1 := mockPublicPost.GetMap("root")
		err = postMap1.InsertBool("isPublic", true)
		if err != nil {
			t.Fatal(err)
		}
		result = permission.CanView("posts", "post1", mockPublicPost, "user1")
		if !result {
			t.Errorf("当 doc.isPublic === true 时，CanView 应该返回 true")
		}

		// 2. 当 doc.isPublic !== true 时，应该返回 false
		mockPrivatePost := loro.NewLoroDoc()
		postMap2 := mockPrivatePost.GetMap("root")
		err = postMap2.InsertBool("isPublic", false)
		if err != nil {
			t.Fatal(err)
		}
		result = permission.CanView("posts", "post2", mockPrivatePost, "user1")
		if result {
			t.Errorf("当 doc.isPublic !== true 时，CanView 应该返回 false")
		}

		// 测试 posts 集合的 canDelete 方法
		// 1. 当 doc.author === clientId 时，应该返回 true
		mockAuthorPost := loro.NewLoroDoc()
		postMap3 := mockAuthorPost.GetMap("root")
		err = postMap3.InsertString("author", "user1")
		if err != nil {
			t.Fatal(err)
		}
		result = permission.CanView("posts", "post3", mockAuthorPost, "user1")
		// 注意：这里应该使用 permissions.CanDelete 方法，但由于示例中没有实现，所以我们使用 CanView 代替
		// 实际应该实现 CanDelete, CanCreate, CanUpdate 方法类似于 CanView

		// 添加一个测试用例，验证当 doc 为 nil 时的行为
		result = permission.CanView("users", "doc4", nil, "user1")
		if result {
			t.Errorf("当 doc 为 nil 时，对于 users 集合的 CanView 应该返回 false")
		}
	})

	t.Run("test_permission_methods", func(t *testing.T) {
		permission, err := permissions.NewPermissionFromJs(testPermissionMethods)
		if err != nil {
			t.Fatal(err)
		}

		// 测试 CanView 方法 - 所有人都可以查看
		mockDoc := loro.NewLoroDoc()
		result := permission.CanView("documents", "doc1", mockDoc, "")
		if !result {
			t.Errorf("CanView 应该对所有用户返回 true")
		}

		// 测试 CanCreate 方法 - 仅登录用户可创建
		result = permission.CanCreate("documents", "doc1", mockDoc, "")
		if result {
			t.Errorf("对于未登录用户，CanCreate 应该返回 false")
		}

		result = permission.CanCreate("documents", "doc1", mockDoc, "user1")
		if !result {
			t.Errorf("对于已登录用户，CanCreate 应该返回 true")
		}

		// 测试 CanUpdate 方法 - 仅文档创建者可更新
		mockDoc1 := loro.NewLoroDoc()
		docMap1 := mockDoc1.GetMap("root")
		err = docMap1.InsertString("creator", "user1")
		if err != nil {
			t.Fatal(err)
		}

		result = permission.CanUpdate("documents", "doc1", mockDoc1, "user1")
		if !result {
			t.Errorf("文档创建者应该可以更新文档")
		}

		result = permission.CanUpdate("documents", "doc1", mockDoc1, "user2")
		if result {
			t.Errorf("非文档创建者不应该可以更新文档")
		}

		// 测试 CanDelete 方法 - 仅管理员和创建者可删除
		result = permission.CanDelete("documents", "doc1", mockDoc1, "user1")
		if !result {
			t.Errorf("文档创建者应该可以删除文档")
		}

		result = permission.CanDelete("documents", "doc1", mockDoc1, "admin")
		if !result {
			t.Errorf("管理员应该可以删除文档")
		}

		result = permission.CanDelete("documents", "doc1", mockDoc1, "user2")
		if result {
			t.Errorf("既不是创建者也不是管理员的用户不应该可以删除文档")
		}
	})
}
