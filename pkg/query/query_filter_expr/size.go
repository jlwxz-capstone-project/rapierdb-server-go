package query_filter_expr

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// SizeExpr 检查数组的长度
type SizeExpr struct {
	Field QueryFilterExpr
	Size  QueryFilterExpr
}

func (e *SizeExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	// 评估字段表达式
	field, err := e.Field.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating field in SIZE: %v", ErrEvalError, err)
	}

	// 检查字段是否为数组
	arr, ok := field.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("%w: expected array in SIZE expression, got %T", ErrTypeError, field.Value)
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
	fieldData, err := e.Field.MarshalJSON()
	if err != nil {
		return nil, err
	}

	sizeData, err := e.Size.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return json.Marshal(SerializedQueryFilterExpr{
		Type: ExprTypeSize,
		O1:   fieldData,
		O2:   sizeData,
	})
}

func (e *SizeExpr) UnmarshalJSON(data []byte) error {
	var s SerializedQueryFilterExpr
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s.Type != ExprTypeSize {
		return fmt.Errorf("expected SIZE expression, got %s", s.Type)
	}
	if s.O1 == nil || s.O2 == nil {
		return fmt.Errorf("missing field or size for SIZE expression")
	}

	field, err := UnmarshalQueryFilterExpr(s.O1)
	if err != nil {
		return fmt.Errorf("failed to unmarshal field: %v", err)
	}

	size, err := UnmarshalQueryFilterExpr(s.O2)
	if err != nil {
		return fmt.Errorf("failed to unmarshal size: %v", err)
	}

	e.Field = field
	e.Size = size
	return nil
}
