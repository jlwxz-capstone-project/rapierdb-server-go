package query_filter_expr

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// ValueExpr 表示一个值
//
//	  type Value =
//		  | bool
//		  | float64 所有数字都是 float64
//		  | string
//		  | []Value
//		  | map[string]Value
//		  | nil
type ValueExpr struct {
	Type  QueryFilterExprType `json:"type"`
	Value any                 `json:"value"`
}

func NewValueExpr(value any) *ValueExpr {
	return &ValueExpr{
		Type:  ExprTypeValue,
		Value: normalizeValue(value),
	}
}

func normalizeValue(value any) any {
	switch v := value.(type) {
	case bool, string, float64, nil:
		return v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32:
		return util.ToFloat64(v)
	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = normalizeValue(item)
		}
		return result
	case map[string]any:
		result := make(map[string]any, len(v))
		for key, item := range v {
			result[key] = normalizeValue(item)
		}
		return result
	default:
		panic(fmt.Errorf("不支持的值类型: %T", value))
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
func (e *ValueExpr) IsInt64() bool   { _, ok := e.Value.(int64); return ok }
func (e *ValueExpr) IsFloat64() bool { _, ok := e.Value.(float64); return ok }
func (e *ValueExpr) IsArray() bool   { _, ok := e.Value.([]any); return ok }
func (e *ValueExpr) IsMap() bool     { _, ok := e.Value.(map[string]any); return ok }
func (e *ValueExpr) IsNil() bool     { return e.Value == nil }
func (e *ValueExpr) IsNumber() bool  { return e.IsFloat64() || e.IsInt64() }

func must[T any](val T, ok bool) T {
	if !ok {
		panic(fmt.Errorf("unexpected conversion"))
	}
	return val
}

func (e *ValueExpr) AsString() string      { val, ok := e.Value.(string); return must(val, ok) }
func (e *ValueExpr) AsInt64() int64        { val, ok := e.Value.(int64); return must(val, ok) }
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
