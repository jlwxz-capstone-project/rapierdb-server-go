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
	Type  QueryFilterExprType `json:"type"`
	Exprs []QueryFilterExpr   `json:"exprs"`
}

func NewOrExpr(exprs []QueryFilterExpr) *OrExpr {
	return &OrExpr{
		Type:  ExprTypeOr,
		Exprs: exprs,
	}
}

func (e *OrExpr) DebugSprint() string {
	exprs := make([]string, len(e.Exprs))
	for i, expr := range e.Exprs {
		exprs[i] = expr.DebugSprint()
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
			return NewValueExpr(true), nil
		}
	}
	return NewValueExpr(false), nil
}

func (e *OrExpr) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func newOrExprFromJson(msg json.RawMessage) (*OrExpr, error) {
	var temp struct {
		Type  QueryFilterExprType `json:"type"`
		Exprs []json.RawMessage   `json:"exprs"`
	}

	if err := json.Unmarshal(msg, &temp); err != nil {
		return nil, err
	}

	exprs := make([]QueryFilterExpr, len(temp.Exprs))
	for i, expr := range temp.Exprs {
		expr, err := NewQueryFilterExprFromJson(expr)
		if err != nil {
			return nil, err
		}
		exprs[i] = expr
	}

	return NewOrExpr(exprs), nil
}
