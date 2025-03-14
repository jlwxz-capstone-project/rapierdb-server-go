package query_filter_expr

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// RegexExpr 正则表达式匹配
type RegexExpr struct {
	O1    QueryFilterExpr
	Regex string
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

func (e *RegexExpr) MarshalJSON() ([]byte, error) {
	o1Data, err := e.O1.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return json.Marshal(SerializedQueryFilterExpr{
		Type:  ExprTypeRegex,
		O1:    o1Data,
		Regex: e.Regex,
	})
}

func (e *RegexExpr) UnmarshalJSON(data []byte) error {
	var s SerializedQueryFilterExpr
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s.Type != ExprTypeRegex {
		return fmt.Errorf("expected REGEX expression, got %s", s.Type)
	}
	if s.O1 == nil || s.Regex == "" {
		return fmt.Errorf("missing operand or pattern for REGEX expression")
	}

	o1, err := UnmarshalQueryFilterExpr(s.O1)
	if err != nil {
		return fmt.Errorf("failed to unmarshal operand: %v", err)
	}

	// 验证正则表达式的有效性
	if _, err := regexp.Compile(s.Regex); err != nil {
		return fmt.Errorf("invalid regex pattern '%s': %v", s.Regex, err)
	}

	e.O1 = o1
	e.Regex = s.Regex
	return nil
}
