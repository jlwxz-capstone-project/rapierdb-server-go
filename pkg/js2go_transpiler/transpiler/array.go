package transpiler

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

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
