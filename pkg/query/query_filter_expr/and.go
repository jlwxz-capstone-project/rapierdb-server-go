package query_filter_expr

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// AndExpr 逻辑与
type AndExpr struct {
	Exprs []QueryFilterExpr
}

func (e *AndExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	for _, expr := range e.Exprs {
		result, err := expr.Eval(doc)
		if err != nil {
			if errors.Is(err, ErrTypeError) {
				return nil, fmt.Errorf("%w: expected boolean in AND expression", err)
			}
			return nil, fmt.Errorf("%w: evaluating AND expression: %v", ErrEvalError, err)
		}
		if v, ok := result.Value.(bool); !ok {
			return nil, fmt.Errorf("%w: expected boolean in AND expression, got %T", ErrTypeError, result.Value)
		} else if !v {
			return &ValueExpr{Value: false}, nil
		}
	}
	return &ValueExpr{Value: true}, nil
}

func (e *AndExpr) MarshalJSON() ([]byte, error) {
	list := make([]json.RawMessage, len(e.Exprs))
	for i, expr := range e.Exprs {
		data, err := expr.MarshalJSON()
		if err != nil {
			return nil, err
		}
		list[i] = data
	}
	return json.Marshal(SerializedQueryFilterExpr{
		Type: ExprTypeAnd,
		List: list,
	})
}

func (e *AndExpr) UnmarshalJSON(data []byte) error {
	var s SerializedQueryFilterExpr
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s.Type != ExprTypeAnd {
		return fmt.Errorf("expected AND expression, got %s", s.Type)
	}
	if len(s.List) == 0 {
		return fmt.Errorf("missing operands for AND expression")
	}

	exprs := make([]QueryFilterExpr, len(s.List))
	for i, item := range s.List {
		expr, err := UnmarshalQueryFilterExpr(item)
		if err != nil {
			return fmt.Errorf("failed to unmarshal expression %d: %v", i, err)
		}
		exprs[i] = expr
	}

	e.Exprs = exprs
	return nil
}
