package transpiler

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js_value"
	pe "github.com/pkg/errors"
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
	// 属性名，可以是 string 或 int
	Prop any
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
				// 如果错误是 ErrPropNotSupport，则继续尝试下一个处理器
				if pe.Is(err, ErrPropNotSupport) {
					continue
				}
				// 如果错误不是 ErrPropNotSupport，则直接返回错误
				return nil, err
			}
			if !success {
				return nil, pe.Wrapf(ErrPropNotSupport, "unsupported property access: obj=%v, access=%v", obj, access)
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
			// 尝试转换为大写后再获取
			fieldKey = strings.ToUpper(fieldKey[:1]) + fieldKey[1:]
			field = val.FieldByName(fieldKey)
			if !field.IsValid() {
				return nil
			}
		}
		return field.Interface()
	}
	return nil
}

func MethodCallHandler(access PropAccess, obj any) (any, error) {
	if !access.IsCall {
		return nil, ErrPropNotSupport
	}

	prop, ok := access.Prop.(string)
	if !ok {
		return nil, pe.Wrapf(ErrPropNotSupport, "property must be a string: %v", access.Prop)
	}

	val := reflect.ValueOf(obj)

	// 先尝试获取方法
	method := val.MethodByName(prop)
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
	field := GetField(obj, prop)
	if field == nil {
		return nil, pe.Wrapf(ErrPropNotSupport, "property not found: %s", prop)
	}

	// 检查属性是否是可调用的函数
	fieldVal := reflect.ValueOf(field)
	if fieldVal.Kind() != reflect.Func {
		return nil, pe.Wrapf(ErrPropNotSupport, "property %s is not callable", access.Prop)
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

	if prop, ok := access.Prop.(string); ok {
		// 获取属性值
		field := GetField(obj, prop)
		if field != nil {
			return field, nil
		}
	}
	return nil, ErrPropNotSupport
}

func ArrayPropAccessHandler(access PropAccess, obj any) (any, error) {
	// 检查是否是切片类型
	val := reflect.ValueOf(obj)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil, ErrPropNotSupport
	}

	switch prop := access.Prop.(type) {
	case string:
		switch access.Prop {
		case "slice":
			if !access.IsCall {
				return nil, ErrPropNotSupport
			}

			// 检查参数
			if len(access.Args) < 1 || len(access.Args) > 2 {
				return nil, pe.Wrapf(ErrPropNotSupport, "slice method requires 1 or 2 arguments")
			}

			// 获取起始位置
			start, ok := toInt(access.Args[0])
			if !ok {
				return nil, pe.Wrapf(ErrPropNotSupport, "first argument of slice must be a number")
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
					return nil, pe.Wrapf(ErrPropNotSupport, "second argument of slice must be a number")
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
				return nil, pe.Wrapf(ErrPropNotSupport, "indexOf method requires 1 argument")
			}

			// 遍历查找元素
			searchVal := access.Args[0]
			for i := 0; i < val.Len(); i++ {
				current := val.Index(i).Interface()
				cmp, err := js_value.DeepComapreJsValue(current, searchVal)
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
				return nil, pe.Wrapf(ErrPropNotSupport, "splice method requires at least 1 argument")
			}

			// 获取起始位置
			start, ok := toInt(access.Args[0])
			if !ok {
				return nil, pe.Wrapf(ErrPropNotSupport, "first argument of splice must be a number")
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
		case "length":
			return val.Len(), nil
		}
	case int:
		index := prop
		if index < 0 || index >= val.Len() {
			return nil, pe.Wrapf(ErrPropNotSupport, "index out of range: %d", index)
		}
		return val.Index(index).Interface(), nil
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
					return nil, pe.Wrapf(ErrPropNotSupport, "substring method requires 1 argument")
				}
				start, ok := access.Args[0].(int)
				if !ok {
					return nil, pe.Wrapf(ErrPropNotSupport, "first argument of substring must be a number")
				}
				if start < 0 {
					start = 0
				}

				if len(access.Args) > 1 {
					end, ok := access.Args[1].(int)
					if !ok {
						return nil, pe.Wrapf(ErrPropNotSupport, "second argument of substring must be a number")
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
					return nil, pe.Wrapf(ErrPropNotSupport, "indexOf method requires 1 argument")
				}
				substr, ok := access.Args[0].(string)
				if !ok {
					return nil, pe.Wrapf(ErrPropNotSupport, "argument of indexOf must be a string")
				}
				return strings.Index(str, substr), nil
			}

		case "replace":
			// 替换字符串
			if access.IsCall {
				if len(access.Args) < 2 {
					return nil, pe.Wrapf(ErrPropNotSupport, "replace method requires 2 arguments")
				}
				old, ok1 := access.Args[0].(string)
				new, ok2 := access.Args[1].(string)
				if !ok1 || !ok2 {
					return nil, pe.Wrapf(ErrPropNotSupport, "arguments of replace must be strings")
				}
				return strings.Replace(str, old, new, 1), nil
			}
		}
	}
	return nil, ErrPropNotSupport
}
