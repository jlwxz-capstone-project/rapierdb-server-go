package util

import (
	"errors"
	"fmt"
)

var (
	ErrTypeError = errors.New("type error")
)

// CompareValues 比较两个值的大小
// 返回值：
//   - 如果 o1 < o2，返回 -1
//   - 如果 o1 = o2，返回 0
//   - 如果 o1 > o2，返回 1
//   - 如果类型不匹配或不支持比较，返回错误
func CompareValues(o1, o2 any) (int, error) {
	// 处理 nil 情况
	if o1 == nil || o2 == nil {
		if o1 == nil && o2 == nil {
			return 0, nil
		}
		if o1 == nil {
			return -1, nil
		}
		return 1, nil
	}

	switch v1 := o1.(type) {
	case bool:
		if v2, ok := o2.(bool); ok {
			if v1 == v2 {
				return 0, nil
			}
			if !v1 && v2 {
				return -1, nil
			}
			return 1, nil
		}
		return 0, fmt.Errorf("%w: comparing bool with %T", ErrTypeError, o2)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		n1 := ToInt64(v1)
		switch v2 := o2.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			n2 := ToInt64(v2)
			if n1 < n2 {
				return -1, nil
			}
			if n1 > n2 {
				return 1, nil
			}
			return 0, nil
		case float32, float64:
			// 整数转浮点数进行比较
			f1 := float64(n1)
			f2 := ToFloat64(v2)
			if f1 < f2 {
				return -1, nil
			}
			if f1 > f2 {
				return 1, nil
			}
			return 0, nil
		default:
			return 0, fmt.Errorf("%w: comparing numeric type with %T", ErrTypeError, o2)
		}
	case float32, float64:
		f1 := ToFloat64(v1)
		switch v2 := o2.(type) {
		case float32, float64:
			f2 := ToFloat64(v2)
			if f1 < f2 {
				return -1, nil
			}
			if f1 > f2 {
				return 1, nil
			}
			return 0, nil
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			// 整数转浮点数进行比较
			f2 := float64(ToInt64(v2))
			if f1 < f2 {
				return -1, nil
			}
			if f1 > f2 {
				return 1, nil
			}
			return 0, nil
		default:
			return 0, fmt.Errorf("%w: comparing float type with %T", ErrTypeError, o2)
		}
	case string:
		if v2, ok := o2.(string); ok {
			if v1 < v2 {
				return -1, nil
			}
			if v1 > v2 {
				return 1, nil
			}
			return 0, nil
		}
		return 0, fmt.Errorf("%w: comparing string with %T", ErrTypeError, o2)
	case []any:
		if v2, ok := o2.([]any); ok {
			if len(v1) != len(v2) {
				return len(v1) - len(v2), nil
			}
			for i := range v1 {
				cmp, err := CompareValues(v1[i], v2[i])
				if err != nil {
					if errors.Is(err, ErrTypeError) {
						return 0, fmt.Errorf("%w: comparing array elements at index %d", err, i)
					}
					return 0, err
				}
				if cmp != 0 {
					return cmp, nil
				}
			}
			return 0, nil
		}
		return 0, fmt.Errorf("%w: comparing array with %T", ErrTypeError, o2)
	case map[string]any:
		if v2, ok := o2.(map[string]any); ok {
			if len(v1) != len(v2) {
				return len(v1) - len(v2), nil
			}
			for k, v := range v1 {
				if v2v, exists := v2[k]; !exists {
					return 1, nil
				} else {
					cmp, err := CompareValues(v, v2v)
					if err != nil {
						if errors.Is(err, ErrTypeError) {
							return 0, fmt.Errorf("%w: comparing map values for key '%s'", err, k)
						}
						return 0, err
					}
					if cmp != 0 {
						return cmp, nil
					}
				}
			}
			return 0, nil
		}
		return 0, fmt.Errorf("%w: comparing map with %T", ErrTypeError, o2)
	default:
		return 0, fmt.Errorf("%w: unsupported type %T", ErrTypeError, o1)
	}
}
