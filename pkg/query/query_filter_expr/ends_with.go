package query_filter_expr

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// EndsWithExpr 检查字符串是否以指定后缀结束
type EndsWithExpr struct {
	Field  QueryFilterExpr
	Suffix QueryFilterExpr
}

func (e *EndsWithExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	// 评估字段表达式
	field, err := e.Field.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating field in ENDS_WITH: %v", ErrEvalError, err)
	}

	// 检查字段是否为字符串
	str, ok := field.Value.(string)
	if !ok {
		return nil, fmt.Errorf("%w: expected string in ENDS_WITH expression, got %T", ErrTypeError, field.Value)
	}

	// 评估后缀表达式
	suffix, err := e.Suffix.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating suffix in ENDS_WITH: %v", ErrEvalError, err)
	}

	// 检查后缀是否为字符串
	suffixStr, ok := suffix.Value.(string)
	if !ok {
		return nil, fmt.Errorf("%w: expected string for suffix in ENDS_WITH expression, got %T", ErrTypeError, suffix.Value)
	}

	return &ValueExpr{Value: strings.HasSuffix(str, suffixStr)}, nil
}

func (e *EndsWithExpr) MarshalJSON() ([]byte, error) {
	fieldData, err := e.Field.MarshalJSON()
	if err != nil {
		return nil, err
	}

	suffixData, err := e.Suffix.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return json.Marshal(SerializedQueryFilterExpr{
		Type: ExprTypeEndsWith,
		O1:   fieldData,
		O2:   suffixData,
	})
}

func (e *EndsWithExpr) UnmarshalJSON(data []byte) error {
	var s SerializedQueryFilterExpr
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s.Type != ExprTypeEndsWith {
		return fmt.Errorf("expected ENDS_WITH expression, got %s", s.Type)
	}
	if s.O1 == nil || s.O2 == nil {
		return fmt.Errorf("missing field or suffix for ENDS_WITH expression")
	}

	field, err := UnmarshalQueryFilterExpr(s.O1)
	if err != nil {
		return fmt.Errorf("failed to unmarshal field: %v", err)
	}

	suffix, err := UnmarshalQueryFilterExpr(s.O2)
	if err != nil {
		return fmt.Errorf("failed to unmarshal suffix: %v", err)
	}

	e.Field = field
	e.Suffix = suffix
	return nil
}
