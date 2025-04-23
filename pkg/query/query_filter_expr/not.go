package query_filter_expr

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// NotExpr 逻辑非
type NotExpr struct {
	Type QueryFilterExprType `json:"type"`
	Expr QueryFilterExpr     `json:"expr"`
}

func NewNotExpr(expr QueryFilterExpr) *NotExpr {
	return &NotExpr{
		Type: ExprTypeNot,
		Expr: expr,
	}
}

func (e *NotExpr) DebugPrint() string {
	return fmt.Sprintf("NotExpr{Expr: %s}", e.Expr.DebugPrint())
}

func (e *NotExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	result, err := e.Expr.Eval(doc)
	if err != nil {
		if errors.Is(err, ErrTypeError) {
			return nil, fmt.Errorf("%w: expected boolean in NOT expression", err)
		}
		return nil, fmt.Errorf("%w: evaluating NOT expression: %v", ErrEvalError, err)
	}
	if v, ok := result.Value.(bool); !ok {
		return nil, fmt.Errorf("%w: expected boolean in NOT expression, got %T", ErrTypeError, result.Value)
	} else {
		return &ValueExpr{Value: !v}, nil
	}
}

func (e *NotExpr) MarshalJSON() ([]byte, error) {
	return json.Marshal(e)
}
