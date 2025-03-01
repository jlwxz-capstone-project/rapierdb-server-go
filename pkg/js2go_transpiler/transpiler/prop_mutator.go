package transpiler

import (
	"fmt"
	"reflect"

	"github.com/cockroachdb/errors"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

// PropMutator 表示赋值处理器
type PropMutator func(obj any, propName any, value any) error

// PropMutateHandler 表示单个赋值处理方法
type PropMutateHandler func(obj any, propName any, value any) error

// NewPropMutator 创建一个赋值处理器，顺序尝试每个处理方法
func NewPropMutator(propMutateHandlers ...PropMutateHandler) PropMutator {
	return func(obj any, propName any, value any) error {
		for _, handler := range propMutateHandlers {
			err := handler(obj, propName, value)
			if err == nil {
				return nil
			}
		}
		return errors.WithStack(fmt.Errorf("cannot assign to property %v of %T", propName, obj))
	}
}

// DefaultPropMutator 默认的赋值处理器
var DefaultPropMutator = NewPropMutator(
	MapPropMutateHandler,
	SlicePropMutateHandler,
	StructPropMutateHandler,
)

// MapPropMutateHandler 处理 map 类型的属性赋值
func MapPropMutateHandler(obj any, propName any, value any) error {
	if m, ok := obj.(map[string]any); ok {
		key := fmt.Sprint(propName)
		m[key] = value
		return nil
	}
	return ErrPropNotSupport
}

// SlicePropMutateHandler 处理切片类型的属性赋值
func SlicePropMutateHandler(obj any, propName any, value any) error {
	val := reflect.ValueOf(obj)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return ErrPropNotSupport
	}

	// 只支持整数索引
	index, ok := toInt(propName)
	if !ok {
		return errors.WithStack(fmt.Errorf("slice index must be an integer: %v", propName))
	}

	// 检查索引范围
	if index < 0 || index >= val.Len() {
		return errors.WithStack(fmt.Errorf("index out of range: %d", index))
	}

	// 设置值
	elemVal := reflect.ValueOf(value)
	if !elemVal.Type().AssignableTo(val.Type().Elem()) {
		return errors.WithStack(fmt.Errorf("cannot assign %T to slice element of type %s",
			value, val.Type().Elem()))
	}

	val.Index(index).Set(elemVal)
	return nil
}

// StructPropMutateHandler 处理结构体类型的属性赋值
func StructPropMutateHandler(obj any, propName any, value any) error {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return ErrPropNotSupport
	}

	// 只支持字符串属性名
	fieldName, ok := propName.(string)
	if !ok {
		fieldName = fmt.Sprint(propName)
	}

	// 查找字段
	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("field not found: %s", fieldName)
	}

	// 检查字段是否可设置
	if !field.CanSet() {
		return errors.WithStack(fmt.Errorf("field cannot be set: %s", fieldName))
	}

	// 设置值
	valueVal := reflect.ValueOf(value)
	if !valueVal.Type().AssignableTo(field.Type()) {
		return errors.WithStack(fmt.Errorf("cannot assign %T to field of type %s",
			value, field.Type()))
	}

	field.Set(valueVal)
	return nil
}

func LoroListPropMutateHandler(obj any, propName any, value any) error {
	if ll, ok := obj.(loro.LoroList); ok {
		switch v := value.(type) {
		case int:
			index := v
			len := ll.GetLen()
			if index < 0 || index >= int(len) {
				return errors.WithStack(fmt.Errorf("index out of range: %d", index))
			}
			// TODO
		}
	}
	return nil
}
