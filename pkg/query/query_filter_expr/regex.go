package query_filter_expr

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// RegexExpr 正则表达式匹配
type RegexExpr struct {
	Type  QueryFilterExprType `json:"type"`
	O1    QueryFilterExpr     `json:"o1"`
	Regex string              `json:"regex"`
}

func NewRegexExpr(o1 QueryFilterExpr, regex string) *RegexExpr {
	return &RegexExpr{
		Type:  ExprTypeRegex,
		O1:    o1,
		Regex: regex,
	}
}

func (e *RegexExpr) DebugPrint() string {
	return fmt.Sprintf("RegexExpr{O1: %s, Regex: %s}", e.O1.DebugPrint(), e.Regex)
}

func (e *RegexExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	o1, err := e.O1.Eval(doc)
	if err != nil {
		return nil, fmt.Errorf("%w: evaluating regex pattern: %v", ErrEvalError, err)
	}
	str, ok := o1.Value.(string)
	if !ok {
		return nil, fmt.Errorf("%w: expected string for regex matching, got %T", ErrTypeError, o1.Value)
	}
	matched, err := regexp.MatchString(e.Regex, str)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid regex pattern '%s': %v", ErrSyntaxError, e.Regex, err)
	}
	return &ValueExpr{Value: matched}, nil
}

func (e *RegexExpr) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func newRegexExprFromJson(msg json.RawMessage) (*RegexExpr, error) {
	var temp struct {
		Type  QueryFilterExprType `json:"type"`
		O1    json.RawMessage     `json:"o1"`
		Regex string              `json:"regex"`
	}

	if err := json.Unmarshal(msg, &temp); err != nil {
		return nil, err
	}

	o1, err := NewQueryFilterExprFromJson(temp.O1)
	if err != nil {
		return nil, err
	}

	return NewRegexExpr(o1, temp.Regex), nil
}
