package types_test

import (
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/types"
	"github.com/stretchr/testify/assert"
)

// TestNewInsertEvent 测试创建插入事件的函数
func TestNewInsertEvent(t *testing.T) {
	// 准备测试数据
	id := "doc123"
	doc := map[string]interface{}{
		"_id":    id,
		"name":   "测试文档",
		"active": true,
	}

	// 调用被测试函数
	event := types.NewInsertEvent(id, doc)

	// 验证结果
	assert.Equal(t, types.OperationInsert, event.Operation)
	assert.Equal(t, id, event.ID)
	assert.Equal(t, doc, event.Doc)
	assert.Nil(t, event.Previous)
}

// TestNewUpdateEvent 测试创建更新事件的函数
func TestNewUpdateEvent(t *testing.T) {
	// 准备测试数据
	id := "doc123"
	newDoc := map[string]interface{}{
		"_id":    id,
		"name":   "更新后的文档",
		"active": true,
	}
	prevDoc := map[string]interface{}{
		"_id":    id,
		"name":   "原始文档",
		"active": false,
	}

	// 调用被测试函数
	event := types.NewUpdateEvent(id, newDoc, prevDoc)

	// 验证结果
	assert.Equal(t, types.OperationUpdate, event.Operation)
	assert.Equal(t, id, event.ID)
	assert.Equal(t, newDoc, event.Doc)
	assert.Equal(t, prevDoc, event.Previous)
}

// TestNewDeleteEvent 测试创建删除事件的函数
func TestNewDeleteEvent(t *testing.T) {
	// 准备测试数据
	id := "doc123"
	prevDoc := map[string]interface{}{
		"_id":    id,
		"name":   "即将删除的文档",
		"active": true,
	}

	// 调用被测试函数
	event := types.NewDeleteEvent(id, prevDoc)

	// 验证结果
	assert.Equal(t, types.OperationDelete, event.Operation)
	assert.Equal(t, id, event.ID)
	assert.Nil(t, event.Doc)
	assert.Equal(t, prevDoc, event.Previous)
}

// TestChangeEventOperationConstants 测试写入操作常量
func TestChangeEventOperationConstants(t *testing.T) {
	// 确保常量值符合预期
	assert.Equal(t, types.WriteOperation("INSERT"), types.OperationInsert)
	assert.Equal(t, types.WriteOperation("UPDATE"), types.OperationUpdate)
	assert.Equal(t, types.WriteOperation("DELETE"), types.OperationDelete)
}
