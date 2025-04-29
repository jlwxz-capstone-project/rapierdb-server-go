package query

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js_value"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/doc_visitor"
	qfe "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/query_filter_expr"
)

// FindManyQuery 表示一个仅用于查询多个文档的查询
// Collection: 集合名称
// Filter: 过滤条件
// Sort: 排序规则
// Skip: 跳过的文档数量
// Limit: 返回的最大文档数量
type FindManyQuery struct {
	Collection string              `json:"collection"`       // 集合名称
	Filter     qfe.QueryFilterExpr `json:"filter,omitempty"` // 过滤条件
	Sort       []SortField         `json:"sort,omitempty"`   // 排序规则
	Skip       int64               `json:"skip,omitempty"`   // 跳过的文档数量
	Limit      int64               `json:"limit,omitempty"`  // 返回的最大文档数量
}

var _ Query = &FindManyQuery{}

func (q *FindManyQuery) isQuery() {}

// DebugSprint 返回查询的调试字符串表示
// 实现 log.DebugPrintable 接口
func (q *FindManyQuery) DebugSprint() string {
	filterStr := q.Filter.DebugSprint()
	sortStr := make([]string, len(q.Sort))
	for i, sort := range q.Sort {
		sortStr[i] = sort.DebugSprint()
	}
	return fmt.Sprintf("FindManyQuery{Collection: %s, Filter: %s, Sort: [%s], Skip: %d, Limit: %d}", q.Collection, filterStr, strings.Join(sortStr, ", "), q.Skip, q.Limit)
}

// SetFilter 设置查询的过滤条件
func (q *FindManyQuery) SetFilter(filter qfe.QueryFilterExpr) {
	q.Filter = filter
}

// AddSort 添加一个排序规则
// 参数:
//   - field: 要排序的字段名
//   - order: 排序顺序（升序或降序）
func (q *FindManyQuery) AddSort(field string, order SortOrder) {
	q.Sort = append(q.Sort, SortField{
		Field: field,
		Order: order,
	})
}

// SetSkip 设置要跳过的文档数量
// 参数:
//   - skip: 要跳过的文档数量，必须为非负数
//
// 返回值:
//   - 如果 skip 为负数，返回错误
func (q *FindManyQuery) SetSkip(skip int64) error {
	if skip < 0 {
		return fmt.Errorf("skip must be non-negative")
	}
	q.Skip = skip
	return nil
}

// SetLimit 设置要返回的最大文档数量
// 参数:
//   - limit: 要返回的最大文档数量，必须为非负数
//
// 返回值:
//   - 如果 limit 为负数，返回错误
func (q *FindManyQuery) SetLimit(limit int64) error {
	if limit < 0 {
		return fmt.Errorf("limit must be non-negative")
	}
	q.Limit = limit
	return nil
}

// Match 检查给定的文档是否匹配查询条件
// 参数:
//   - doc: 要检查的文档
//
// 返回值:
//   - 如果文档匹配查询条件，返回 true
//   - 如果发生错误，返回 false 和错误信息
func (q *FindManyQuery) Match(doc *loro.LoroDoc) (bool, error) {
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
// 参数:
//   - doc1: 第一个文档
//   - doc2: 第二个文档
//
// 返回值：
//   - 如果 doc1 < doc2，返回 -1
//   - 如果 doc1 = doc2，返回 0
//   - 如果 doc1 > doc2，返回 1
//   - 如果发生错误，返回错误信息
func (q *FindManyQuery) Compare(doc1, doc2 *loro.LoroDoc) (int, error) {
	for _, sort := range q.Sort {
		// 获取字段值
		v1, err := doc_visitor.VisitDocByPath(doc1, sort.Field)
		if err != nil {
			return 0, err
		}

		v2, err := doc_visitor.VisitDocByPath(doc2, sort.Field)
		if err != nil {
			return 0, err
		}

		cmp, err := js_value.DeepComapreJsValue(v1, v2)
		if err != nil {
			return 0, err
		}

		if cmp != 0 {
			if sort.Order == SortOrderAsc {
				return cmp, nil
			} else {
				return -cmp, nil
			}
		}
	}

	return 0, nil
}

func (q *FindManyQuery) Encode() ([]byte, error) {
	var temp struct {
		Type uint64 `json:"type"`
		*FindManyQuery
	}
	temp.Type = FIND_MANY_QUERY_TYPE
	temp.FindManyQuery = q

	jsonBytes, err := json.Marshal(temp)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}

func DecodeFindManyQuery(data []byte) (*FindManyQuery, error) {
	var temp struct {
		Collection string          `json:"collection"`
		Filter     json.RawMessage `json:"filter,omitempty"`
		Sort       []SortField     `json:"sort,omitempty"`
		Skip       int64           `json:"skip,omitempty"`
		Limit      int64           `json:"limit,omitempty"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return nil, err
	}
	var filter qfe.QueryFilterExpr
	if len(temp.Filter) > 0 {
		var err error
		filter, err = qfe.NewQueryFilterExprFromJson(temp.Filter)
		if err != nil {
			return nil, err
		}
	}
	return &FindManyQuery{
		Collection: temp.Collection,
		Filter:     filter,
		Sort:       temp.Sort,
		Skip:       temp.Skip,
		Limit:      temp.Limit,
	}, nil
}
