package query_filter_expr

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js_value"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	pe "github.com/pkg/errors"
)

// EqExpr 相等比较
type EqExpr struct {
	Type QueryFilterExprType `json:"type"`
	O1   QueryFilterExpr     `json:"o1"`
	O2   QueryFilterExpr     `json:"o2"`
}

func NewEqExpr(o1 QueryFilterExpr, o2 QueryFilterExpr) *EqExpr {
	return &EqExpr{
		Type: ExprTypeEq,
		O1:   o1,
		O2:   o2,
	}
}

func (e *EqExpr) DebugPrint() string {
	return fmt.Sprintf("EqExpr{O1: %s, O2: %s}", e.O1.DebugPrint(), e.O2.DebugPrint())
}

func (e *EqExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	o1, err := e.O1.Eval(doc)
	if err != nil {
		return nil, pe.Wrapf(ErrEvalError, "evaluating left operand of EQ: %v", err)
	}
	o2, err := e.O2.Eval(doc)
	if err != nil {
		return nil, pe.Wrapf(ErrEvalError, "evaluating right operand of EQ: %v", err)
	}
	cmp, err := js_value.DeepComapreJsValue(o1.Value, o2.Value)
	if err != nil {
		return nil, err
	}
	return NewValueExpr(cmp == 0), nil
}

func (e *EqExpr) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func newEqExprFromJson(msg json.RawMessage) (*EqExpr, error) {
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

	return NewEqExpr(o1, o2), nil
}
