package main

import (
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	qfe "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/query_filter_expr"
	"github.com/stretchr/testify/assert"
)

// 创建测试文档
func createTestDoc2(name string, age int64, score float64) *loro.LoroDoc {
	doc := loro.NewLoroDoc()
	root := doc.GetMap("root")
	root.InsertString("name", name)
	root.InsertI64("age", age)
	root.InsertDouble("score", score)
	return doc
}

// 测试基本的过滤功能
func TestQueryFilter(t *testing.T) {
	doc1 := createTestDoc2("Alice", 25, 85.5)
	doc2 := createTestDoc2("Bob", 30, 92.0)

	tests := []struct {
		name     string
		filter   qfe.QueryFilterExpr
		doc      *loro.LoroDoc
		expected bool
	}{
		{
			"年龄等于",
			&qfe.EqExpr{
				O1: &qfe.FieldValueExpr{Path: "root/age"},
				O2: &qfe.ValueExpr{Value: int64(25)},
			},
			doc1,
			true,
		},
		{
			"年龄大于",
			&qfe.GtExpr{
				O1: &qfe.FieldValueExpr{Path: "root/age"},
				O2: &qfe.ValueExpr{Value: int64(25)},
			},
			doc2,
			true,
		},
		{
			"分数范围",
			&qfe.AndExpr{
				Exprs: []qfe.QueryFilterExpr{
					&qfe.GteExpr{
						O1: &qfe.FieldValueExpr{Path: "root/score"},
						O2: &qfe.ValueExpr{Value: 80.0},
					},
					&qfe.LtExpr{
						O1: &qfe.FieldValueExpr{Path: "root/score"},
						O2: &qfe.ValueExpr{Value: 90.0},
					},
				},
			},
			doc1,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := query.NewQuery()
			query.SetFilter(tt.filter)

			matched, err := query.Match(tt.doc)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, matched)
		})
	}
}

// 测试排序功能
func TestQuerySort(t *testing.T) {
	doc1 := createTestDoc2("Alice", 25, 85.5)
	doc2 := createTestDoc2("Bob", 30, 92.0)
	doc3 := createTestDoc2("Charlie", 25, 88.0)

	tests := []struct {
		name     string
		sorts    []query.SortField
		doc1     *loro.LoroDoc
		doc2     *loro.LoroDoc
		expected int
	}{
		{
			"按年龄升序",
			[]query.SortField{{Field: "root/age", Order: query.SortOrderAsc}},
			doc1,
			doc2,
			-1,
		},
		{
			"按年龄降序",
			[]query.SortField{{Field: "root/age", Order: query.SortOrderDesc}},
			doc1,
			doc2,
			1,
		},
		{
			"按年龄和分数排序",
			[]query.SortField{
				{Field: "root/age", Order: query.SortOrderAsc},
				{Field: "root/score", Order: query.SortOrderDesc},
			},
			doc1,
			doc3,
			-1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := query.NewQuery()
			for _, sort := range tt.sorts {
				query.AddSort(sort.Field, sort.Order)
			}

			cmp, err := query.Compare(tt.doc1, tt.doc2)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, cmp)
		})
	}
}

// 测试分页参数
func TestQueryPagination(t *testing.T) {
	query := query.NewQuery()

	// 测试有效的分页参数
	err := query.SetSkip(10)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), query.Skip)

	err = query.SetLimit(20)
	assert.NoError(t, err)
	assert.Equal(t, int64(20), query.Limit)

	// 测试无效的分页参数
	err = query.SetSkip(-1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "skip must be non-negative")

	err = query.SetLimit(-1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "limit must be non-negative")
}

// 测试 JSON 序列化和反序列化
func TestQuerySerialization(t *testing.T) {
	q := query.NewQuery()

	// 设置过滤条件
	q.SetFilter(&qfe.GtExpr{
		O1: &qfe.FieldValueExpr{Path: "root/age"},
		O2: &qfe.ValueExpr{Value: int64(25)},
	})

	// 添加排序规则
	q.AddSort("root/age", query.SortOrderAsc)
	q.AddSort("root/score", query.SortOrderDesc)

	// 设置分页参数
	q.SetSkip(10)
	q.SetLimit(20)

	// 序列化
	data, err := q.MarshalJSON()
	assert.NoError(t, err)

	// 反序列化
	query2 := query.NewQuery()
	err = query2.UnmarshalJSON(data)
	assert.NoError(t, err)

	// 验证反序列化后的结果
	assert.Equal(t, q.Skip, query2.Skip)
	assert.Equal(t, q.Limit, query2.Limit)
	assert.Equal(t, len(q.Sort), len(query2.Sort))
	for i := range q.Sort {
		assert.Equal(t, q.Sort[i].Field, query2.Sort[i].Field)
		assert.Equal(t, q.Sort[i].Order, query2.Sort[i].Order)
	}
}
