package query_filter_expr

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// FieldValueExpr 表示文档中指定路径的值
type FieldValueExpr struct {
	Path string
}

func (e *FieldValueExpr) DebugPrint() string {
	return fmt.Sprintf("FieldValueExpr{Path: %s}", e.Path)
}

func (e *FieldValueExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	if e.Path == "" {
		return nil, fmt.Errorf("%w: empty path", ErrFieldError)
	}
	valueOrContainer := doc.GetByPath(e.Path)
	if valueOrContainer == nil {
		return nil, fmt.Errorf("%w: path=%s", ErrFieldError, e.Path)
	}
	goValue, err := valueOrContainer.ToGoObject()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to convert value: %v", ErrEvalError, err)
	}
	return &ValueExpr{Value: goValue}, nil
}

func (e *FieldValueExpr) MarshalJSON() ([]byte, error) {
	return json.Marshal(SerializedQueryFilterExpr{
		Type: ExprTypeFieldValue,
		Path: e.Path,
	})
}

func (e *FieldValueExpr) UnmarshalJSON(data []byte) error {
	var s SerializedQueryFilterExpr
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s.Type != ExprTypeFieldValue {
		return fmt.Errorf("expected field value expression, got %s", s.Type)
	}
	e.Path = s.Path
	return nil
}
