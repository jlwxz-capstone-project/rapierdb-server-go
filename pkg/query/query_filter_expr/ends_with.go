package query_filter_expr

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// EndsWithExpr 检查字符串是否以指定后缀结束
type EndsWithExpr struct {
	Type   QueryFilterExprType `json:"type"`
	Target QueryFilterExpr     `json:"target"`
	Suffix QueryFilterExpr     `json:"suffix"`
}

func NewEndsWithExpr(target QueryFilterExpr, suffix QueryFilterExpr) *EndsWithExpr {
	return &EndsWithExpr{
		Type:   ExprTypeEndsWith,
		Target: target,
		Suffix: suffix,
	}
}

func (e *EndsWithExpr) DebugPrint() string {
	return fmt.Sprintf("EndsWithExpr{Target: %s, Suffix: %s}", e.Target.DebugPrint(), e.Suffix.DebugPrint())
}

func (e *EndsWithExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	// 评估字段表达式
	target, err := e.Target.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating target in ENDS_WITH: %v", ErrEvalError, err)
	}

	// 检查字段是否为字符串
	str, ok := target.Value.(string)
	if !ok {
		return nil, fmt.Errorf("%w: expected string in ENDS_WITH expression, got %T", ErrTypeError, target.Value)
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

func (e *EndsWithExpr) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func newEndsWithExprFromJson(msg json.RawMessage) (*EndsWithExpr, error) {
	var temp struct {
		Type   QueryFilterExprType `json:"type"`
		Target json.RawMessage     `json:"target"`
		Suffix json.RawMessage     `json:"suffix"`
	}

	if err := json.Unmarshal(msg, &temp); err != nil {
		return nil, err
	}

	target, err := NewQueryFilterExprFromJson(temp.Target)
	if err != nil {
		return nil, err
	}

	suffix, err := NewQueryFilterExprFromJson(temp.Suffix)
	if err != nil {
		return nil, err
	}

	return NewEndsWithExpr(target, suffix), nil
}
