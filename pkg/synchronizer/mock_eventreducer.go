package synchronizer

import (
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/types"
)

// 一个 EventReducer 会根据 ListeningQuery lq 和 TransactionOp
// 计算出应该采取什么 Action 更新 lq 的结果集
type EventReducer interface {
	Reduce(lq ListeningQuery, op storage_engine.TransactionOp) types.ActionName
}

type MockEventReducer struct{}

func (m *MockEventReducer) Reduce(lq ListeningQuery, op storage_engine.TransactionOp) types.ActionName {
	return types.ActionRunFullQueryAgain // mock 实例总是重新运行查询
}

func NewMockEventReducer() EventReducer {
	return &MockEventReducer{}
}
