package eventreduce

import (
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/bdd"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/db_conn"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
)

// 一个 EventReducer 会根据 ListeningQuery lq 和 TransactionOp
// 计算出应该采取什么 Action 更新 lq 的结果集
type EventReducer interface {
	Reduce(lq query.ListeningQuery, op db_conn.TransactionOp) ActionName
}

var eventReducer EventReducer = NewMockEventReducer()

func GetEventReducer() EventReducer {
	// TODO
	return eventReducer
}

/////////////////

type MockEventReducer struct{}

var _ EventReducer = &MockEventReducer{}

func (m *MockEventReducer) Reduce(lq query.ListeningQuery, op db_conn.TransactionOp) ActionName {
	return ActionRunFullQueryAgain // mock 实例总是重新运行查询
}

func NewMockEventReducer() *MockEventReducer {
	return &MockEventReducer{}
}

/////////////////

type BddEventReducer struct {
	bdd *bdd.SimpleBdd
}

var _ EventReducer = &BddEventReducer{}

func NewBddEventReducerFromMinimalString(minimalBddString string) (*BddEventReducer, error) {
	bdd, err := bdd.NewBddFromMinimalString(minimalBddString)
	if err != nil {
		return nil, err
	}
	return &BddEventReducer{bdd: bdd}, nil
}

func (r *BddEventReducer) Reduce(lq query.ListeningQuery, op db_conn.TransactionOp) ActionName {
	panic("not implemented")
}
