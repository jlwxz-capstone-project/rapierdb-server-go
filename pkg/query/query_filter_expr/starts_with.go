package query_filter_expr

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	pe "github.com/pkg/errors"
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
		return nil, pe.Wrapf(ErrEvalError, "evaluating target in STARTS_WITH: %v", err)
	}

	// 检查字段是否为字符串
	str, ok := target.Value.(string)
	if !ok {
		return nil, pe.Wrapf(ErrTypeError, "expected string in STARTS_WITH expression, got %T", target.Value)
	}

	// 评估前缀表达式
	prefix, err := e.Prefix.Eval(doc)
	if err != nil {
		return nil, pe.Wrapf(ErrEvalError, "evaluating prefix in STARTS_WITH: %v", err)
	}

	// 检查前缀是否为字符串
	prefixStr, ok := prefix.Value.(string)
	if !ok {
		return nil, pe.Wrapf(ErrTypeError, "expected string for prefix in STARTS_WITH expression, got %T", prefix.Value)
	}

	return NewValueExpr(strings.HasPrefix(str, prefixStr)), nil
}

func (e *StartsWithExpr) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func newStartsWithExprFromJson(msg json.RawMessage) (*StartsWithExpr, error) {
	var temp struct {
		Type   QueryFilterExprType `json:"type"`
		Target json.RawMessage     `json:"target"`
		Prefix json.RawMessage     `json:"prefix"`
	}

	if err := json.Unmarshal(msg, &temp); err != nil {
		return nil, err
	}

	target, err := NewQueryFilterExprFromJson(temp.Target)
	if err != nil {
		return nil, err
	}

	prefix, err := NewQueryFilterExprFromJson(temp.Prefix)
	if err != nil {
		return nil, err
	}

	return NewStartsWithExpr(target, prefix), nil
}
