package main

import (
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/bdd"
	"github.com/stretchr/testify/assert"
)

func TestFindSimilarNode(t *testing.T) {
	t.Run("should be equal to equal node of other bdd", func(t *testing.T) {
		tt := bdd.NewAllEqualTable(3)
		bddVal := bdd.CreateBddFromTruthTable(tt)
		nodes := bddVal.GetNodesOfLevel(1)

		var first *bdd.InternalNode
		for _, node := range nodes {
			if first == nil {
				first = node.AsInternalNode()
				break
			}
		}

		bddVal2 := bdd.CreateBddFromTruthTable(tt)
		nodes2 := bddVal2.GetNodesOfLevel(1)

		var first2 *bdd.InternalNode
		for _, node := range nodes2 {
			if first2 == nil {
				first2 = node.AsInternalNode()
				break
			}
		}

		found := bdd.FindSimilarNode(first.AsNode(), []*bdd.Node{first2.AsNode()})
		assert.NotNil(t, found)
	})

	t.Run("should not find itself", func(t *testing.T) {
		tt := bdd.NewAllEqualTable(3)
		bddVal := bdd.CreateBddFromTruthTable(tt)
		nodes := bddVal.GetNodesOfLevel(1)

		var first *bdd.InternalNode
		for _, node := range nodes {
			if first == nil {
				first = node.AsInternalNode()
				break
			}
		}

		found := bdd.FindSimilarNode(first.AsNode(), []*bdd.Node{first.AsNode()})
		assert.Nil(t, found)
	})

	t.Run("should be not equal to root node", func(t *testing.T) {
		tt := bdd.NewAllEqualTable(3)
		bddVal := bdd.CreateBddFromTruthTable(tt)
		nodes := bddVal.GetNodesOfLevel(1)

		var first *bdd.InternalNode
		for _, node := range nodes {
			if first == nil {
				first = node.AsInternalNode()
				break
			}
		}

		found := bdd.FindSimilarNode(first.AsNode(), []*bdd.Node{bddVal.AsNode()})
		assert.Nil(t, found)
	})
}
