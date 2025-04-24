package query_filter_expr

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// ContainsExpr 检查字符串是否包含指定子串
type ContainsExpr struct {
	Type   QueryFilterExprType `json:"type"`
	Target QueryFilterExpr     `json:"target"`
	Substr QueryFilterExpr     `json:"substr"`
}

func NewContainsExpr(target QueryFilterExpr, substr QueryFilterExpr) *ContainsExpr {
	return &ContainsExpr{
		Type:   ExprTypeContains,
		Target: target,
		Substr: substr,
	}
}

func (e *ContainsExpr) DebugPrint() string {
	return fmt.Sprintf("ContainsExpr{Target: %s, Substr: %s}", e.Target.DebugPrint(), e.Substr.DebugPrint())
}

func (e *ContainsExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	// 评估字段表达式
	target, err := e.Target.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating target in CONTAINS: %v", ErrEvalError, err)
	}

	// 检查字段是否为字符串
	str, ok := target.Value.(string)
	if !ok {
		return nil, fmt.Errorf("%w: expected string in CONTAINS expression, got %T", ErrTypeError, target.Value)
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

func (e *ContainsExpr) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func newContainsExprFromJson(msg json.RawMessage) (*ContainsExpr, error) {
	var temp struct {
		Type   QueryFilterExprType `json:"type"`
		Target json.RawMessage     `json:"target"`
		Substr json.RawMessage     `json:"substr"`
	}

	if err := json.Unmarshal(msg, &temp); err != nil {
		return nil, err
	}

	target, err := NewQueryFilterExprFromJson(temp.Target)
	if err != nil {
		return nil, err
	}

	substr, err := NewQueryFilterExprFromJson(temp.Substr)
	if err != nil {
		return nil, err
	}

	return NewContainsExpr(target, substr), nil
}
