package main

import (
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/bdd"
)

func TestCreateBddFromTruthTable(t *testing.T) {
	t.Run("应该创建一个BDD", func(t *testing.T) {
		bddVal := bdd.CreateBddFromTruthTable(bdd.NewExampleTruthTable(3))
		if bddVal == nil {
			t.Fatal("BDD不应该为nil")
		}

		if err := bdd.EnsureCorrectBdd(bddVal); err != nil {
			t.Fatalf("BDD验证失败: %v", err)
		}
	})

	t.Run("应该创建一个较大的BDD", func(t *testing.T) {
		bddVal := bdd.CreateBddFromTruthTable(bdd.NewExampleTruthTable(5))
		if bddVal == nil {
			t.Fatal("BDD不应该为nil")
		}

		if err := bdd.EnsureCorrectBdd(bddVal); err != nil {
			t.Fatalf("大型BDD验证失败: %v", err)
		}
	})
}
