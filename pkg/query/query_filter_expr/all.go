package query_filter_expr

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// AllExpr 检查数组是否包含所有指定元素
type AllExpr struct {
	Type   QueryFilterExprType `json:"type"`
	Target QueryFilterExpr     `json:"target"`
	Items  []QueryFilterExpr   `json:"items"`
}

func (e *AllExpr) DebugPrint() string {
	items := make([]string, len(e.Items))
	for i, item := range e.Items {
		items[i] = item.DebugPrint()
	}
	itemsStr := fmt.Sprintf("[%s]", strings.Join(items, ", "))
	return fmt.Sprintf("AllExpr{Target: %s, Items: %s}", e.Target.DebugPrint(), itemsStr)
}

func (e *AllExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	// 评估目标表达式
	target, err := e.Target.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating target in ALL: %v", ErrEvalError, err)
	}

	// 检查目标是否为数组
	arr, ok := target.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("%w: expected array in ALL expression, got %T", ErrTypeError, target.Value)
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
			cmp, err := util.CompareValues(arrItem, itemValue.Value)
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

func (e *AllExpr) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func NewAllExpr(target QueryFilterExpr, items []QueryFilterExpr) *AllExpr {
	return &AllExpr{
		Type:   ExprTypeAll, // Assuming ExprTypeAll is the correct constant
		Target: target,
		Items:  items,
	}
}

func newAllExprFromJson(msg json.RawMessage) (*AllExpr, error) {
	var temp struct {
		Type   QueryFilterExprType `json:"type"`
		Target json.RawMessage     `json:"target"`
		Items  []json.RawMessage   `json:"items"`
	}

	if err := json.Unmarshal(msg, &temp); err != nil {
		return nil, err
	}

	target, err := NewQueryFilterExprFromJson(temp.Target)
	if err != nil {
		return nil, err
	}

	items := make([]QueryFilterExpr, len(temp.Items))
	for i, item := range temp.Items {
		item, err := NewQueryFilterExprFromJson(item)
		if err != nil {
			return nil, err
		}
		items[i] = item
	}

	return NewAllExpr(target, items), nil
}
