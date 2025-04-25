package query_filter_expr

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js_value"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// ValueExpr 表示一个值
type ValueExpr struct {
	Type  QueryFilterExprType `json:"type"`
	Value js_value.JsValue    `json:"value"`
}

func NewValueExpr(value any) *ValueExpr {
	queryValue, err := js_value.ToJsValue(value)
	if err != nil {
		panic(err)
	}

	return &ValueExpr{
		Type:  ExprTypeValue,
		Value: queryValue,
	}
}

func (e *ValueExpr) DebugPrint() string {
	return fmt.Sprintf("ValueExpr{Value: %v}", e.Value)
}

func (e *ValueExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	return e, nil
}

func (e *ValueExpr) IsBool() bool    { _, ok := e.Value.(bool); return ok }
func (e *ValueExpr) IsString() bool  { _, ok := e.Value.(string); return ok }
func (e *ValueExpr) IsFloat64() bool { _, ok := e.Value.(float64); return ok }
func (e *ValueExpr) IsArray() bool   { _, ok := e.Value.([]any); return ok }
func (e *ValueExpr) IsMap() bool     { _, ok := e.Value.(map[string]any); return ok }
func (e *ValueExpr) IsNil() bool     { return e.Value == nil }
func (e *ValueExpr) IsNumber() bool  { return e.IsFloat64() }

func must[T any](val T, ok bool) T {
	if !ok {
		panic(fmt.Errorf("unexpected conversion"))
	}
	return val
}

func (e *ValueExpr) AsString() string      { val, ok := e.Value.(string); return must(val, ok) }
func (e *ValueExpr) AsFloat64() float64    { val, ok := e.Value.(float64); return must(val, ok) }
func (e *ValueExpr) AsBool() bool          { val, ok := e.Value.(bool); return must(val, ok) }
func (e *ValueExpr) AsArray() []any        { val, ok := e.Value.([]any); return must(val, ok) }
func (e *ValueExpr) AsMap() map[string]any { val, ok := e.Value.(map[string]any); return must(val, ok) }

func (e *ValueExpr) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func newValueExprFromJson(msg json.RawMessage) (*ValueExpr, error) {
	var temp struct {
		Type  QueryFilterExprType `json:"type"`
		Value any                 `json:"value"`
	}

	if err := json.Unmarshal(msg, &temp); err != nil {
		return nil, err
	}

	return NewValueExpr(temp.Value), nil
}
