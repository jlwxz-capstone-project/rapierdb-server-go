package main

import (
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestQueryBuilder(t *testing.T) {
	// 创建测试文档
	doc := loro.NewLoroDoc()
	root := doc.GetMap("root")
	root.InsertString("name", "Alice")
	root.InsertI64("age", 25)
	root.InsertDouble("score", 85.5)

	tests := []struct {
		name     string
		builder  func() *query.Query
		expected bool
	}{
		{
			"简单等于比较",
			func() *query.Query {
				return query.NewQueryBuilder().
					Filter(query.Eq("root/name", "Alice")).
					Build()
			},
			true,
		},
		{
			"复合条件 AND",
			func() *query.Query {
				return query.NewQueryBuilder().
					Filter(query.And(
						query.Gt("root/age", int64(20)),
						query.Lt("root/age", int64(30)),
					)).
					Build()
			},
			true,
		},
		{
			"复合条件 OR",
			func() *query.Query {
				return query.NewQueryBuilder().
					Filter(query.Or(
						query.Eq("root/name", "Alice"),
						query.Eq("root/name", "Bob"),
					)).
					Build()
			},
			true,
		},
		{
			"范围查询",
			func() *query.Query {
				return query.NewQueryBuilder().
					Filter(query.And(
						query.Gte("root/score", 80.0),
						query.Lte("root/score", 90.0),
					)).
					Build()
			},
			true,
		},
		{
			"IN 查询",
			func() *query.Query {
				return query.NewQueryBuilder().
					Filter(query.In("root/name", "Alice", "Bob", "Charlie")).
					Build()
			},
			true,
		},
		{
			"NOT IN 查询",
			func() *query.Query {
				return query.NewQueryBuilder().
					Filter(query.Nin("root/name", "Bob", "Charlie")).
					Build()
			},
			true,
		},
		{
			"EXISTS 查询",
			func() *query.Query {
				return query.NewQueryBuilder().
					Filter(query.Exists("root/name")).
					Build()
			},
			true,
		},
		{
			"REGEX 查询",
			func() *query.Query {
				return query.NewQueryBuilder().
					Filter(query.Regex("root/name", "^Al.*$")).
					Build()
			},
			true,
		},
		{
			"NOT 查询",
			func() *query.Query {
				return query.NewQueryBuilder().
					Filter(query.Not(query.Eq("root/name", "Bob"))).
					Build()
			},
			true,
		},
		{
			"带排序和分页的查询",
			func() *query.Query {
				return query.NewQueryBuilder().
					Filter(query.Gt("root/age", int64(20))).
					Sort("root/age", query.SortOrderAsc).
					Sort("root/score", query.SortOrderDesc).
					Skip(0).
					Limit(10).
					Build()
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.builder()

			// 验证查询条件
			matched, err := q.Match(doc)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, matched)

			// 验证查询是否可以序列化和反序列化
			data, err := q.MarshalJSON()
			assert.NoError(t, err)

			query2 := query.NewQuery()
			err = query2.UnmarshalJSON(data)
			assert.NoError(t, err)

			matched2, err := query2.Match(doc)
			assert.NoError(t, err)
			assert.Equal(t, matched, matched2)
		})
	}
}

// 测试错误情况
func TestQueryBuilderErrors(t *testing.T) {
	doc := loro.NewLoroDoc()
	root := doc.GetMap("root")
	root.InsertString("name", "Alice")

	tests := []struct {
		name        string
		builder     func() *query.Query
		expectedErr string
	}{
		{
			"字段不存在",
			func() *query.Query {
				return query.NewQueryBuilder().
					Filter(query.Eq("root/nonexistent", "value")).
					Build()
			},
			"field error",
		},
		{
			"类型不匹配",
			func() *query.Query {
				return query.NewQueryBuilder().
					Filter(query.Eq("root/name", 42)). // 字符串字段和数字比较
					Build()
			},
			"type error",
		},
		{
			"无效的正则表达式",
			func() *query.Query {
				return query.NewQueryBuilder().
					Filter(query.Regex("root/name", "[")).
					Build()
			},
			"invalid regex pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := tt.builder()
			_, err := query.Match(doc)
			assert.Error(t, err)
			assert.ErrorContains(t, err, tt.expectedErr)
		})
	}
}
