package query_filter_expr

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// NotExpr 逻辑非
type NotExpr struct {
	Expr QueryFilterExpr
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
	exprData, err := e.Expr.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return json.Marshal(SerializedQueryFilterExpr{
		Type: ExprTypeNot,
		O1:   exprData,
	})
}

func (e *NotExpr) UnmarshalJSON(data []byte) error {
	var s SerializedQueryFilterExpr
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s.Type != ExprTypeNot {
		return fmt.Errorf("expected NOT expression, got %s", s.Type)
	}
	if s.O1 == nil {
		return fmt.Errorf("missing operand for NOT expression")
	}

	expr, err := UnmarshalQueryFilterExpr(s.O1)
	if err != nil {
		return fmt.Errorf("failed to unmarshal operand: %v", err)
	}

	e.Expr = expr
	return nil
}
