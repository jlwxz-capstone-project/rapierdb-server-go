package query_filter_expr

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js_value"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// NeExpr 不等比较
type NeExpr struct {
	Type QueryFilterExprType `json:"type"`
	O1   QueryFilterExpr     `json:"o1"`
	O2   QueryFilterExpr     `json:"o2"`
}

func NewNeExpr(o1 QueryFilterExpr, o2 QueryFilterExpr) *NeExpr {
	return &NeExpr{
		Type: ExprTypeNe,
		O1:   o1,
		O2:   o2,
	}
}

func (e *NeExpr) DebugPrint() string {
	return fmt.Sprintf("NeExpr{O1: %s, O2: %s}", e.O1.DebugPrint(), e.O2.DebugPrint())
}

func (e *NeExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	o1, err := e.O1.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating left operand of NE: %v", ErrEvalError, err)
	}
	o2, err := e.O2.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating right operand of NE: %v", ErrEvalError, err)
	}
	cmp, err := js_value.DeepComapreJsValue(o1.Value, o2.Value)
	if err != nil {
		return nil, err
	}
	return NewValueExpr(cmp != 0), nil
}

func (e *NeExpr) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func newNeExprFromJson(msg json.RawMessage) (*NeExpr, error) {
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

	return NewNeExpr(o1, o2), nil
}
