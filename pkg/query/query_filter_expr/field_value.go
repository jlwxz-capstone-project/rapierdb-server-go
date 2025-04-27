package query_filter_expr

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js_value"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/doc_visitor"
	pe "github.com/pkg/errors"
)

// FieldValueExpr 表示文档中指定路径的值
type FieldValueExpr struct {
	Type QueryFilterExprType `json:"type"`
	Path QueryFilterExpr     `json:"path"`
}

func NewFieldValueExpr(path QueryFilterExpr) *FieldValueExpr {
	return &FieldValueExpr{
		Type: ExprTypeFieldValue,
		Path: path,
	}
}

func (e *FieldValueExpr) DebugSprint() string {
	return fmt.Sprintf("FieldValueExpr{Path: %s}", e.Path)
}

func (e *FieldValueExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	pathExpr, err := e.Path.Eval(doc)
	if err != nil {
		return nil, pe.Wrapf(ErrEvalError, "evaluating path in FIELD_VALUE: %v", err)
	}

	if !pathExpr.IsString() {
		return nil, pe.Wrapf(ErrTypeError, "expected string path in FIELD_VALUE expression, got %T", pathExpr.Value)
	}

	path := pathExpr.AsString()
	if !isValidPath(path) {
		return nil, pe.Wrapf(ErrTypeError, "invalid path in FIELD_VALUE expression")
	}

	val, err := doc_visitor.VisitDocByPath(doc, path)
	if err != nil {
		return nil, pe.Wrapf(ErrFieldError, "path=%s", path)
	}

	goValue, err := js_value.ToJsValue(val)
	if err != nil {
		return nil, pe.Wrapf(ErrEvalError, "failed to convert value: %v", err)
	}
	return NewValueExpr(goValue), nil
}

func (e *FieldValueExpr) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func newFieldValueExprFromJson(msg json.RawMessage) (*FieldValueExpr, error) {
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

	return NewFieldValueExpr(path), nil
}
