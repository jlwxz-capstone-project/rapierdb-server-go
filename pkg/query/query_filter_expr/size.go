package query_filter_expr

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// SizeExpr 检查数组的长度
type SizeExpr struct {
	Type   QueryFilterExprType `json:"type"`
	Target QueryFilterExpr     `json:"target"`
	Size   QueryFilterExpr     `json:"size"`
}

func NewSizeExpr(target QueryFilterExpr, size QueryFilterExpr) *SizeExpr {
	return &SizeExpr{
		Type:   ExprTypeSize,
		Target: target,
		Size:   size,
	}
}

func (e *SizeExpr) DebugPrint() string {
	return fmt.Sprintf("SizeExpr{Target: %s, Size: %s}", e.Target.DebugPrint(), e.Size.DebugPrint())
}

func (e *SizeExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	// 评估字段表达式
	target, err := e.Target.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating target in SIZE: %v", ErrEvalError, err)
	}

	// 检查字段是否为数组
	arr, ok := target.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("%w: expected array in SIZE expression, got %T", ErrTypeError, target.Value)
	}

	// 评估大小表达式
	size, err := e.Size.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating size in SIZE: %v", ErrEvalError, err)
	}

	// 检查大小是否为整数
	sizeValue, ok := size.Value.(int64)
	if !ok {
		return nil, fmt.Errorf("%w: expected integer in SIZE expression, got %T", ErrTypeError, size.Value)
	}

	return &ValueExpr{Value: int64(len(arr)) == sizeValue}, nil
}

func (e *SizeExpr) MarshalJSON() ([]byte, error) {
	return json.Marshal(e)
}
