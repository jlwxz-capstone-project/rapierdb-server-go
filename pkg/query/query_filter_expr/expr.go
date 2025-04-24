package query_filter_expr

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	pe "github.com/pkg/errors"
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

func isValidPath(path string) bool {
	if path == "" {
		return false
	}

	for i := 0; i < len(path)-1; i++ {
		if path[i] == '/' && path[i+1] == '/' {
			return false
		}
	}

	parts := strings.Split(path, "/")
	for _, part := range parts {
		if part != "" {
			return true
		}
	}

	return false
}

// QueryFilterExpr 定义查询表达式接口
type QueryFilterExpr interface {
	log.DebugPrintable
	Eval(doc *loro.LoroDoc) (*ValueExpr, error)
	ToJSON() ([]byte, error)
}

// NewQueryFilterExprFromJson 从 JSON 数据中解析查询表达式
func NewQueryFilterExprFromJson(data []byte) (QueryFilterExpr, error) {
	var typeInfo struct {
		Type QueryFilterExprType `json:"type"`
	}

	if err := json.Unmarshal(data, &typeInfo); err != nil {
		return nil, err
	}

	switch typeInfo.Type {
	case ExprTypeAll:
		return newAllExprFromJson(data)
	case ExprTypeAnd:
		return newAndExprFromJson(data)
	case ExprTypeContains:
		return newContainsExprFromJson(data)
	case ExprTypeEndsWith:
		return newEndsWithExprFromJson(data)
	case ExprTypeEq:
		return newEqExprFromJson(data)
	case ExprTypeExists:
		return newExistsExprFromJson(data)
	case ExprTypeFieldValue:
		return newFieldValueExprFromJson(data)
	case ExprTypeGt:
		return newGtExprFromJson(data)
	case ExprTypeIn:
		return newInExprFromJson(data)
	case ExprTypeGte:
		return newGteExprFromJson(data)
	case ExprTypeLt:
		return newLtExprFromJson(data)
	case ExprTypeLte:
		return newLteExprFromJson(data)
	case ExprTypeNe:
		return newNeExprFromJson(data)
	case ExprTypeNin:
		return newNinExprFromJson(data)
	case ExprTypeNot:
		return newNotExprFromJson(data)
	case ExprTypeOr:
		return newOrExprFromJson(data)
	case ExprTypeRegex:
		return newRegexExprFromJson(data)
	case ExprTypeSize:
		return newSizeExprFromJson(data)
	case ExprTypeStartsWith:
		return newStartsWithExprFromJson(data)
	case ExprTypeValue:
		return newValueExprFromJson(data)
	default:
		return nil, pe.WithStack(fmt.Errorf("unknown expression type: %s", typeInfo.Type))
	}
}
