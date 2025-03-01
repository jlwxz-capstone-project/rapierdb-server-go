package query

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/transpiler"
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
