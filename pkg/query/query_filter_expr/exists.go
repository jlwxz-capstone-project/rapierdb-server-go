package query_filter_expr

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// ExistsExpr 检查字段是否存在
type ExistsExpr struct {
	Type QueryFilterExprType `json:"type"`
	Path QueryFilterExpr     `json:"path"`
}

func NewExistsExpr(path QueryFilterExpr) *ExistsExpr {
	return &ExistsExpr{
		Type: ExprTypeExists,
		Path: path,
	}
}

func (e *ExistsExpr) DebugPrint() string {
	return fmt.Sprintf("ExistsExpr{Path: %s}", e.Path.DebugPrint())
}

func (e *ExistsExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	// 获取字段路径
	pathExpr, err := e.Path.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating path in EXISTS: %v", ErrEvalError, err)
	}

	if !pathExpr.IsString() {
		return nil, fmt.Errorf("%w: expected string path in EXISTS expression, got %T", ErrTypeError, pathExpr.Value)
	}

	path := pathExpr.AsString()
	if !isValidPath(path) {
		return nil, fmt.Errorf("%w: invalid path in EXISTS expression", ErrTypeError)
	}

	// 直接检查字段是否存在
	valueOrContainer := doc.GetByPath(path)
	return &ValueExpr{Value: valueOrContainer != nil}, nil
}

func (e *ExistsExpr) MarshalJSON() ([]byte, error) {
	return json.Marshal(e)
}
