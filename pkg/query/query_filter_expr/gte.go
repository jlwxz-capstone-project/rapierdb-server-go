package query_filter_expr

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js_value"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// GteExpr 大于等于比较
type GteExpr struct {
	Type QueryFilterExprType `json:"type"`
	O1   QueryFilterExpr     `json:"o1"`
	O2   QueryFilterExpr     `json:"o2"`
}

func NewGteExpr(o1 QueryFilterExpr, o2 QueryFilterExpr) *GteExpr {
	return &GteExpr{
		Type: ExprTypeGte,
		O1:   o1,
		O2:   o2,
	}
}

func (e *GteExpr) DebugSprint() string {
	return fmt.Sprintf("GteExpr{O1: %s, O2: %s}", e.O1.DebugSprint(), e.O2.DebugSprint())
}

func (e *GteExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	o1, err := e.O1.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating left operand of GTE: %v", ErrEvalError, err)
	}
	o2, err := e.O2.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating right operand of GTE: %v", ErrEvalError, err)
	}
	cmp, err := js_value.DeepComapreJsValue(o1.Value, o2.Value)
	if err != nil {
		return nil, err
	}
	return NewValueExpr(cmp >= 0), nil
}

func (e *GteExpr) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func newGteExprFromJson(msg json.RawMessage) (*GteExpr, error) {
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

	return NewGteExpr(o1, o2), nil
}
