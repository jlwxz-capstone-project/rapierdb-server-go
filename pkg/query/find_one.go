package query

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	qfe "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/query_filter_expr"
)

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

// DebugSprint 返回查询的调试字符串表示
// 实现 log.DebugPrintable 接口
func (q *FindOneQuery) DebugSprint() string {
	return fmt.Sprintf("FindOneQuery{Collection: %s, Filter: %s}", q.Collection, q.Filter.DebugSprint())
}

// SetFilter 设置查询的过滤条件
func (q *FindOneQuery) SetFilter(filter qfe.QueryFilterExpr) {
	q.Filter = filter
}

// Match 检查给定的文档是否匹配查询条件
// 返回值:
//   - 如果文档匹配查询条件，返回 true
//   - 如果发生错误，返回 false 和错误信息
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

func (q *FindOneQuery) Encode() ([]byte, error) {
	var temp struct {
		Type uint64 `json:"type"`
		*FindOneQuery
	}
	temp.Type = FIND_ONE_QUERY_TYPE
	temp.FindOneQuery = q

	jsonBytes, err := json.Marshal(temp)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}

func DecodeFindOneQuery(data []byte) (*FindOneQuery, error) {
	var temp struct {
		Collection string          `json:"collection"`
		Filter     json.RawMessage `json:"filter,omitempty"`
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
	return &FindOneQuery{
		Collection: temp.Collection,
		Filter:     filter,
	}, nil
}
