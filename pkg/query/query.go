package query

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	qfe "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/query_filter_expr"
)

// SortOrder 表示排序顺序
type SortOrder int

const (
	SortOrderAsc  SortOrder = 1  // 升序
	SortOrderDesc SortOrder = -1 // 降序
)

// SortField 表示排序字段
type SortField struct {
	Field string    `json:"field"` // 字段路径
	Order SortOrder `json:"order"` // 排序顺序
}

// Query 表示一个完整的查询
type Query struct {
	Filter qfe.QueryFilterExpr `json:"filter,omitempty"` // 过滤条件
	Sort   []SortField         `json:"sort,omitempty"`   // 排序规则
	Skip   int64               `json:"skip,omitempty"`   // 跳过的文档数量
	Limit  int64               `json:"limit,omitempty"`  // 返回的最大文档数量
}

// NewQuery 创建一个新的查询
func NewQuery() *Query {
	return &Query{
		Sort: make([]SortField, 0),
	}
}

// SetFilter 设置过滤条件
func (q *Query) SetFilter(filter qfe.QueryFilterExpr) {
	q.Filter = filter
}

// AddSort 添加排序规则
func (q *Query) AddSort(field string, order SortOrder) {
	q.Sort = append(q.Sort, SortField{
		Field: field,
		Order: order,
	})
}

// SetSkip 设置跳过的文档数量
func (q *Query) SetSkip(skip int64) error {
	if skip < 0 {
		return fmt.Errorf("skip must be non-negative")
	}
	q.Skip = skip
	return nil
}

// SetLimit 设置返回的最大文档数量
func (q *Query) SetLimit(limit int64) error {
	if limit < 0 {
		return fmt.Errorf("limit must be non-negative")
	}
	q.Limit = limit
	return nil
}

// Match 检查文档是否匹配查询条件
func (q *Query) Match(doc *loro.LoroDoc) (bool, error) {
	if q.Filter == nil {
		return true, nil
	}

	result, err := q.Filter.Eval(doc)
	if err != nil {
		return false, fmt.Errorf("evaluating filter: %v", err)
	}

	matched, ok := result.Value.(bool)
	if !ok {
		return false, fmt.Errorf("filter result must be boolean, got %T", result.Value)
	}

	return matched, nil
}

// Compare 比较两个文档在排序规则下的顺序
// 返回值：
//   - 如果 doc1 < doc2，返回 -1
//   - 如果 doc1 = doc2，返回 0
//   - 如果 doc1 > doc2，返回 1
func (q *Query) Compare(doc1, doc2 *loro.LoroDoc) (int, error) {
	for _, sort := range q.Sort {
		// 获取字段值
		field1 := doc1.GetMap("root").Get(sort.Field)
		field2 := doc2.GetMap("root").Get(sort.Field)

		// 获取实际值
		value1, err := field1.ToGoObject()
		if err != nil {
			return 0, fmt.Errorf("getting value from first document: %v", err)
		}

		value2, err := field2.ToGoObject()
		if err != nil {
			return 0, fmt.Errorf("getting value from second document: %v", err)
		}

		// 比较字段值
		cmp, err := qfe.CompareValues(value1, value2)
		if err != nil {
			return 0, fmt.Errorf("comparing field '%s': %v", sort.Field, err)
		}

		// 如果字段值不相等，根据排序顺序返回比较结果
		if cmp != 0 {
			if sort.Order == SortOrderDesc {
				cmp = -cmp
			}
			return cmp, nil
		}
	}

	// 所有字段都相等
	return 0, nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (q *Query) MarshalJSON() ([]byte, error) {
	type Alias Query
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(q),
	})
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (q *Query) UnmarshalJSON(data []byte) error {
	// 创建一个临时结构体来存储原始数据
	var raw struct {
		Filter json.RawMessage `json:"filter,omitempty"`
		Sort   []SortField     `json:"sort,omitempty"`
		Skip   int64           `json:"skip,omitempty"`
		Limit  int64           `json:"limit,omitempty"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// 反序列化 Filter
	if raw.Filter != nil {
		filter, err := qfe.UnmarshalQueryFilterExpr(raw.Filter)
		if err != nil {
			return fmt.Errorf("unmarshaling filter: %v", err)
		}
		q.Filter = filter
	}

	// 反序列化其他字段
	q.Sort = raw.Sort
	q.Skip = raw.Skip
	q.Limit = raw.Limit

	return nil
}
