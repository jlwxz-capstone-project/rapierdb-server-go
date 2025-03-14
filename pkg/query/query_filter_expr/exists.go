package query_filter_expr

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// ExistsExpr 检查字段是否存在
type ExistsExpr struct {
	Field QueryFilterExpr
}

func (e *ExistsExpr) DebugPrint() string {
	return fmt.Sprintf("ExistsExpr{Field: %s}", e.Field.DebugPrint())
}

func (e *ExistsExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	// 获取字段路径
	pathExpr, ok := e.Field.(*FieldValueExpr)
	if !ok {
		return nil, fmt.Errorf("%w: EXISTS operator requires a field path", ErrSyntaxError)
	}

	// 检查路径是否为空
	if pathExpr.Path == "" {
		return nil, fmt.Errorf("%w: empty path", ErrFieldError)
	}

	// 直接检查字段是否存在
	valueOrContainer := doc.GetByPath(pathExpr.Path)
	return &ValueExpr{Value: valueOrContainer != nil}, nil
}

func (e *ExistsExpr) MarshalJSON() ([]byte, error) {
	fieldData, err := e.Field.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return json.Marshal(SerializedQueryFilterExpr{
		Type: ExprTypeExists,
		O1:   fieldData,
	})
}

func (e *ExistsExpr) UnmarshalJSON(data []byte) error {
	var s SerializedQueryFilterExpr
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s.Type != ExprTypeExists {
		return fmt.Errorf("expected EXISTS expression, got %s", s.Type)
	}
	if s.O1 == nil {
		return fmt.Errorf("missing field for EXISTS expression")
	}

	field, err := UnmarshalQueryFilterExpr(s.O1)
	if err != nil {
		return fmt.Errorf("failed to unmarshal field: %v", err)
	}

	e.Field = field
	return nil
}
