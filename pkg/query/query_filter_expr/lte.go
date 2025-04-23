package query_filter_expr

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
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

func (e *LteExpr) DebugPrint() string {
	return fmt.Sprintf("LteExpr{O1: %s, O2: %s}", e.O1.DebugPrint(), e.O2.DebugPrint())
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
	cmp, err := util.CompareValues(o1.Value, o2.Value)
	if err != nil {
		return nil, err
	}
	return &ValueExpr{Value: cmp <= 0}, nil
}

func (e *LteExpr) MarshalJSON() ([]byte, error) {
	return json.Marshal(e)
}
