package query_filter_expr

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
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

func (e *FieldValueExpr) DebugPrint() string {
	return fmt.Sprintf("FieldValueExpr{Path: %s}", e.Path)
}

func (e *FieldValueExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	pathExpr, err := e.Path.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating path in FIELD_VALUE: %v", ErrEvalError, err)
	}

	if !pathExpr.IsString() {
		return nil, fmt.Errorf("%w: expected string path in FIELD_VALUE expression, got %T", ErrTypeError, pathExpr.Value)
	}

	path := pathExpr.AsString()
	if !isValidPath(path) {
		return nil, fmt.Errorf("%w: invalid path in FIELD_VALUE expression", ErrTypeError)
	}

	valueOrContainer := doc.GetByPath(path)
	if valueOrContainer == nil {
		return nil, fmt.Errorf("%w: path=%s", ErrFieldError, path)
	}
	goValue, err := valueOrContainer.ToGoObject()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to convert value: %v", ErrEvalError, err)
	}
	return &ValueExpr{Value: goValue}, nil
}

func (e *FieldValueExpr) MarshalJSON() ([]byte, error) {
	return json.Marshal(e)
}
