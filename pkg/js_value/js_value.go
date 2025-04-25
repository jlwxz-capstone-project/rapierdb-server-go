package js_value

import (
	"errors"
	"unsafe"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
	pe "github.com/pkg/errors"
)

type JsValue any

var (
	ErrCompareTypeMismatch = errors.New("compare type mismatch")
	ErrImpossibleType      = errors.New("impossible type")
)

func isNil(v any) bool {
	return (*[2]uintptr)(unsafe.Pointer(&v))[1] == 0
}

// ToJsValue 尝试将任意类型的值转换为 JsValue
//
//	type JsValue =
//		| bool
//		| float64
//		| string
//		| []JsValue
//		| map[string]JsValue
//		| nil
func ToJsValue(value any) (JsValue, error) {
	if isNil(value) {
		return nil, nil
	}

	switch v := value.(type) {
	case *loro.LoroMap:
		return v.ToGoObject()
	case *loro.LoroList:
		return v.ToGoObject()
	case *loro.LoroMovableList:
		return v.ToGoObject()
	case *loro.LoroText:
		return v.ToString()
	case map[string]any:
		result := make(map[string]any, len(v))
		for key, item := range v {
			tmp, err := ToJsValue(item)
			if err != nil {
				return nil, err
			}
			result[key] = tmp
		}
		return result, nil
	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			tmp, err := ToJsValue(item)
			if err != nil {
				return nil, err
			}
			result[i] = tmp
		}
		return result, nil
	case bool, string, float64:
		return v, nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32:
		return util.ToFloat64(v), nil
	default:
		return nil, pe.Wrapf(ErrImpossibleType, "type=%T", v)
	}
}

func DeepComapreJsValue(v1, v2 JsValue) (int, error) {
	if isNil(v1) || isNil(v2) {
		if isNil(v1) && isNil(v2) {
			return 0, nil
		}
		if isNil(v1) {
			return -1, nil
		}
		return 1, nil
	}

	v1bool, v1isBool := v1.(bool)
	if v1isBool {
		v2bool, v2isBool := v2.(bool)
		if v2isBool {
			if v1bool == v2bool {
				return 0, nil
			}
			if !v1bool && v2bool {
				return -1, nil
			}
			return 1, nil
		}
		return 0, pe.Wrapf(ErrCompareTypeMismatch, "o1=%v, o2=%v", v1, v2)
	}

	v1number, v1isNumber := v1.(float64)
	if v1isNumber {
		v2number, v2isNumber := v2.(float64)
		if v2isNumber {
			if v1number == v2number {
				return 0, nil
			}
			if v1number < v2number {
				return -1, nil
			}
			return 1, nil
		}
		return 0, pe.Wrapf(ErrCompareTypeMismatch, "o1=%v, o2=%v", v1, v2)
	}

	v1string, v1isString := v1.(string)
	if v1isString {
		v2string, v2isString := v2.(string)
		if v2isString {
			if v1string == v2string {
				return 0, nil
			}
			if v1string < v2string {
				return -1, nil
			}
			return 1, nil
		}
		return 0, pe.Wrapf(ErrCompareTypeMismatch, "o1=%v, o2=%v", v1, v2)
	}

	v1array, v1isArray := v1.([]any)
	if v1isArray {
		v2array, v2isArray := v2.([]any)
		if v2isArray {
			if len(v1array) != len(v2array) {
				return len(v1array) - len(v2array), nil
			}
			for i, item1 := range v1array {
				item2 := v2array[i]
				cmp, err := DeepComapreJsValue(item1, item2)
				if err != nil {
					return 0, err
				}
				if cmp != 0 {
					return cmp, nil
				}
			}
			return 0, nil
		}
		return 0, pe.Wrapf(ErrCompareTypeMismatch, "o1=%v, o2=%v", v1, v2)
	}

	v1map, v1isMap := v1.(map[string]any)
	if v1isMap {
		v2map, v2isMap := v2.(map[string]any)
		if v2isMap {
			if len(v1map) != len(v2map) {
				return len(v1map) - len(v2map), nil
			}
			for key1, item1 := range v1map {
				item2 := v2map[key1]
				cmp, err := DeepComapreJsValue(item1, item2)
				if err != nil {
					return 0, err
				}
				if cmp != 0 {
					return cmp, nil
				}
			}
			return 0, nil
		}
		return 0, pe.Wrapf(ErrCompareTypeMismatch, "o1=%v, o2=%v", v1, v2)
	}

	return 0, pe.Wrapf(ErrImpossibleType, "o1=%v, o2=%v", v1, v2)
}

func DeepEqualJsValue(v1, v2 JsValue) (bool, error) {
	cmp, err := DeepComapreJsValue(v1, v2)
	if err != nil {
		return false, err
	}
	return cmp == 0, nil
}
