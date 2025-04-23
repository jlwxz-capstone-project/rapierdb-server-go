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
	return json.Marshal(e)
}
