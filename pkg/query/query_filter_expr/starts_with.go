package query_filter_expr

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// StartsWithExpr 检查字符串是否以指定前缀开始
type StartsWithExpr struct {
	Field  QueryFilterExpr
	Prefix QueryFilterExpr
}

func (e *StartsWithExpr) DebugPrint() string {
	return fmt.Sprintf("StartsWithExpr{Field: %s, Prefix: %s}", e.Field.DebugPrint(), e.Prefix.DebugPrint())
}

func (e *StartsWithExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	// 评估字段表达式
	field, err := e.Field.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating field in STARTS_WITH: %v", ErrEvalError, err)
	}

	// 检查字段是否为字符串
	str, ok := field.Value.(string)
	if !ok {
		return nil, fmt.Errorf("%w: expected string in STARTS_WITH expression, got %T", ErrTypeError, field.Value)
	}

	// 评估前缀表达式
	prefix, err := e.Prefix.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating prefix in STARTS_WITH: %v", ErrEvalError, err)
	}

	// 检查前缀是否为字符串
	prefixStr, ok := prefix.Value.(string)
	if !ok {
		return nil, fmt.Errorf("%w: expected string for prefix in STARTS_WITH expression, got %T", ErrTypeError, prefix.Value)
	}

	return &ValueExpr{Value: strings.HasPrefix(str, prefixStr)}, nil
}

func (e *StartsWithExpr) MarshalJSON() ([]byte, error) {
	fieldData, err := e.Field.MarshalJSON()
	if err != nil {
		return nil, err
	}

	prefixData, err := e.Prefix.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return json.Marshal(SerializedQueryFilterExpr{
		Type: ExprTypeStartsWith,
		O1:   fieldData,
		O2:   prefixData,
	})
}

func (e *StartsWithExpr) UnmarshalJSON(data []byte) error {
	var s SerializedQueryFilterExpr
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s.Type != ExprTypeStartsWith {
		return fmt.Errorf("expected STARTS_WITH expression, got %s", s.Type)
	}
	if s.O1 == nil || s.O2 == nil {
		return fmt.Errorf("missing field or prefix for STARTS_WITH expression")
	}

	field, err := UnmarshalQueryFilterExpr(s.O1)
	if err != nil {
		return fmt.Errorf("failed to unmarshal field: %v", err)
	}

	prefix, err := UnmarshalQueryFilterExpr(s.O2)
	if err != nil {
		return fmt.Errorf("failed to unmarshal prefix: %v", err)
	}

	e.Field = field
	e.Prefix = prefix
	return nil
}
