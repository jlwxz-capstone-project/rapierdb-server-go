package eventreduce

import (
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/types"
)

// 导出主要函数
var (
	// 操作计算函数
	CalculateActionFromMapFunc  = CalculateActionFromMap
	CalculateActionNameFunc     = CalculateActionName
	CalculateActionFunctionFunc = CalculateActionFunction
	RunActionFunc               = RunAction
)

// 从其他包导出的类型
type (
	ActionName           = types.ActionName
	ActionFunction       = types.ActionFunction
	ActionFunctionInput  = types.ActionFunctionInput
	StateSet             = types.StateSet
	StateSetToActionMap  = types.StateSetToActionMap
	ChangeEvent          = types.ChangeEvent
	ChangeEventBase      = types.ChangeEventBase
	ChangeEventDelete    = types.ChangeEventDelete
	ChangeEventInsert    = types.ChangeEventInsert
	ChangeEventUpdate    = types.ChangeEventUpdate
	QueryParams          = types.QueryParams
	ResultKeyDocumentMap = types.ResultKeyDocumentMap
	StateResolveFunction = types.StateResolveFunction
	WriteOperation       = types.WriteOperation
)
