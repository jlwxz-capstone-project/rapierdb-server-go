package query

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/transpiler"
	qfe "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/query_filter_expr"
)

type DbWrapper struct {
	QueryExecutor *QueryExecutor
}

func DbWrapperAccessHandler(access transpiler.PropAccess, obj any) (any, error) {
	if dbWrapper, ok := obj.(*DbWrapper); ok {
		if !access.IsCall {
			if collection, ok := access.Prop.(string); ok {
				ok := dbWrapper.QueryExecutor.StorageEngine.IsValidCollection(collection)
				if !ok {
					return nil, errors.WithStack(fmt.Errorf("collection %s not found", collection))
				}
				return &CollectionWrapper{
					QueryExecutor: dbWrapper.QueryExecutor,
					Collection:    collection,
				}, nil
			}
		}
	}
	return nil, transpiler.ErrPropNotSupport
}

type CollectionWrapper struct {
	QueryExecutor *QueryExecutor
	Collection    string
}

func CollectionWrapperAccessHandler(access transpiler.PropAccess, obj any) (any, error) {
	if _, ok := obj.(*CollectionWrapper); ok {
		if access.IsCall {
			if access.Prop == "query" {
				if len(access.Args) != 1 {
					return nil, errors.WithStack(fmt.Errorf("query expects 1 argument"))
				}
				filter := transpiler.GetField(access.Args[0], "filter")
				sort := transpiler.GetField(access.Args[0], "sort")
				skip := transpiler.GetField(access.Args[0], "skip")
				limit := transpiler.GetField(access.Args[0], "limit")
				fmt.Println(filter, sort, skip, limit)
			}
		}
	}
	return nil, transpiler.ErrPropNotSupport
}

func ToQueryFilterExpr(v any) (qfe.QueryFilterExpr, error) {
	switch val := v.(type) {
	case qfe.QueryFilterExpr:
		return val, nil
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string, nil:
		return &qfe.ValueExpr{Value: val}, nil
	case []interface{}:
		return &qfe.ValueExpr{Value: val}, nil
	case map[string]interface{}:
		return &qfe.ValueExpr{Value: val}, nil
	default:
		return nil, errors.WithStack(fmt.Errorf("unsupported value type: %T", v))
	}
}

// 说明！下面的 Wrapper 函数在出错时会直接 panic！！！！！

func EqWrapper(o1 any, o2 any) qfe.QueryFilterExpr {
	o1_, err := ToQueryFilterExpr(o1)
	if err != nil {
		panic(err)
	}
	o2_, err := ToQueryFilterExpr(o2)
	if err != nil {
		panic(err)
	}
	return &qfe.EqExpr{
		O1: o1_,
		O2: o2_,
	}
}

func FieldWrapper(field string) qfe.QueryFilterExpr {
	return &qfe.FieldValueExpr{
		Path: field,
	}
}
