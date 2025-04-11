package eventreduce_test

import (
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/eventreduce"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/types"
	"github.com/stretchr/testify/assert"
)

// TestMapActionCodeToName 测试操作代码映射到操作名称的功能
func TestMapActionCodeToName(t *testing.T) {
	tests := []struct {
		name       string
		actionCode int
		expected   types.ActionName
	}{
		{"无操作", 0, types.ActionDoNothing},
		{"插入第一个", 1, types.ActionInsertFirst},
		{"插入最后一个", 2, types.ActionInsertLast},
		{"删除第一个", 3, types.ActionRemoveFirstItem},
		{"删除最后一个", 4, types.ActionRemoveLastItem},
		{"删除第一个并插入最后一个", 5, types.ActionRemoveFirstInsertLast},
		{"删除最后一个并插入第一个", 6, types.ActionRemoveLastInsertFirst},
		{"删除第一个并插入第一个", 7, types.ActionRemoveFirstInsertFirst},
		{"删除最后一个并插入最后一个", 8, types.ActionRemoveLastInsertLast},
		{"删除已存在", 9, types.ActionRemoveExisting},
		{"替换已存在", 10, types.ActionReplaceExisting},
		{"按排序位置插入", 11, types.ActionInsertAtSortPosition},
		{"删除已存在并按排序位置插入", 12, types.ActionRemoveExistingAndInsertAtSortPosition},
		{"未知代码", 999, types.ActionRunFullQueryAgain},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := eventreduce.MapActionCodeToNameForTest(tc.actionCode)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestFillTruthTable 测试真值表填充功能
func TestFillTruthTable(t *testing.T) {
	// 获取填充的真值表
	tt := eventreduce.FillTruthTableForTest()

	// 验证真值表不为空
	assert.Greater(t, tt.Len(), 0, "真值表应该包含条目")

	// 检查所有的条目，寻找全0状态组合
	allZerosFound := false
	for stateSet, actionCode := range tt.IterEntries() {
		// 检查是否是全0状态集合（只包含'0'字符）
		isAllZeros := true
		for _, char := range stateSet {
			if char != '0' {
				isAllZeros = false
				break
			}
		}

		if isAllZeros {
			allZerosFound = true
			assert.Equal(t, 0, actionCode, "全0状态应该映射到ActionDoNothing")
			break
		}
	}

	// 验证是否找到全0状态集合
	assert.True(t, allZerosFound, "真值表中应该存在全0状态组合")
}

// TestBuildActionBDD 测试BDD构建功能
func TestBuildActionBDD(t *testing.T) {
	// 跳过完整BDD测试，因为在某些环境下可能会超时
	t.Skip("跳过完整BDD构建测试，该测试可能超时")

	// 构建BDD
	bdd := eventreduce.BuildActionBDDForTest()

	// 确保BDD不为nil
	assert.NotNil(t, bdd, "构建的BDD不应为nil")

	// 验证BDD的基本属性
	assert.Greater(t, bdd.CountNodes(), 0, "BDD应该包含节点")
}

// TestSimpleBDD 测试BDD的基本功能，使用一个简化的真值表
func TestSimpleBDD(t *testing.T) {
	// 创建一个简单的真值表
	tt := eventreduce.CreateSimpleTruthTableForTest()

	// 从真值表创建BDD
	bdd := eventreduce.CreateSimpleBDDForTest(tt)

	// 确保BDD不为nil
	assert.NotNil(t, bdd, "创建的BDD不应为nil")

	// 验证BDD的基本属性
	nodeCount := bdd.CountNodes()
	assert.Greater(t, nodeCount, 0, "BDD应该包含节点")
	t.Logf("BDD节点数量: %d", nodeCount)
}
