package main

import (
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/bdd"
	"github.com/stretchr/testify/assert"
)

func TestApplyReductionRule(t *testing.T) {
	t.Run("should remove itself", func(t *testing.T) {
		tt := bdd.NewAllEqualTable(3)
		bddVal := bdd.CreateBddFromTruthTable(tt)
		nodes := bddVal.GetNodesOfLevel(2)

		first := nodes[0].AsInternalNode()
		first.ApplyRuductionRule()
		bdd.EnsureCorrectBdd(bddVal)
		assert.True(t, bddVal.Branches.GetBranch("0").GetBranches().GetBranch("0").IsLeafNode())

		second := nodes[1].AsInternalNode()
		second.ApplyRuductionRule()
		bdd.EnsureCorrectBdd(bddVal)
		assert.True(t, bddVal.Branches.GetBranch("0").GetBranches().GetBranch("1").IsLeafNode())
	})

	t.Run("should work on deeper bdd itself", func(t *testing.T) {
		tt := bdd.NewAllEqualTable(4)
		bddVal := bdd.CreateBddFromTruthTable(tt)
		nodes := bddVal.GetNodesOfLevel(2)
		first := nodes[0].AsInternalNode()
		first.ApplyRuductionRule()
		bdd.EnsureCorrectBdd(bddVal)
		assert.True(t, bddVal.
			GetBranches().
			GetBranch("0").
			GetBranches().
			GetBranch("0").
			GetBranches().
			GetBranch("0").
			IsLeafNode(),
		)
	})
}

func TestApplyEliminationRule(t *testing.T) {
	t.Run("should remove the found one", func(t *testing.T) {
		tt := bdd.NewAllEqualTable(3)
		bddVal := bdd.CreateBddFromTruthTable(tt)
		nodes := bddVal.GetNodesOfLevel(1)
		first := nodes[0].AsInternalNode()
		first.ApplyEliminationRule(nil)
		bdd.EnsureCorrectBdd(bddVal)
		assert.Equal(t,
			bddVal.Branches.GetBranch("0").Id,
			bddVal.Branches.GetBranch("1").Id,
		)

		nodes = bddVal.GetNodesOfLevel(1)
		second := nodes[0].AsInternalNode()
		second.ApplyEliminationRule(nil)
		bdd.EnsureCorrectBdd(bddVal)
	})
}

func TestMinimize(t *testing.T) {
	t.Run("should return a minimized bdd", func(t *testing.T) {
		tt := bdd.NewAllEqualTable(3)
		bddVal := bdd.CreateBddFromTruthTable(tt)
		bddVal.Minimize()
		assert.True(t, bddVal.Branches.GetBranch("0").IsLeafNode())
		assert.True(t, bddVal.Branches.GetBranch("1").IsLeafNode())
		bdd.EnsureCorrectBdd(bddVal)
	})

	t.Run("should not crash on random table", func(t *testing.T) {
		tt := bdd.NewExampleTruthTable(3)
		tt.Set("000", 1)
		tt.Set("001", 1)
		tt.Set("010", 1)
		bddVal := bdd.CreateBddFromTruthTable(tt)
		bddVal.Minimize()
		assert.True(t, bddVal.Branches.GetBranch("0").GetBranches().GetBranch("0").IsLeafNode())
		bdd.EnsureCorrectBdd(bddVal)
	})

	t.Run("should not crash on a really big table", func(t *testing.T) {
		depth := 11
		tt := bdd.NewRandomTable(depth)
		bddVal := bdd.CreateBddFromTruthTable(tt)
		bddVal.Minimize()
		bdd.EnsureCorrectBdd(bddVal)
	})
}

func TestCountNodes(t *testing.T) {
	t.Run("should be smaller after minimize", func(t *testing.T) {
		tt := bdd.NewAllEqualTable(3)
		bddVal := bdd.CreateBddFromTruthTable(tt)
		before := bddVal.CountNodes()
		bddVal.Minimize()
		after := bddVal.CountNodes()
		assert.Greater(t, before, after)
	})
}

func TestRemoveIrrelevantLeafNodes(t *testing.T) {
	t.Run("should remove all irrelevant nodes", func(t *testing.T) {
		unknown := -1
		tt := bdd.NewExampleTruthTable(5)
		tt.Set("00001", unknown)
		tt.Set("00000", unknown)
		tt.Set("00101", unknown)
		bddVal := bdd.CreateBddFromTruthTable(tt)
		bddVal.RemoveIrrelevantLeafNodes(unknown)
		for _, node := range bddVal.GetLeafNodes() {
			assert.NotEqual(t, unknown, node.Value)
		}
	})

	t.Run("should work on a big table", func(t *testing.T) {
		tt := bdd.NewRandomUnknownTable(6)
		bddVal := bdd.CreateBddFromTruthTable(tt)
		bddVal.RemoveIrrelevantLeafNodes(bdd.Unknown)
		for _, node := range bddVal.GetLeafNodes() {
			assert.NotEqual(t, bdd.Unknown, node.Value)
		}
	})
}
