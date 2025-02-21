package transpiler

import (
	"fmt"
	"reflect"
)

func DataPropAccessHandler(access PropAccess, obj any) (any, error) {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 添加函数调用处理
	if val.Kind() == reflect.Struct {
		// 尝试获取方法
		method := val.MethodByName(access.Prop)
		if method.IsValid() {
			if access.IsCall {
				// 处理函数调用
				args := make([]reflect.Value, len(access.Args))
				for i, arg := range access.Args {
					args[i] = reflect.ValueOf(arg)
				}
				results := method.Call(args)
				if len(results) > 0 {
					return results[0].Interface(), nil
				}
				return nil, nil
			}
			return method.Interface(), nil
		}
	}

	// 非方法调用的属性访问
	switch val.Kind() {
	case reflect.Map:
		// 对于 map，先尝试直接访问
		if m, ok := obj.(map[string]any); ok {
			if val, ok := m[access.Prop]; ok {
				return val, nil
			}
		}
		// 如果直接访问失败，尝试使用反射
		mapVal := val.MapIndex(reflect.ValueOf(access.Prop))
		if !mapVal.IsValid() {
			// 尝试查找方法
			method := val.MethodByName(access.Prop)
			if method.IsValid() {
				return method.Interface(), nil
			}
			return nil, fmt.Errorf("property not found: %s", access.Prop)
		}
		return mapVal.Interface(), nil

	case reflect.Struct:
		field := val.FieldByName(access.Prop)
		if !field.IsValid() {
			// 尝试查找方法
			method := val.MethodByName(access.Prop)
			if method.IsValid() {
				return method.Interface(), nil
			}
			return nil, fmt.Errorf("property not found: %s", access.Prop)
		}
		return field.Interface(), nil

	case reflect.Slice, reflect.Array:
		// 对于切片和数组，只支持 length 属性
		if access.Prop == "length" {
			return val.Len(), nil
		}
		return nil, fmt.Errorf("unsupported slice property: %s", access.Prop)

	default:
		return nil, fmt.Errorf("unsupported object type: %v", val.Kind())
	}
}
