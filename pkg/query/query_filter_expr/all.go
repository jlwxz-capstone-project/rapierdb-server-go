package query_filter_expr

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js_value"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	pe "github.com/pkg/errors"
)

// AllExpr 检查数组是否包含所有指定元素
type AllExpr struct {
	Type   QueryFilterExprType `json:"type"`
	Target QueryFilterExpr     `json:"target"`
	Items  []QueryFilterExpr   `json:"items"`
}

func (e *AllExpr) DebugSprint() string {
	items := make([]string, len(e.Items))
	for i, item := range e.Items {
		items[i] = item.DebugSprint()
	}
	itemsStr := fmt.Sprintf("[%s]", strings.Join(items, ", "))
	return fmt.Sprintf("AllExpr{Target: %s, Items: %s}", e.Target.DebugSprint(), itemsStr)
}

func (e *AllExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	// 评估目标表达式
	target, err := e.Target.Eval(doc)
	if err != nil {
		return nil, pe.Wrapf(ErrEvalError, "evaluating target in ALL: %v", err)
	}

	// 检查目标是否为数组
	arr, ok := target.Value.([]any)
	if !ok {
		return nil, pe.Wrapf(ErrTypeError, "expected array in ALL expression, got %T", target.Value)
	}

	// 检查数组是否为空
	if len(arr) == 0 {
		return NewValueExpr(false), nil
	}

	// 检查每个要求的元素是否都在数组中
	for _, item := range e.Items {
		itemValue, err := item.Eval(doc)
		if err != nil {
			return nil, pe.Wrapf(ErrEvalError, "evaluating item in ALL: %v", err)
		}

		found := false
		for _, arrItem := range arr {
			cmp, err := js_value.DeepComapreJsValue(arrItem, itemValue.Value)
			if err != nil {
				if pe.Is(err, js_value.ErrCompareTypeMismatch) {
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
			return NewValueExpr(false), nil
		}
	}

	return NewValueExpr(true), nil
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
