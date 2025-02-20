package query_filter_expr

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// EqExpr 相等比较
type EqExpr struct {
	O1 QueryFilterExpr
	O2 QueryFilterExpr
}

func (e *EqExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	o1, err := e.O1.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating left operand of EQ: %v", ErrEvalError, err)
	}
	o2, err := e.O2.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating right operand of EQ: %v", ErrEvalError, err)
	}
	cmp, err := util.CompareValues(o1.Value, o2.Value)
	if err != nil {
		return nil, err
	}
	return &ValueExpr{Value: cmp == 0}, nil
}

func (e *EqExpr) MarshalJSON() ([]byte, error) {
	o1Data, err := e.O1.MarshalJSON()
	if err != nil {
		return nil, err
	}
	o2Data, err := e.O2.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return json.Marshal(SerializedQueryFilterExpr{
		Type: ExprTypeEq,
		O1:   o1Data,
		O2:   o2Data,
	})
}

func (e *EqExpr) UnmarshalJSON(data []byte) error {
	var s SerializedQueryFilterExpr
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s.Type != ExprTypeEq {
		return fmt.Errorf("expected EQ expression, got %s", s.Type)
	}
	if s.O1 == nil || s.O2 == nil {
		return fmt.Errorf("missing operands for EQ expression")
	}

	o1, err := UnmarshalQueryFilterExpr(s.O1)
	if err != nil {
		return fmt.Errorf("failed to unmarshal left operand: %v", err)
	}
	o2, err := UnmarshalQueryFilterExpr(s.O2)
	if err != nil {
		return fmt.Errorf("failed to unmarshal right operand: %v", err)
	}

	e.O1 = o1
	e.O2 = o2
	return nil
}
