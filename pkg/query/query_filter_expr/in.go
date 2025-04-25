package query_filter_expr

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js_value"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// InExpr 包含比较
type InExpr struct {
	Type QueryFilterExprType `json:"type"`
	O1   QueryFilterExpr     `json:"o1"`
	O2   []QueryFilterExpr   `json:"o2"`
}

func NewInExpr(o1 QueryFilterExpr, o2 []QueryFilterExpr) *InExpr {
	return &InExpr{
		Type: ExprTypeIn,
		O1:   o1,
		O2:   o2,
	}
}

func (e *InExpr) DebugPrint() string {
	o2 := make([]string, len(e.O2))
	for i, expr := range e.O2 {
		o2[i] = expr.DebugPrint()
	}
	o2Str := fmt.Sprintf("[%s]", strings.Join(o2, ", "))
	return fmt.Sprintf("InExpr{O1: %s, O2: %s}", e.O1.DebugPrint(), o2Str)
}

func (e *InExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	o1, err := e.O1.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating target of IN: %v", ErrEvalError, err)
	}
	for _, elem := range e.O2 {
		elemValue, err := elem.Eval(doc)
		if err != nil {
			return nil, fmt.Errorf("%w: evaluating element of IN: %v", ErrEvalError, err)
		}
		cmp, err := js_value.DeepComapreJsValue(o1.Value, elemValue.Value)
		if err != nil {
			return nil, err
		}
		if cmp == 0 {
			return &ValueExpr{Value: true}, nil
		}
	}
	return &ValueExpr{Value: false}, nil
}

func (e *InExpr) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func newInExprFromJson(msg json.RawMessage) (*InExpr, error) {
	var temp struct {
		Type QueryFilterExprType `json:"type"`
		O1   json.RawMessage     `json:"o1"`
		O2   []json.RawMessage   `json:"o2"`
	}

	if err := json.Unmarshal(msg, &temp); err != nil {
		return nil, err
	}

	o1, err := NewQueryFilterExprFromJson(temp.O1)
	if err != nil {
		return nil, err
	}

	o2 := make([]QueryFilterExpr, len(temp.O2))
	for i, expr := range temp.O2 {
		o2[i], err = NewQueryFilterExprFromJson(expr)
		if err != nil {
			return nil, err
		}
	}

	return NewInExpr(o1, o2), nil
}
