package main

import (
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/bdd"
	"github.com/stretchr/testify/assert"
)

func TestResolve(t *testing.T) {
	t.Run("should have the same values as the truth table", func(t *testing.T) {
		size := 8
		tt := bdd.NewExampleTruthTable(size)
		bddVal := bdd.CreateBddFromTruthTable(tt)
		resolvers := bdd.GetResolverFunctions(size, false)

		for key, val := range tt {
			res := bddVal.Resolve(resolvers, key)
			assert.Equal(t, val, res)
		}
	})

	t.Run("should have the same values after minimize", func(t *testing.T) {
		size := 8
		tt := bdd.NewExampleTruthTable(size)
		bddVal := bdd.CreateBddFromTruthTable(tt)
		resolvers := bdd.GetResolverFunctions(size, false)

		bddVal.Minimize()
		for key, val := range tt {
			res := bddVal.Resolve(resolvers, key)
			assert.Equal(t, val, res)
		}
	})

	t.Run("should work for random table", func(t *testing.T) {
		size := 8
		tt := bdd.NewRandomTable(size)
		bddVal := bdd.CreateBddFromTruthTable(tt)
		resolvers := bdd.GetResolverFunctions(size, false)

		for key, val := range tt {
			res := bddVal.Resolve(resolvers, key)
			assert.Equal(t, val, res)
		}
	})
}
