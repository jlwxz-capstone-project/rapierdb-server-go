package query_filter_expr

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// StartsWithExpr 检查字符串是否以指定前缀开始
type StartsWithExpr struct {
	Type   QueryFilterExprType `json:"type"`
	Target QueryFilterExpr     `json:"target"`
	Prefix QueryFilterExpr     `json:"prefix"`
}

func NewStartsWithExpr(target QueryFilterExpr, prefix QueryFilterExpr) *StartsWithExpr {
	return &StartsWithExpr{
		Type:   ExprTypeStartsWith,
		Target: target,
		Prefix: prefix,
	}
}

func (e *StartsWithExpr) DebugPrint() string {
	return fmt.Sprintf("StartsWithExpr{Target: %s, Prefix: %s}", e.Target.DebugPrint(), e.Prefix.DebugPrint())
}

func (e *StartsWithExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	// 评估字段表达式
	target, err := e.Target.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating target in STARTS_WITH: %v", ErrEvalError, err)
	}

	// 检查字段是否为字符串
	str, ok := target.Value.(string)
	if !ok {
		return nil, fmt.Errorf("%w: expected string in STARTS_WITH expression, got %T", ErrTypeError, target.Value)
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
	return json.Marshal(e)
}
