package query_filter_expr

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// NinExpr 不包含比较
type NinExpr struct {
	O1 QueryFilterExpr
	O2 []QueryFilterExpr
}

func (e *NinExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	o1, err := e.O1.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating target of NIN: %v", ErrEvalError, err)
	}
	for _, elem := range e.O2 {
		elemValue, err := elem.Eval(doc)
		if err != nil {
			return nil, fmt.Errorf("%w: evaluating element of NIN: %v", ErrEvalError, err)
		}
		cmp, err := util.CompareValues(o1.Value, elemValue.Value)
		if err != nil {
			return nil, err
		}
		if cmp == 0 {
			return &ValueExpr{Value: false}, nil
		}
	}
	return &ValueExpr{Value: true}, nil
}

func (e *NinExpr) MarshalJSON() ([]byte, error) {
	o1Data, err := e.O1.MarshalJSON()
	if err != nil {
		return nil, err
	}
	list := make([]json.RawMessage, len(e.O2))
	for i, expr := range e.O2 {
		data, err := expr.MarshalJSON()
		if err != nil {
			return nil, err
		}
		list[i] = data
	}
	return json.Marshal(SerializedQueryFilterExpr{
		Type: ExprTypeNin,
		O1:   o1Data,
		List: list,
	})
}

func (e *NinExpr) UnmarshalJSON(data []byte) error {
	var s SerializedQueryFilterExpr
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s.Type != ExprTypeNin {
		return fmt.Errorf("expected NIN expression, got %s", s.Type)
	}
	if s.O1 == nil || len(s.List) == 0 {
		return fmt.Errorf("missing operands for NIN expression")
	}

	o1, err := UnmarshalQueryFilterExpr(s.O1)
	if err != nil {
		return fmt.Errorf("failed to unmarshal target: %v", err)
	}

	list := make([]QueryFilterExpr, len(s.List))
	for i, item := range s.List {
		expr, err := UnmarshalQueryFilterExpr(item)
		if err != nil {
			return fmt.Errorf("failed to unmarshal list item %d: %v", i, err)
		}
		list[i] = expr
	}

	e.O1 = o1
	e.O2 = list
	return nil
}
