package transpiler

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// PropAccess 表示一次属性访问
//
// 例如：obj.method(arg1, arg2) 对应的 PropAccess 为：
//
//	PropAccess{
//		Prop: "method",
//		Args: []any{"arg1", "arg2"},
//		IsCall: true,
//	}
//
// obj.name 对应的 PropAccess 为：
//
//	PropAccess{
//		Prop: "name",
//	}
type PropAccess struct {
	// 属性名
	Prop string
	// 如果是函数调用，这里是参数
	Args []any
	// 是否是函数调用
	IsCall bool
}

type PropGetter func(chain []PropAccess, obj any) (any, error)

type PropAccessHandler func(access PropAccess, obj any) (any, error)

func NewPropGetter(propAccessHandlers ...PropAccessHandler) PropGetter {
	return func(chain []PropAccess, obj any) (any, error) {
		result := obj
		for _, access := range chain {
			success := false
			for _, handler := range propAccessHandlers {
				resultNew, err := handler(access, result)
				if err == nil {
					success = true
					result = resultNew
					break
				}
			}
			if !success {
				return nil, ErrPropNotSupport
			}
		}
		return result, nil
	}
}

// DefaultPropGetter 默认的属性访问器，用于根据 JavaScript 的属性访问获取正确的值
//
// chain 是属性访问链，obj 是根对象。比如 obj.name.slice(1, 2).toUpperCase() 对应的 chain 为：
//
//	chain = []PropAccessor{
//		{Prop: "name"},
//		{Prop: "slice", Args: []any{1, 2}, IsCall: true},
//		{Prop: "toUpperCase", IsCall: true},
//	}
var DefaultPropGetter = NewPropGetter(
	StringPropAccessHandler,
	ArrayPropAccessHandler,
	MethodCallHandler,
	DataFieldAccessHandler,
)

func GetField(obj any, fieldKey string) any {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Map:
		// 对于 map，使用反射获取字段值
		fieldKeyVal := reflect.ValueOf(fieldKey)
		fieldVal := val.MapIndex(fieldKeyVal)
		if !fieldVal.IsValid() {
			return nil
		}
		return fieldVal.Interface()

	case reflect.Struct:
		// 对于结构体，使用反射获得字段值
		field := val.FieldByName(fieldKey)
		if !field.IsValid() {
			return nil
		}
		return field.Interface()
	}
	return nil
}

func MethodCallHandler(access PropAccess, obj any) (any, error) {
	if !access.IsCall {
		return nil, ErrPropNotSupport
	}

	val := reflect.ValueOf(obj)

	// 先尝试获取方法
	method := val.MethodByName(access.Prop)
	if method.IsValid() {
		args := make([]reflect.Value, len(access.Args))
		for i, arg := range access.Args {
			args[i] = reflect.ValueOf(arg)
		}
		results := method.Call(args)
		if len(results) == 0 {
			return nil, nil
		}
		// Js 不支持多值返回，因此仅返回第一个返回值
		return results[0].Interface(), nil
	}

	// 如果获取方法失败，再尝试获取属性，如果获取到的属性值是函数也行
	field := GetField(obj, access.Prop)
	if field == nil {
		return nil, fmt.Errorf("property not found: %s", access.Prop)
	}

	// 检查属性是否是可调用的函数
	fieldVal := reflect.ValueOf(field)
	if fieldVal.Kind() != reflect.Func {
		return nil, fmt.Errorf("property %s is not callable", access.Prop)
	}

	// 调用函数
	args := make([]reflect.Value, len(access.Args))
	for i, arg := range access.Args {
		args[i] = reflect.ValueOf(arg)
	}
	results := fieldVal.Call(args)
	if len(results) == 0 {
		return nil, nil
	}
	// Js 不支持多值返回，因此仅返回第一个返回值
	return results[0].Interface(), nil
}

func DataFieldAccessHandler(access PropAccess, obj any) (any, error) {
	if access.IsCall {
		return nil, ErrPropNotSupport
	}

	// 获取属性值
	field := GetField(obj, access.Prop)
	if field == nil {
		return nil, fmt.Errorf("property not found: %s", access.Prop)
	}
	return field, nil
}

func ArrayPropAccessHandler(access PropAccess, obj any) (any, error) {
	// 检查是否是切片类型
	val := reflect.ValueOf(obj)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil, ErrPropNotSupport
	}

	switch access.Prop {
	case "length":
		return val.Len(), nil

	case "slice":
		if !access.IsCall {
			return nil, ErrPropNotSupport
		}

		// 检查参数
		if len(access.Args) < 1 || len(access.Args) > 2 {
			return nil, fmt.Errorf("%w: slice method requires 1 or 2 arguments", ErrInCall)
		}

		// 获取起始位置
		start, ok := toInt(access.Args[0])
		if !ok {
			return nil, fmt.Errorf("%w: first argument of slice must be a number", ErrInCall)
		}

		// 处理负数索引
		if start < 0 {
			start = val.Len() + start
		}
		if start < 0 {
			start = 0
		}

		end := val.Len()
		if len(access.Args) > 1 {
			// 获取结束位置
			if e, ok := toInt(access.Args[1]); ok {
				if e < 0 {
					end = val.Len() + e
				} else {
					end = e
				}
			} else {
				return nil, fmt.Errorf("%w: second argument of slice must be a number", ErrInCall)
			}
		}

		// 边界检查
		if start > val.Len() {
			start = val.Len()
		}
		if end > val.Len() {
			end = val.Len()
		}
		if end < start {
			end = start
		}

		// 直接返回切片
		return val.Slice(start, end).Interface(), nil

	case "indexOf":
		if !access.IsCall {
			return nil, ErrPropNotSupport
		}

		if len(access.Args) < 1 {
			return nil, fmt.Errorf("%w: indexOf method requires 1 argument", ErrInCall)
		}

		// 遍历查找元素
		searchVal := access.Args[0]
		for i := 0; i < val.Len(); i++ {
			current := val.Index(i).Interface()
			// 使用 CompareValues 进行比较
			cmp, err := util.CompareValues(current, searchVal)
			if err == nil && cmp == 0 {
				return i, nil
			}
		}
		return -1, nil

	case "join":
		if !access.IsCall {
			return nil, ErrPropNotSupport
		}

		// 默认分隔符
		separator := ","
		if len(access.Args) > 0 {
			if sep, ok := access.Args[0].(string); ok {
				separator = sep
			}
		}

		// 构建字符串
		var result strings.Builder
		for i := 0; i < val.Len(); i++ {
			if i > 0 {
				result.WriteString(separator)
			}
			result.WriteString(fmt.Sprint(val.Index(i).Interface()))
		}
		return result.String(), nil

	case "splice":
		if !access.IsCall {
			return nil, ErrPropNotSupport
		}

		// 检查参数
		if len(access.Args) < 1 {
			return nil, fmt.Errorf("%w: splice method requires at least 1 argument", ErrInCall)
		}

		// 获取起始位置
		start, ok := toInt(access.Args[0])
		if !ok {
			return nil, fmt.Errorf("%w: first argument of splice must be a number", ErrInCall)
		}

		// 处理负数索引
		if start < 0 {
			start = val.Len() + start
		}
		if start < 0 {
			start = 0
		}
		if start > val.Len() {
			start = val.Len()
		}

		// 获取删除数量
		deleteCount := val.Len() - start
		if len(access.Args) > 1 {
			if count, ok := toInt(access.Args[1]); ok {
				if count < 0 {
					count = 0
				}
				if start+count > val.Len() {
					deleteCount = val.Len() - start
				} else {
					deleteCount = count
				}
			} else {
				return nil, fmt.Errorf("%w: second argument of splice must be a number", ErrInCall)
			}
		}

		// 创建新切片存储结果
		newLen := val.Len() - deleteCount + len(access.Args) - 2
		if newLen < 0 {
			newLen = 0
		}

		// 创建新的切片，类型与原切片相同
		newSlice := reflect.MakeSlice(val.Type(), newLen, newLen)

		// 复制前半部分
		if start > 0 {
			reflect.Copy(newSlice.Slice(0, start), val.Slice(0, start))
		}

		// 插入新元素
		insertCount := len(access.Args) - 2
		if insertCount > 0 {
			for i := 0; i < insertCount; i++ {
				newVal := reflect.ValueOf(access.Args[i+2])
				if !newVal.Type().AssignableTo(val.Type().Elem()) {
					return nil, fmt.Errorf("%w: cannot insert value of type %v into slice of type %v",
						ErrInCall, newVal.Type(), val.Type().Elem())
				}
				newSlice.Index(start + i).Set(newVal)
			}
		}

		// 复制后半部分
		if start+deleteCount < val.Len() {
			reflect.Copy(
				newSlice.Slice(start+insertCount, newLen),
				val.Slice(start+deleteCount, val.Len()),
			)
		}

		// 直接返回切片
		return newSlice.Interface(), nil
	}

	return nil, ErrPropNotSupport
}

func StringPropAccessHandler(access PropAccess, obj any) (any, error) {
	if str, ok := obj.(string); ok {
		switch access.Prop {
		case "length":
			return len(str), nil

		case "toLowerCase":
			if access.IsCall {
				return strings.ToLower(str), nil
			}
			return strings.ToLower, nil

		case "toUpperCase":
			if access.IsCall {
				return strings.ToUpper(str), nil
			}
			return strings.ToUpper, nil

		case "trim":
			if access.IsCall {
				return strings.TrimSpace(str), nil
			}
			return strings.TrimSpace, nil

		case "substring":
			if access.IsCall {
				if len(access.Args) < 1 {
					return nil, fmt.Errorf("%w: substring method requires 1 argument", ErrInCall)
				}
				start, ok := access.Args[0].(int)
				if !ok {
					return nil, fmt.Errorf("%w: first argument of substring must be a number", ErrInCall)
				}
				if start < 0 {
					start = 0
				}

				if len(access.Args) > 1 {
					end, ok := access.Args[1].(int)
					if !ok {
						return nil, fmt.Errorf("%w: second argument of substring must be a number", ErrInCall)
					}
					if end > len(str) {
						end = len(str)
					}
					obj = str[start:end]
				} else {
					obj = str[start:]
				}
				return obj, nil
			}

		case "indexOf":
			// 查找子串位置
			if access.IsCall {
				if len(access.Args) < 1 {
					return nil, fmt.Errorf("%w: indexOf method requires 1 argument", ErrInCall)
				}
				substr, ok := access.Args[0].(string)
				if !ok {
					return nil, fmt.Errorf("%w: argument of indexOf must be a string", ErrInCall)
				}
				return strings.Index(str, substr), nil
			}

		case "replace":
			// 替换字符串
			if access.IsCall {
				if len(access.Args) < 2 {
					return nil, fmt.Errorf("%w: replace method requires 2 arguments", ErrInCall)
				}
				old, ok1 := access.Args[0].(string)
				new, ok2 := access.Args[1].(string)
				if !ok1 || !ok2 {
					return nil, fmt.Errorf("%w: arguments of replace must be strings", ErrInCall)
				}
				return strings.Replace(str, old, new, 1), nil
			}
		}
	}
	return nil, ErrPropNotSupport
}
