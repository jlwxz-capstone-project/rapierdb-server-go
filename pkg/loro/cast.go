package loro

import "fmt"

func ToInt64(v interface{}) int64 {
	switch v := v.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	}
	panic(fmt.Sprintf("unsupported type: %T", v))
}

func ToFloat64(v interface{}) float64 {
	switch v := v.(type) {
	case float32:
		return float64(v)
	case float64:
		return v
	}
	panic(fmt.Sprintf("unsupported type: %T", v))
}
