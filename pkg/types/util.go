package types

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
)

// LastOfArray 获取数组的最后一个元素
func LastOfArray[T any](arr []T) T {
	return arr[len(arr)-1]
}

// RandomOfArray 从数组中随机获取一个元素
func RandomOfArray[T any](items []T) T {
	return items[rand.Intn(len(items))]
}

// ShuffleArray 打乱数组元素顺序
func ShuffleArray[T any](arr []T) []T {
	result := make([]T, len(arr))
	copy(result, arr)

	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	return result
}

// NormalizeSortField 标准化排序字段
// 输入: "-age"
// 输出: "age"
func NormalizeSortField(field string) string {
	if strings.HasPrefix(field, "-") {
		return field[1:]
	}
	return field
}

// GetSortFieldsOfQuery 获取查询的排序字段
func GetSortFieldsOfQuery(query MongoQuery) []string {
	if len(query.Sort) == 0 {
		// 如果没有设置排序顺序，使用主键
		return []string{"_id"}
	}

	result := make([]string, 0, len(query.Sort))
	for _, field := range query.Sort {
		result = append(result, NormalizeSortField(field))
	}

	return result
}

// ReplaceCharAt 替换字符串中指定位置的字符
func ReplaceCharAt(str string, index int, replacement string) string {
	return str[:index] + replacement + str[index+len(replacement):]
}

// MapToObject 将 map 转换为对象
func MapToObject[K comparable, V any](m map[K]V) map[string]V {
	result := make(map[string]V)
	for k, v := range m {
		key := ""
		// 将键转换为字符串
		switch t := any(k).(type) {
		case string:
			key = t
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			key = fmt.Sprintf("%v", t)
		default:
			// 如果无法转换为字符串，则使用 reflect 获取字符串表示
			key = fmt.Sprintf("%v", t)
		}
		result[key] = v
	}
	return result
}

// ObjectToMap 将对象转换为 map
func ObjectToMap[V any](obj map[string]V) map[string]V {
	return obj
}

// CloneMap 克隆 map
func CloneMap[K comparable, V any](m map[K]V) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		result[k] = v
	}
	return result
}

// FlatClone 对对象进行浅拷贝
func FlatClone[T any](obj T) T {
	return obj
}

// EnsureNotFalsy 确保值不为空
func EnsureNotFalsy[T any](obj T) (T, error) {
	if reflect.ValueOf(obj).IsZero() {
		return obj, errors.New("ensureNotFalsy() is falsy")
	}
	return obj, nil
}

// MergeSets 合并多个集合
func MergeSets[T comparable](sets []map[T]struct{}) map[T]struct{} {
	result := make(map[T]struct{})
	for _, set := range sets {
		for item := range set {
			result[item] = struct{}{}
		}
	}
	return result
}

// RoundToTwoDecimals 将数字四舍五入到两位小数
func RoundToTwoDecimals(num float64) float64 {
	return math.Round(num*100) / 100
}

// IsObject 判断值是否为对象
func IsObject(value interface{}) bool {
	if value == nil {
		return false
	}

	kind := reflect.TypeOf(value).Kind()
	return kind == reflect.Map || kind == reflect.Struct || kind == reflect.Ptr
}

// GetProperty 获取对象的属性值
func GetProperty(obj map[string]interface{}, path string, defaultValue ...interface{}) interface{} {
	var value interface{}
	if len(defaultValue) > 0 {
		value = defaultValue[0]
	}

	if obj == nil || path == "" {
		if len(defaultValue) > 0 {
			return value
		}
		return obj
	}

	pathArray := strings.Split(path, ".")
	if len(pathArray) == 0 {
		return value
	}

	current := obj
	for i, key := range pathArray {
		// 如果当前对象为空，返回默认值
		if current == nil {
			return value
		}

		// 如果是数组索引
		if IsStringIndex(current, key) {
			if i == len(pathArray)-1 {
				return nil
			}
			return value
		}

		// 获取属性值
		if val, ok := current[key]; ok {
			if i == len(pathArray)-1 {
				return val
			}

			// 如果不是最后一个 key，且值是 map 类型，继续递归
			if nextMap, ok := val.(map[string]interface{}); ok {
				current = nextMap
			} else {
				// 如果值不是 map 类型，无法继续递归，返回默认值
				return value
			}
		} else {
			// 如果 key 不存在，返回默认值
			return value
		}
	}

	// 如果整个路径都已处理，返回当前值
	return current
}

// IsStringIndex 判断是否是字符串形式的数组索引
func IsStringIndex(obj map[string]interface{}, key string) bool {
	// 检查是否是数组且 key 是数字
	if arr, ok := obj["__array"].([]interface{}); ok {
		if index, err := strconv.Atoi(key); err == nil {
			return index >= 0 && index < len(arr)
		}
	}
	return false
}
