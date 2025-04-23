package query_filter_expr

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/pkg/errors"
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
	MarshalJSON() ([]byte, error)
}

func UnmarshalQueryFilterExpr(data []byte) (QueryFilterExpr, error) {
	var rawFields map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawFields); err != nil {
		return nil, pe.WithStack(errors.Wrapf(err, "error pre-unmarshalling into raw fields"))
	}

	typeBytes, ok := rawFields["type"]
	if !ok {
		return nil, pe.WithStack(fmt.Errorf("missing type field"))
	}

	var exprType QueryFilterExprType
	if err := json.Unmarshal(typeBytes, &exprType); err != nil {
		return nil, pe.WithStack(errors.Wrapf(err, "error unmarshalling type field"))
	}

	switch exprType {
	case ExprTypeAll:
		var allExpr AllExpr
		if err := json.Unmarshal(data, &allExpr); err != nil {
			return nil, pe.WithStack(errors.Wrapf(err, "error unmarshalling all expression"))
		}
		return &allExpr, nil
	case ExprTypeAnd:
		var andExpr AndExpr
		if err := json.Unmarshal(data, &andExpr); err != nil {
			return nil, pe.WithStack(errors.Wrapf(err, "error unmarshalling and expression"))
		}
		return &andExpr, nil
	case ExprTypeContains:
		var containsExpr ContainsExpr
		if err := json.Unmarshal(data, &containsExpr); err != nil {
			return nil, pe.WithStack(errors.Wrapf(err, "error unmarshalling contains expression"))
		}
		return &containsExpr, nil
	case ExprTypeEndsWith:
		var endsWithExpr EndsWithExpr
		if err := json.Unmarshal(data, &endsWithExpr); err != nil {
			return nil, pe.WithStack(errors.Wrapf(err, "error unmarshalling ends with expression"))
		}
		return &endsWithExpr, nil
	case ExprTypeEq:
		var eqExpr EqExpr
		if err := json.Unmarshal(data, &eqExpr); err != nil {
			return nil, pe.WithStack(errors.Wrapf(err, "error unmarshalling eq expression"))
		}
		return &eqExpr, nil
	case ExprTypeExists:
		var existsExpr ExistsExpr
		if err := json.Unmarshal(data, &existsExpr); err != nil {
			return nil, pe.WithStack(errors.Wrapf(err, "error unmarshalling exists expression"))
		}
		return &existsExpr, nil
	case ExprTypeFieldValue:
		var fieldValueExpr FieldValueExpr
		if err := json.Unmarshal(data, &fieldValueExpr); err != nil {
			return nil, pe.WithStack(errors.Wrapf(err, "error unmarshalling field value expression"))
		}
		return &fieldValueExpr, nil
	case ExprTypeGt:
		var gtExpr GtExpr
		if err := json.Unmarshal(data, &gtExpr); err != nil {
			return nil, pe.WithStack(errors.Wrapf(err, "error unmarshalling gt expression"))
		}
		return &gtExpr, nil
	case ExprTypeGte:
		var gteExpr GteExpr
		if err := json.Unmarshal(data, &gteExpr); err != nil {
			return nil, pe.WithStack(errors.Wrapf(err, "error unmarshalling gte expression"))
		}
		return &gteExpr, nil
	case ExprTypeIn:
		var inExpr InExpr
		if err := json.Unmarshal(data, &inExpr); err != nil {
			return nil, pe.WithStack(errors.Wrapf(err, "error unmarshalling in expression"))
		}
		return &inExpr, nil
	case ExprTypeLt:
		var ltExpr LtExpr
		if err := json.Unmarshal(data, &ltExpr); err != nil {
			return nil, pe.WithStack(errors.Wrapf(err, "error unmarshalling lt expression"))
		}
		return &ltExpr, nil
	case ExprTypeLte:
		var lteExpr LteExpr
		if err := json.Unmarshal(data, &lteExpr); err != nil {
			return nil, pe.WithStack(errors.Wrapf(err, "error unmarshalling lte expression"))
		}
		return &lteExpr, nil
	case ExprTypeNe:
		var neExpr NeExpr
		if err := json.Unmarshal(data, &neExpr); err != nil {
			return nil, pe.WithStack(errors.Wrapf(err, "error unmarshalling ne expression"))
		}
		return &neExpr, nil
	case ExprTypeNin:
		var ninExpr NinExpr
		if err := json.Unmarshal(data, &ninExpr); err != nil {
			return nil, pe.WithStack(errors.Wrapf(err, "error unmarshalling nin expression"))
		}
		return &ninExpr, nil
	case ExprTypeNot:
		var notExpr NotExpr
		if err := json.Unmarshal(data, &notExpr); err != nil {
			return nil, pe.WithStack(errors.Wrapf(err, "error unmarshalling not expression"))
		}
		return &notExpr, nil
	case ExprTypeOr:
		var orExpr OrExpr
		if err := json.Unmarshal(data, &orExpr); err != nil {
			return nil, pe.WithStack(errors.Wrapf(err, "error unmarshalling or expression"))
		}
		return &orExpr, nil
	case ExprTypeRegex:
		var regexExpr RegexExpr
		if err := json.Unmarshal(data, &regexExpr); err != nil {
			return nil, pe.WithStack(errors.Wrapf(err, "error unmarshalling regex expression"))
		}
		return &regexExpr, nil
	case ExprTypeSize:
		var sizeExpr SizeExpr
		if err := json.Unmarshal(data, &sizeExpr); err != nil {
			return nil, pe.WithStack(errors.Wrapf(err, "error unmarshalling size expression"))
		}
		return &sizeExpr, nil
	case ExprTypeStartsWith:
		var startsWithExpr StartsWithExpr
		if err := json.Unmarshal(data, &startsWithExpr); err != nil {
			return nil, pe.WithStack(errors.Wrapf(err, "error unmarshalling starts with expression"))
		}
		return &startsWithExpr, nil
	case ExprTypeValue:
		var valueExpr ValueExpr
		if err := json.Unmarshal(data, &valueExpr); err != nil {
			return nil, pe.WithStack(errors.Wrapf(err, "error unmarshalling value expression"))
		}
		return &valueExpr, nil
	default:
		return nil, pe.WithStack(fmt.Errorf("unknown expression type: %s", exprType))
	}
}
