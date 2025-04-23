package query_filter_expr

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// NinExpr 不包含比较
type NinExpr struct {
	Type QueryFilterExprType `json:"type"`
	O1   QueryFilterExpr     `json:"o1"`
	O2   []QueryFilterExpr   `json:"o2"`
}

func NewNinExpr(o1 QueryFilterExpr, o2 []QueryFilterExpr) *NinExpr {
	return &NinExpr{
		Type: ExprTypeNin,
		O1:   o1,
		O2:   o2,
	}
}

func (e *NinExpr) DebugPrint() string {
	o2 := make([]string, len(e.O2))
	for i, expr := range e.O2 {
		o2[i] = expr.DebugPrint()
	}
	o2Str := fmt.Sprintf("[%s]", strings.Join(o2, ", "))
	return fmt.Sprintf("NinExpr{O1: %s, O2: %s}", e.O1.DebugPrint(), o2Str)
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
	return json.Marshal(e)
}
