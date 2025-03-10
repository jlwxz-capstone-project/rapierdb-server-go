package query

import (
	"encoding/json"
	"sort"
)

func StableStringify[T Query](obj T) (string, error) {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}

	var parsed interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		return "", err
	}

	normalized := normalizeValue(parsed)

	result, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

func normalizeValue(v interface{}) interface{} {
	switch v := v.(type) {
	case map[string]interface{}:
		normalized := make(map[string]interface{})
		keys := make([]string, 0, len(v))

		for k := range v {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		for _, k := range keys {
			normalized[k] = normalizeValue(v[k])
		}

		return normalized

	case []interface{}:
		normalized := make([]interface{}, len(v))

		if isSortFieldArray(v) {
			type sortItem struct {
				field string
				value interface{}
			}

			items := make([]sortItem, 0, len(v))

			for _, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					if field, ok := m["field"].(string); ok {
						items = append(items, sortItem{
							field: field,
							value: normalizeValue(item),
						})
					}
				}
			}

			sort.Slice(items, func(i, j int) bool {
				return items[i].field < items[j].field
			})

			for i, item := range items {
				normalized[i] = item.value
			}
		} else {
			for i, item := range v {
				normalized[i] = normalizeValue(item)
			}
		}

		return normalized

	default:
		return v
	}
}

func isSortFieldArray(v []interface{}) bool {
	if len(v) == 0 {
		return false
	}

	if m, ok := v[0].(map[string]interface{}); ok {
		_, hasField := m["field"]
		_, hasOrder := m["order"]
		return hasField && hasOrder
	}

	return false
}
