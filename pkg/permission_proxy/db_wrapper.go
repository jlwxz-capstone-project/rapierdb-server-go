package permission_proxy

import (
	"fmt"
	"unsafe"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/transpiler"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	qfe "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/query_filter_expr"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query_executor"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
	"github.com/pkg/errors"
)

// a wrapper around a query executor, used as a readonly
// view of a database in the permission system
type DbWrapper struct {
	QueryExecutor *query_executor.QueryExecutor
}

// allow users to write `db["<collection_name>"]` to get a collection wrapper
func DbWrapperAccessHandler(access transpiler.PropAccess, obj any) (any, error) {
	if dbWrapper, ok := obj.(*DbWrapper); ok {
		if !access.IsCall {
			if collection, ok := access.Prop.(string); ok {
				ok := dbWrapper.QueryExecutor.IsValidCollection(collection)
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
	QueryExecutor *query_executor.QueryExecutor
	Collection    string
}

// allow users to write
//
//	<collection_wrapper>.findMany({
//	  filter: {...},
//	  sort: [...],
//	  skip: 0,
//	  limit: 10,
//	})
//
// to create and execute a findMany query
//
// also allow users to write
//
//	<collection_wrapper>.findOne({
//	  filter: {...},
//	})
//
// to create and execute a findOne query
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

				q := &query.FindManyQuery{
					Collection: cw.Collection,
				}

				// construct filter
				if filter, ok := filter.(qfe.QueryFilterExpr); ok {
					q.Filter = filter
				} else {
					return nil, errors.WithStack(fmt.Errorf("invalid query: filter must be a QueryFilterExpr"))
				}

				// construct sort
				if sort != nil {
					if sortAnyArray, ok := sort.([]any); ok {
						sortArray := make([]query.SortField, len(sortAnyArray))
						for i, v := range sortAnyArray {
							if sortField, ok := v.(query.SortField); ok {
								sortArray[i] = sortField
							} else {
								return nil, errors.WithStack(fmt.Errorf("invalid query: sort must be a []SortField"))
							}
						}
					} else {
						return nil, errors.WithStack(fmt.Errorf("invalid query: sort must be a []SortField"))
					}
				}

				// construct skip
				if skip != nil {
					switch v := skip.(type) {
					case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
						q.Skip = util.ToInt64(v)
					default:
						return nil, errors.WithStack(fmt.Errorf("invalid query: skip must be an int-like value"))
					}
				}

				// construct limit
				if limit != nil {
					switch v := limit.(type) {
					case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
						q.Limit = util.ToInt64(v)
					default:
						return nil, errors.WithStack(fmt.Errorf("invalid query: limit must be an int-like value"))
					}
				}

				// execute query
				docs, err := cw.QueryExecutor.FindMany(q)
				if err != nil {
					return nil, errors.WithStack(err)
				}

				return docs, nil
			} else if access.Prop == "findOne" {
				if len(access.Args) != 1 {
					return nil, errors.WithStack(fmt.Errorf("query expects 1 argument"))
				}

				q := &query.FindOneQuery{
					Collection: cw.Collection,
				}
				filter := transpiler.GetField(access.Args[0], "filter")
				if filter, ok := filter.(qfe.QueryFilterExpr); ok {
					q.Filter = filter
				} else {
					return nil, errors.WithStack(fmt.Errorf("invalid query: filter must be a QueryFilterExpr"))
				}

				doc, err := cw.QueryExecutor.FindOne(q)
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
		Path: qfe.NewValueExpr(field),
	}
}

func SortAscWrapper(path string) query.SortField {
	return query.SortField{
		Field: path,
		Order: query.SortOrderAsc,
	}
}

func SortDescWrapper(path string) query.SortField {
	return query.SortField{
		Field: path,
		Order: query.SortOrderDesc,
	}
}

func isNil(v any) bool {
	return (*[2]uintptr)(unsafe.Pointer(&v))[1] == 0
}

func LogWrapper(args ...any) {
	fmt.Println(args...)
}

func DocWithIdAccessHandler(access transpiler.PropAccess, obj any) (any, error) {
	if docWithId, ok := obj.(*query.DocWithId); ok {
		if !access.IsCall {
			if prop, ok := access.Prop.(string); ok {
				if prop == "id" {
					return docWithId.DocId, nil
				} else {
					// delegate to LoroDocAccessHandler
					return LoroDocAccessHandler(access, docWithId.Doc)
				}
			}
		}
	}
	return nil, transpiler.ErrPropNotSupport
}
