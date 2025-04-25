package query_filter_expr

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/doc_visitor"
	pe "github.com/pkg/errors"
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
		return nil, pe.Wrapf(ErrEvalError, "evaluating path in EXISTS: %v", err)
	}

	if !pathExpr.IsString() {
		return nil, pe.Wrapf(ErrTypeError, "expected string path in EXISTS expression, got %T", pathExpr.Value)
	}

	path := pathExpr.AsString()
	if !isValidPath(path) {
		return nil, pe.Wrapf(ErrTypeError, "invalid path in EXISTS expression: %s", path)
	}

	_, err = doc_visitor.VisitDocByPath(doc, path)
	notFound := err != nil && errors.Is(err, doc_visitor.PathNotFoundError)
	return NewValueExpr(!notFound), nil
}

func (e *ExistsExpr) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func newExistsExprFromJson(msg json.RawMessage) (*ExistsExpr, error) {
	var temp struct {
		Type QueryFilterExprType `json:"type"`
		Path json.RawMessage     `json:"path"`
	}

	if err := json.Unmarshal(msg, &temp); err != nil {
		return nil, err
	}

	path, err := NewQueryFilterExprFromJson(temp.Path)
	if err != nil {
		return nil, err
	}

	return NewExistsExpr(path), nil
}
