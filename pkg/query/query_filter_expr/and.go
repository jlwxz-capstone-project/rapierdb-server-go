package query_filter_expr

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	pe "github.com/pkg/errors"
)

// AndExpr 逻辑与
type AndExpr struct {
	Type  QueryFilterExprType `json:"type"`
	Exprs []QueryFilterExpr   `json:"exprs"`
}

func NewAndExpr(exprs []QueryFilterExpr) *AndExpr {
	return &AndExpr{
		Type:  ExprTypeAnd,
		Exprs: exprs,
	}
}

func (e *AndExpr) DebugSprint() string {
	exprs := make([]string, len(e.Exprs))
	for i, expr := range e.Exprs {
		exprs[i] = expr.DebugSprint()
	}
	return fmt.Sprintf("AndExpr{Exprs: %s}", strings.Join(exprs, " && "))
}

func (e *AndExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	for _, expr := range e.Exprs {
		result, err := expr.Eval(doc)
		if err != nil {
			if pe.Is(err, ErrTypeError) {
				return nil, pe.Wrapf(ErrTypeError, "expected boolean in AND expression: %v", err)
			}
			return nil, pe.Wrapf(ErrEvalError, "evaluating AND expression: %v", err)
		}
		if v, ok := result.Value.(bool); !ok {
			return nil, pe.Wrapf(ErrTypeError, "expected boolean in AND expression, got %T", result.Value)
		} else if !v {
			return NewValueExpr(false), nil
		}
	}
	return NewValueExpr(true), nil
}

func (e *AndExpr) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func newAndExprFromJson(msg json.RawMessage) (*AndExpr, error) {
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
	return NewAndExpr(exprs), nil
}
