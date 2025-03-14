package query_filter_expr

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// QueryFilterExprType 表示表达式类型的枚举
type QueryFilterExprType string

const (
	ExprTypeValue      QueryFilterExprType = "value"    // 值表达式
	ExprTypeFieldValue QueryFilterExprType = "field"    // 字段值表达式
	ExprTypeEq         QueryFilterExprType = "eq"       // 相等比较
	ExprTypeNe         QueryFilterExprType = "ne"       // 不等比较
	ExprTypeGt         QueryFilterExprType = "gt"       // 大于比较
	ExprTypeGte        QueryFilterExprType = "gte"      // 大于等于比较
	ExprTypeLt         QueryFilterExprType = "lt"       // 小于比较
	ExprTypeLte        QueryFilterExprType = "lte"      // 小于等于比较
	ExprTypeIn         QueryFilterExprType = "in"       // 包含比较
	ExprTypeNin        QueryFilterExprType = "nin"      // 不包含比较
	ExprTypeAnd        QueryFilterExprType = "and"      // 逻辑与
	ExprTypeOr         QueryFilterExprType = "or"       // 逻辑或
	ExprTypeNot        QueryFilterExprType = "not"      // 逻辑非
	ExprTypeRegex      QueryFilterExprType = "regex"    // 正则表达式匹配
	ExprTypeExists     QueryFilterExprType = "exists"   // 字段存在检查
	ExprTypeAll        QueryFilterExprType = "all"      // 数组包含所有元素
	ExprTypeSize       QueryFilterExprType = "size"     // 数组长度检查
	ExprTypeStartsWith QueryFilterExprType = "starts"   // 字符串前缀检查
	ExprTypeEndsWith   QueryFilterExprType = "ends"     // 字符串后缀检查
	ExprTypeContains   QueryFilterExprType = "contains" // 字符串包含检查
	// ExprTypeType       ExprType = "type"     // 类型检查
)

// QueryFilterExpr 定义查询表达式接口
type QueryFilterExpr interface {
	log.DebugPrintable
	Eval(doc *loro.LoroDoc) (*ValueExpr, error)
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
}

// SerializedQueryFilterExpr 表示序列化后的表达式
type SerializedQueryFilterExpr struct {
	Type  QueryFilterExprType `json:"type"`            // 表达式类型
	O1    json.RawMessage     `json:"o1,omitempty"`    // 第一个操作数
	O2    json.RawMessage     `json:"o2,omitempty"`    // 第二个操作数
	List  []json.RawMessage   `json:"list,omitempty"`  // 用于 IN/NIN/AND/OR 等需要多个操作数的表达式
	Value interface{}         `json:"value,omitempty"` // 用于 ValueExpr
	Path  string              `json:"path,omitempty"`  // 用于 FieldValueExpr
	Regex string              `json:"regex,omitempty"` // 用于 RegexExpr
}

// UnmarshalQueryFilterExpr 从 JSON 数据反序列化为 QueryExpr
func UnmarshalQueryFilterExpr(data []byte) (QueryFilterExpr, error) {
	var s SerializedQueryFilterExpr
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	var expr QueryFilterExpr
	switch s.Type {
	case ExprTypeValue:
		expr = &ValueExpr{}
	case ExprTypeFieldValue:
		expr = &FieldValueExpr{}
	case ExprTypeEq:
		expr = &EqExpr{}
	case ExprTypeNe:
		expr = &NeExpr{}
	case ExprTypeGt:
		expr = &GtExpr{}
	case ExprTypeGte:
		expr = &GteExpr{}
	case ExprTypeLt:
		expr = &LtExpr{}
	case ExprTypeLte:
		expr = &LteExpr{}
	case ExprTypeIn:
		expr = &InExpr{}
	case ExprTypeNin:
		expr = &NinExpr{}
	case ExprTypeAnd:
		expr = &AndExpr{}
	case ExprTypeOr:
		expr = &OrExpr{}
	case ExprTypeNot:
		expr = &NotExpr{}
	case ExprTypeRegex:
		expr = &RegexExpr{}
	case ExprTypeExists:
		expr = &ExistsExpr{}
	case ExprTypeAll:
		expr = &AllExpr{}
	case ExprTypeSize:
		expr = &SizeExpr{}
	case ExprTypeStartsWith:
		expr = &StartsWithExpr{}
	case ExprTypeEndsWith:
		expr = &EndsWithExpr{}
	case ExprTypeContains:
		expr = &ContainsExpr{}
	default:
		return nil, fmt.Errorf("unsupported expression type: %s", s.Type)
	}

	if err := expr.UnmarshalJSON(data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s expression: %v", s.Type, err)
	}

	return expr, nil
}
