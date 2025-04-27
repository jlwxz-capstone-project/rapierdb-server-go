package query_filter_expr

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js_value"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// LteExpr 小于等于比较
type LteExpr struct {
	Type QueryFilterExprType `json:"type"`
	O1   QueryFilterExpr     `json:"o1"`
	O2   QueryFilterExpr     `json:"o2"`
}

func NewLteExpr(o1 QueryFilterExpr, o2 QueryFilterExpr) *LteExpr {
	return &LteExpr{
		Type: ExprTypeLte,
		O1:   o1,
		O2:   o2,
	}
}

func (e *LteExpr) DebugSprint() string {
	return fmt.Sprintf("LteExpr{O1: %s, O2: %s}", e.O1.DebugSprint(), e.O2.DebugSprint())
}

func (e *LteExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	o1, err := e.O1.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating left operand of LTE: %v", ErrEvalError, err)
	}
	o2, err := e.O2.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating right operand of LTE: %v", ErrEvalError, err)
	}
	cmp, err := js_value.DeepComapreJsValue(o1.Value, o2.Value)
	if err != nil {
		return nil, err
	}
	return NewValueExpr(cmp <= 0), nil
}

func (e *LteExpr) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func newLteExprFromJson(msg json.RawMessage) (*LteExpr, error) {
	var temp struct {
		Type QueryFilterExprType `json:"type"`
		O1   json.RawMessage     `json:"o1"`
		O2   json.RawMessage     `json:"o2"`
	}

	if err := json.Unmarshal(msg, &temp); err != nil {
		return nil, err
	}

	o1, err := NewQueryFilterExprFromJson(temp.O1)
	if err != nil {
		return nil, err
	}

	o2, err := NewQueryFilterExprFromJson(temp.O2)
	if err != nil {
		return nil, err
	}

	return NewLteExpr(o1, o2), nil
}
