package query_filter_expr

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// AllExpr 检查数组是否包含所有指定元素
type AllExpr struct {
	Field QueryFilterExpr
	Items []QueryFilterExpr
}

func (e *AllExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	// 评估字段表达式
	field, err := e.Field.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating field in ALL: %v", ErrEvalError, err)
	}

	// 检查字段是否为数组
	arr, ok := field.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("%w: expected array in ALL expression, got %T", ErrTypeError, field.Value)
	}

	// 检查数组是否为空
	if len(arr) == 0 {
		return &ValueExpr{Value: false}, nil
	}

	// 检查每个要求的元素是否都在数组中
	for _, item := range e.Items {
		itemValue, err := item.Eval(doc)
		if err != nil {
			return nil, fmt.Errorf("%w: evaluating item in ALL: %v", ErrEvalError, err)
		}

		found := false
		for _, arrItem := range arr {
			cmp, err := CompareValues(arrItem, itemValue.Value)
			if err != nil {
				if errors.Is(err, ErrTypeError) {
					continue // 类型不匹配，继续检查下一个元素
				}
				return nil, err
			}
			if cmp == 0 {
				found = true
				break
			}
		}

		if !found {
			return &ValueExpr{Value: false}, nil
		}
	}

	return &ValueExpr{Value: true}, nil
}

func (e *AllExpr) MarshalJSON() ([]byte, error) {
	fieldData, err := e.Field.MarshalJSON()
	if err != nil {
		return nil, err
	}

	items := make([]json.RawMessage, len(e.Items))
	for i, item := range e.Items {
		data, err := item.MarshalJSON()
		if err != nil {
			return nil, err
		}
		items[i] = data
	}

	return json.Marshal(SerializedQueryFilterExpr{
		Type: ExprTypeAll,
		O1:   fieldData,
		List: items,
	})
}

func (e *AllExpr) UnmarshalJSON(data []byte) error {
	var s SerializedQueryFilterExpr
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s.Type != ExprTypeAll {
		return fmt.Errorf("expected ALL expression, got %s", s.Type)
	}
	if s.O1 == nil || len(s.List) == 0 {
		return fmt.Errorf("missing field or items for ALL expression")
	}

	field, err := UnmarshalQueryFilterExpr(s.O1)
	if err != nil {
		return fmt.Errorf("failed to unmarshal field: %v", err)
	}

	items := make([]QueryFilterExpr, len(s.List))
	for i, item := range s.List {
		expr, err := UnmarshalQueryFilterExpr(item)
		if err != nil {
			return fmt.Errorf("failed to unmarshal item %d: %v", i, err)
		}
		items[i] = expr
	}

	e.Field = field
	e.Items = items
	return nil
}
