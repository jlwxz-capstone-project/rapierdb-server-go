package query_filter_expr

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// OrExpr 逻辑或
type OrExpr struct {
	Exprs []QueryFilterExpr
}

func (e *OrExpr) DebugPrint() string {
	exprs := make([]string, len(e.Exprs))
	for i, expr := range e.Exprs {
		exprs[i] = expr.DebugPrint()
	}
	return fmt.Sprintf("OrExpr{Exprs: %s}", strings.Join(exprs, " || "))
}

func (e *OrExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	for _, expr := range e.Exprs {
		result, err := expr.Eval(doc)
		if err != nil {
			if errors.Is(err, ErrTypeError) {
				return nil, fmt.Errorf("%w: expected boolean in OR expression", err)
			}
			return nil, fmt.Errorf("%w: evaluating OR expression: %v", ErrEvalError, err)
		}
		if v, ok := result.Value.(bool); !ok {
			return nil, fmt.Errorf("%w: expected boolean in OR expression, got %T", ErrTypeError, result.Value)
		} else if v {
			return &ValueExpr{Value: true}, nil
		}
	}
	return &ValueExpr{Value: false}, nil
}

func (e *OrExpr) MarshalJSON() ([]byte, error) {
	list := make([]json.RawMessage, len(e.Exprs))
	for i, expr := range e.Exprs {
		data, err := expr.MarshalJSON()
		if err != nil {
			return nil, err
		}
		list[i] = data
	}
	return json.Marshal(SerializedQueryFilterExpr{
		Type: ExprTypeOr,
		List: list,
	})
}

func (e *OrExpr) UnmarshalJSON(data []byte) error {
	var s SerializedQueryFilterExpr
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s.Type != ExprTypeOr {
		return fmt.Errorf("expected OR expression, got %s", s.Type)
	}
	if len(s.List) == 0 {
		return fmt.Errorf("missing operands for OR expression")
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
