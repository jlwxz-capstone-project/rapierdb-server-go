package query_filter_expr

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// InExpr 包含比较
type InExpr struct {
	O1 QueryFilterExpr
	O2 []QueryFilterExpr
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
		cmp, err := util.CompareValues(o1.Value, elemValue.Value)
		if err != nil {
			return nil, err
		}
		if cmp == 0 {
			return &ValueExpr{Value: true}, nil
		}
	}
	return &ValueExpr{Value: false}, nil
}

func (e *InExpr) MarshalJSON() ([]byte, error) {
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
		Type: ExprTypeIn,
		O1:   o1Data,
		List: list,
	})
}

func (e *InExpr) UnmarshalJSON(data []byte) error {
	var s SerializedQueryFilterExpr
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s.Type != ExprTypeIn {
		return fmt.Errorf("expected IN expression, got %s", s.Type)
	}
	if s.O1 == nil || len(s.List) == 0 {
		return fmt.Errorf("missing operands for IN expression")
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
