package transpiler

import (
	"fmt"
	"reflect"
	"sync"
)

// 全局方法缓存
var methodCache = sync.Map{}

func MethodPropAccessHandler(access PropAccess, obj any) (any, error) {
	if access.IsCall {
		val := reflect.ValueOf(obj)
		typ := val.Type()

		// 生成缓存key
		cacheKey := typ.Name() + "." + access.Prop

		// 从缓存获取方法和索引
		var method reflect.Value
		if miAny, ok := methodCache.Load(cacheKey); ok {
			mi := miAny.(int)
			method = val.Method(mi)
		} else {
			method = val.MethodByName(access.Prop)
			if !method.IsValid() {
				return nil, fmt.Errorf("method not found: %s", access.Prop)
			}
			mi := -1
			// 查找方法索引
			for i := 0; i < typ.NumMethod(); i++ {
				if typ.Method(i).Name == access.Prop {
					mi = i
					break
				}
			}

			methodCache.Store(cacheKey, mi)
		}

		args := make([]reflect.Value, len(access.Args))
		for i, arg := range access.Args {
			args[i] = reflect.ValueOf(arg)
		}

		results := method.Call(args)
		if len(results) == 0 {
			return nil, nil
		}
		return results[0].Interface(), nil
	}
	return nil, ErrPropNotSupport
}
