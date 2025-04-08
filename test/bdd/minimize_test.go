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

		second := nodes[1].AsInternalNode()
		second.ApplyRuductionRule()
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
		tt["000"] = 1
		tt["001"] = 1
		tt["010"] = 1
		bddVal := bdd.CreateBddFromTruthTable(tt)
		bddVal.Minimize()
		bdd.EnsureCorrectBdd(bddVal)
	})
}
