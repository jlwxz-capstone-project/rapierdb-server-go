package query

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	qfe "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/query_filter_expr"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
	pe "github.com/pkg/errors"
)

const (
	FIND_MANY_QUERY_TYPE uint64 = 1
	FIND_ONE_QUERY_TYPE  uint64 = 2
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

func (s *SortField) DebugPrint() string {
	orderStr := "asc"
	if s.Order == SortOrderDesc {
		orderStr = "desc"
	}
	return fmt.Sprintf("SortField{Field: %s, Order: %s}", s.Field, orderStr)
}

type Query interface {
	log.DebugPrintable
	isQuery()
	Encode() ([]byte, error)
}

// FindOneQuery 表示一个仅用于查询单个文档的查询
//
//	Collection: 集合名称
//	Filter: 过滤条件
type FindOneQuery struct {
	Collection string              `json:"collection"`       // 集合名称
	Filter     qfe.QueryFilterExpr `json:"filter,omitempty"` // 过滤条件
}

var _ Query = &FindOneQuery{}

func (q *FindOneQuery) isQuery() {}

func (q *FindOneQuery) DebugPrint() string {
	return fmt.Sprintf("FindOneQuery{Collection: %s, Filter: %s}", q.Collection, q.Filter.DebugPrint())
}

// FindManyQuery 表示一个仅用于查询多个文档的查询
//
//	Collection: 集合名称
//	Filter: 过滤条件
//	Sort: 排序规则
//	Skip: 跳过的文档数量
//	Limit: 返回的最大文档数量
type FindManyQuery struct {
	Collection string              `json:"collection"`       // 集合名称
	Filter     qfe.QueryFilterExpr `json:"filter,omitempty"` // 过滤条件
	Sort       []SortField         `json:"sort,omitempty"`   // 排序规则
	Skip       int64               `json:"skip,omitempty"`   // 跳过的文档数量
	Limit      int64               `json:"limit,omitempty"`  // 返回的最大文档数量
}

var _ Query = &FindManyQuery{}

func (q *FindManyQuery) isQuery() {}

func (q *FindManyQuery) DebugPrint() string {
	filterStr := q.Filter.DebugPrint()
	sortStr := make([]string, len(q.Sort))
	for i, sort := range q.Sort {
		sortStr[i] = sort.DebugPrint()
	}
	return fmt.Sprintf("FindManyQuery{Collection: %s, Filter: %s, Sort: [%s], Skip: %d, Limit: %d}", q.Collection, filterStr, strings.Join(sortStr, ", "), q.Skip, q.Limit)
}

func NewFindOneQuery() *FindOneQuery {
	return &FindOneQuery{}
}

func NewFindManyQuery() *FindManyQuery {
	return &FindManyQuery{}
}

func (q *FindOneQuery) SetFilter(filter qfe.QueryFilterExpr) {
	q.Filter = filter
}

func (q *FindManyQuery) SetFilter(filter qfe.QueryFilterExpr) {
	q.Filter = filter
}

func (q *FindManyQuery) AddSort(field string, order SortOrder) {
	q.Sort = append(q.Sort, SortField{
		Field: field,
		Order: order,
	})
}

func (q *FindManyQuery) SetSkip(skip int64) error {
	if skip < 0 {
		return fmt.Errorf("skip must be non-negative")
	}
	q.Skip = skip
	return nil
}

func (q *FindManyQuery) SetLimit(limit int64) error {
	if limit < 0 {
		return fmt.Errorf("limit must be non-negative")
	}
	q.Limit = limit
	return nil
}

func (q *FindOneQuery) Match(doc *loro.LoroDoc) (bool, error) {
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
// 返回值：
//   - 如果 doc1 < doc2，返回 -1
//   - 如果 doc1 = doc2，返回 0
//   - 如果 doc1 > doc2，返回 1
//
// TODO 不支持深度排序，比如按 `a.b.c` 排序
func (q *FindManyQuery) Compare(doc1, doc2 *loro.LoroDoc) (int, error) {
	for _, sort := range q.Sort {
		// 获取字段值
		field1 := doc1.GetDataMap().Get(sort.Field)
		field2 := doc2.GetDataMap().Get(sort.Field)

		if field1 == nil {
			return 0, pe.WithStack(fmt.Errorf("field1 %s not found", sort.Field))
		}
		if field2 == nil {
			return 0, pe.WithStack(fmt.Errorf("field2 %s not found", sort.Field))
		}

		if field1.IsContainer() || field2.IsContainer() {
			return 0, pe.WithStack(fmt.Errorf("container type not supported for sorting"))
		}

		field1Value, err := field1.GetValue()
		if err != nil {
			return 0, pe.WithStack(fmt.Errorf("getting value from first document: %v", err))
		}
		if !field1Value.IsComparable() {
			return 0, pe.WithStack(fmt.Errorf("comparable type required for sorting"))
		}

		field2Value, err := field2.GetValue()
		if err != nil {
			return 0, pe.WithStack(fmt.Errorf("getting value from second document: %v", err))
		}
		if !field2Value.IsComparable() {
			return 0, pe.WithStack(fmt.Errorf("comparable type required for sorting"))
		}

		cmp, err := field1Value.Compare(field2Value)
		if err != nil {
			return 0, pe.WithStack(fmt.Errorf("comparing values: %v", err))
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

func (q *FindOneQuery) MarshalJSON() ([]byte, error) {
	type Alias FindOneQuery
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(q),
	})
}

func (q *FindOneQuery) UnmarshalJSON(data []byte) error {
	var raw struct {
		Filter json.RawMessage `json:"filter,omitempty"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// 反序列化 Filter
	if raw.Filter != nil {
		filter, err := qfe.NewQueryFilterExprFromJson(raw.Filter)
		if err != nil {
			return fmt.Errorf("unmarshaling filter: %v", err)
		}
		q.Filter = filter
	}

	return nil
}

// FindManyQuery 的序列化方法
func (q *FindManyQuery) MarshalJSON() ([]byte, error) {
	type Alias FindManyQuery
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(q),
	})
}

// FindManyQuery 的反序列化方法
func (q *FindManyQuery) UnmarshalJSON(data []byte) error {
	var raw struct {
		Collection string          `json:"collection"`
		Filter     json.RawMessage `json:"filter,omitempty"`
		Sort       []SortField     `json:"sort,omitempty"`
		Skip       int64           `json:"skip,omitempty"`
		Limit      int64           `json:"limit,omitempty"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// 反序列化 Filter
	if raw.Filter != nil {
		filter, err := qfe.NewQueryFilterExprFromJson(raw.Filter)
		if err != nil {
			return fmt.Errorf("unmarshaling filter: %v", err)
		}
		q.Filter = filter
	}

	// 反序列化其他字段
	q.Collection = raw.Collection
	q.Sort = raw.Sort
	q.Skip = raw.Skip
	q.Limit = raw.Limit

	return nil
}

func (q *FindOneQuery) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	util.WriteVarUint(buf, FIND_ONE_QUERY_TYPE)
	json, err := q.MarshalJSON()
	if err != nil {
		return nil, err
	}
	util.WriteBytes(buf, json)
	return buf.Bytes(), nil
}

func (q *FindManyQuery) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	util.WriteVarUint(buf, FIND_MANY_QUERY_TYPE)
	json, err := q.MarshalJSON()
	if err != nil {
		return nil, err
	}
	util.WriteBytes(buf, json)
	return buf.Bytes(), nil
}

func decodeFindOneQueryBody(data []byte) (*FindOneQuery, error) {
	query := NewFindOneQuery()
	err := query.UnmarshalJSON(data)
	if err != nil {
		return nil, err
	}
	return query, nil
}

func decodeFindManyQueryBody(data []byte) (*FindManyQuery, error) {
	query := NewFindManyQuery()
	err := query.UnmarshalJSON(data)
	if err != nil {
		return nil, err
	}
	return query, nil
}

func DecodeQuery(data []byte) (Query, error) {
	buf := bytes.NewBuffer(data)
	queryType, err := util.ReadVarUint(buf)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := util.ReadBytes(buf)
	if err != nil {
		return nil, err
	}

	if queryType == FIND_ONE_QUERY_TYPE {
		return decodeFindOneQueryBody(bodyBytes)
	} else if queryType == FIND_MANY_QUERY_TYPE {
		return decodeFindManyQueryBody(bodyBytes)
	} else {
		return nil, pe.WithStack(fmt.Errorf("invalid query type: %d", queryType))
	}
}

func DecodeFindOneQuery(data []byte) (*FindOneQuery, error) {
	query, err := DecodeQuery(data)
	if err != nil {
		return nil, err
	}
	findOneQuery, ok := query.(*FindOneQuery)
	if !ok {
		return nil, pe.WithStack(fmt.Errorf("invalid query type: %T", query))
	}
	return findOneQuery, nil
}

func DecodeFindManyQuery(data []byte) (*FindManyQuery, error) {
	query, err := DecodeQuery(data)
	if err != nil {
		return nil, err
	}
	findManyQuery, ok := query.(*FindManyQuery)
	if !ok {
		return nil, pe.WithStack(fmt.Errorf("invalid query type: %T", query))
	}
	return findManyQuery, nil
}
