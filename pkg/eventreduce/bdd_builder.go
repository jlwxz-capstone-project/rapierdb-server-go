package eventreduce

import "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/bdd"

func buildBDD() *bdd.RootNode {
	tt := bdd.NewEmptyTruthTable()
	fillTruthTable(tt)
	bddVal := bdd.CreateBddFromTruthTable(tt)
	bddVal.Minimize()
	return bddVal
}

func fillTruthTable(tt bdd.TruthTable) {
	panic("not implemented")
}
