package query_filter_expr

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// ContainsExpr 检查字符串是否包含指定子串
type ContainsExpr struct {
	Field  QueryFilterExpr
	Substr QueryFilterExpr
}

func (e *ContainsExpr) DebugPrint() string {
	return fmt.Sprintf("ContainsExpr{Field: %s, Substr: %s}", e.Field.DebugPrint(), e.Substr.DebugPrint())
}

func (e *ContainsExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	// 评估字段表达式
	field, err := e.Field.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating field in CONTAINS: %v", ErrEvalError, err)
	}

	// 检查字段是否为字符串
	str, ok := field.Value.(string)
	if !ok {
		return nil, fmt.Errorf("%w: expected string in CONTAINS expression, got %T", ErrTypeError, field.Value)
	}

	// 评估子串表达式
	substr, err := e.Substr.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating substring in CONTAINS: %v", ErrEvalError, err)
	}

	// 检查子串是否为字符串
	substrStr, ok := substr.Value.(string)
	if !ok {
		return nil, fmt.Errorf("%w: expected string for substring in CONTAINS expression, got %T", ErrTypeError, substr.Value)
	}

	return &ValueExpr{Value: strings.Contains(str, substrStr)}, nil
}

func (e *ContainsExpr) MarshalJSON() ([]byte, error) {
	fieldData, err := e.Field.MarshalJSON()
	if err != nil {
		return nil, err
	}

	substrData, err := e.Substr.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return json.Marshal(SerializedQueryFilterExpr{
		Type: ExprTypeContains,
		O1:   fieldData,
		O2:   substrData,
	})
}

func (e *ContainsExpr) UnmarshalJSON(data []byte) error {
	var s SerializedQueryFilterExpr
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s.Type != ExprTypeContains {
		return fmt.Errorf("expected CONTAINS expression, got %s", s.Type)
	}
	if s.O1 == nil || s.O2 == nil {
		return fmt.Errorf("missing field or substring for CONTAINS expression")
	}

	field, err := UnmarshalQueryFilterExpr(s.O1)
	if err != nil {
		return fmt.Errorf("failed to unmarshal field: %v", err)
	}

	substr, err := UnmarshalQueryFilterExpr(s.O2)
	if err != nil {
		return fmt.Errorf("failed to unmarshal substring: %v", err)
	}

	e.Field = field
	e.Substr = substr
	return nil
}
