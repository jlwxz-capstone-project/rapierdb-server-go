package main

import (
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/bdd"
)

func TestCreateBddFromTruthTable(t *testing.T) {
	t.Run("应该创建一个 BDD", func(t *testing.T) {
		tt := bdd.NewExampleTruthTable(3)
		bddVal := bdd.CreateBddFromTruthTable(tt)
		if bddVal == nil {
			t.Fatal("BDD 不应该为 nil")
		}

		if err := bdd.EnsureCorrectBdd(bddVal); err != nil {
			t.Fatalf("BDD 验证失败: %v", err)
		}
	})

	t.Run("应该创建一个较大的 BDD", func(t *testing.T) {
		bddVal := bdd.CreateBddFromTruthTable(bdd.NewExampleTruthTable(5))
		if bddVal == nil {
			t.Fatal("BDD 不应该为 nil")
		}

		if err := bdd.EnsureCorrectBdd(bddVal); err != nil {
			t.Fatalf("大型 BDD 验证失败: %v", err)
		}
	})
}
