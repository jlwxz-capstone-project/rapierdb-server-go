package query

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/transpiler"
	qfe "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/query_filter_expr"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
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
	if cw, ok := obj.(*CollectionWrapper); ok {
		if access.IsCall {
			if access.Prop == "findMany" {
				if len(access.Args) != 1 {
					return nil, errors.WithStack(fmt.Errorf("query expects 1 argument"))
				}
				filter := transpiler.GetField(access.Args[0], "filter")
				sort := transpiler.GetField(access.Args[0], "sort")
				skip := transpiler.GetField(access.Args[0], "skip")
				limit := transpiler.GetField(access.Args[0], "limit")

				query := &FindManyQuery{}

				if filter, ok := filter.(qfe.QueryFilterExpr); ok {
					query.Filter = filter
				} else {
					return nil, errors.WithStack(fmt.Errorf("invalid query: filter must be a QueryFilterExpr"))
				}

				if sort != nil {
					if sortAnyArray, ok := sort.([]any); ok {
						sortArray := make([]SortField, len(sortAnyArray))
						for i, v := range sortAnyArray {
							if sortField, ok := v.(SortField); ok {
								sortArray[i] = sortField
							} else {
								return nil, errors.WithStack(fmt.Errorf("invalid query: sort must be a []SortField"))
							}
						}
					} else {
						return nil, errors.WithStack(fmt.Errorf("invalid query: sort must be a []SortField"))
					}
				}

				if skip != nil {
					switch v := skip.(type) {
					case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
						query.Skip = util.ToInt64(v)
					default:
						return nil, errors.WithStack(fmt.Errorf("invalid query: skip must be an int-like value"))
					}
				}

				if limit != nil {
					switch v := limit.(type) {
					case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
						query.Limit = util.ToInt64(v)
					default:
						return nil, errors.WithStack(fmt.Errorf("invalid query: limit must be an int-like value"))
					}
				}

				docs, err := cw.QueryExecutor.FindMany(cw.Collection, query)
				if err != nil {
					return nil, errors.WithStack(err)
				}

				return docs, nil
			} else if access.Prop == "findOne" {
				if len(access.Args) != 1 {
					return nil, errors.WithStack(fmt.Errorf("query expects 1 argument"))
				}

				query := &FindOneQuery{}
				filter := transpiler.GetField(access.Args[0], "filter")
				if filter, ok := filter.(qfe.QueryFilterExpr); ok {
					query.Filter = filter
				} else {
					return nil, errors.WithStack(fmt.Errorf("invalid query: filter must be a QueryFilterExpr"))
				}

				doc, err := cw.QueryExecutor.FindOne(cw.Collection, query)
				if err != nil {
					return nil, errors.WithStack(err)
				}

				return doc, nil
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

func SortAscWrapper(path string) SortField {
	return SortField{
		Field: path,
		Order: SortOrderAsc,
	}
}

func SortDescWrapper(path string) SortField {
	return SortField{
		Field: path,
		Order: SortOrderDesc,
	}
}

// TODO 仅用于测试
func LogWrapper(msg any) {
	fmt.Println(msg)
}
