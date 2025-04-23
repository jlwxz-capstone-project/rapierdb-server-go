package query_filter_expr

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// GtExpr 大于比较
type GtExpr struct {
	Type QueryFilterExprType `json:"type"`
	O1   QueryFilterExpr     `json:"o1"`
	O2   QueryFilterExpr     `json:"o2"`
}

func NewGtExpr(o1 QueryFilterExpr, o2 QueryFilterExpr) *GtExpr {
	return &GtExpr{
		Type: ExprTypeGt,
		O1:   o1,
		O2:   o2,
	}
}

func (e *GtExpr) DebugPrint() string {
	return fmt.Sprintf("GtExpr{O1: %s, O2: %s}", e.O1.DebugPrint(), e.O2.DebugPrint())
}

func (e *GtExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	o1, err := e.O1.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating left operand of GT: %v", ErrEvalError, err)
	}
	o2, err := e.O2.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating right operand of GT: %v", ErrEvalError, err)
	}
	cmp, err := util.CompareValues(o1.Value, o2.Value)
	if err != nil {
		return nil, err
	}
	return &ValueExpr{Value: cmp > 0}, nil
}

func (e *GtExpr) MarshalJSON() ([]byte, error) {
	return json.Marshal(e)
}
