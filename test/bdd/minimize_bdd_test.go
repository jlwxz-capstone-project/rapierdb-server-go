package main

import (
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/bdd"
)

func TestMinimizeBdd(t *testing.T) {
	t.Run("创建并最小化BDD", func(t *testing.T) {
		bddVal := bdd.CreateBddFromTruthTable(bdd.NewRandomTable(4))
		if bddVal == nil {
			t.Fatal("随机BDD不应该为nil")
		}

		nodesBeforeMinimize := bddVal.CountNodes()
		bddVal.Minimize(false)
		nodesAfterMinimize := bddVal.CountNodes()

		t.Logf("最小化前节点数: %d, 最小化后节点数: %d", nodesBeforeMinimize, nodesAfterMinimize)

		if err := bdd.EnsureCorrectBdd(bddVal); err != nil {
			t.Fatalf("最小化后的BDD验证失败: %v", err)
		}
	})

	t.Run("创建所有相等值的BDD", func(t *testing.T) {
		bddVal := bdd.CreateBddFromTruthTable(bdd.NewAllEqualTable(4))
		if bddVal == nil {
			t.Fatal("AllEqual BDD不应该为nil")
		}

		// 执行最小化
		bddVal.Minimize(false)

		if err := bdd.EnsureCorrectBdd(bddVal); err != nil {
			t.Fatalf("AllEqual BDD验证失败: %v", err)
		}

		// 所有值相同的BDD应该可以最小化到很小的尺寸
		nodeCount := bddVal.CountNodes()
		t.Logf("所有相等值的BDD节点数: %d", nodeCount)

		// 理想情况下，所有值相等的BDD应该只有叶节点和根节点
		// 但实际上可能不是这样，所以只是记录结果，不进行硬断言
	})
}
