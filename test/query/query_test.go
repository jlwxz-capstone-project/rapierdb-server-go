package main

import (
	"bytes"
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	qfe "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/query_filter_expr"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
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
			query := query.NewFindOneQuery()
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
			query := query.NewFindManyQuery()
			for _, sort := range tt.sorts {
				query.AddSort(sort.Field, sort.Order)
			}

			cmp, err := query.Compare(tt.doc1, tt.doc2)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, cmp)
		})
	}
}

func TestQueryPagination(t *testing.T) {
	query := query.NewFindManyQuery()

	err := query.SetSkip(10)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), query.Skip)

	err = query.SetLimit(20)
	assert.NoError(t, err)
	assert.Equal(t, int64(20), query.Limit)

	err = query.SetSkip(-1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "skip must be non-negative")

	err = query.SetLimit(-1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "limit must be non-negative")
}

func TestQueryEncodeDecode(t *testing.T) {
	t.Run("FindOneQuery", func(t *testing.T) {
		query1 := query.NewFindOneQuery()

		fieldExpr := &qfe.FieldValueExpr{Path: "name"}
		valueExpr := &qfe.ValueExpr{Value: "张三"}
		eqExpr := &qfe.EqExpr{O1: fieldExpr, O2: valueExpr}

		query1.SetFilter(eqExpr)

		encoded, err := query1.Encode()
		assert.NoError(t, err)
		assert.NotNil(t, encoded)

		decoded, err := query.DecodeFindOneQuery(encoded)
		assert.NoError(t, err)
		assert.NotNil(t, decoded)

		reEncoded, err := decoded.Encode()
		assert.NoError(t, err)
		assert.Equal(t, encoded, reEncoded)
	})

	t.Run("FindManyQuery", func(t *testing.T) {
		query1 := query.NewFindManyQuery()

		fieldExpr := &qfe.FieldValueExpr{Path: "age"}
		valueExpr := &qfe.ValueExpr{Value: int64(18)}
		gtExpr := &qfe.GtExpr{O1: fieldExpr, O2: valueExpr}

		query1.SetFilter(gtExpr)

		query1.AddSort("age", query.SortOrderDesc)
		query1.AddSort("name", query.SortOrderAsc)

		err := query1.SetSkip(10)
		assert.NoError(t, err)

		err = query1.SetLimit(20)
		assert.NoError(t, err)

		encoded, err := query1.Encode()
		assert.NoError(t, err)
		assert.NotNil(t, encoded)

		decoded, err := query.DecodeFindManyQuery(encoded)
		assert.NoError(t, err)
		assert.NotNil(t, decoded)

		reEncoded, err := decoded.Encode()
		assert.NoError(t, err)
		assert.Equal(t, encoded, reEncoded)

		assert.Equal(t, int64(10), decoded.Skip)
		assert.Equal(t, int64(20), decoded.Limit)
		assert.Equal(t, 2, len(decoded.Sort))
		assert.Equal(t, "age", decoded.Sort[0].Field)
		assert.Equal(t, query.SortOrderDesc, decoded.Sort[0].Order)
		assert.Equal(t, "name", decoded.Sort[1].Field)
		assert.Equal(t, query.SortOrderAsc, decoded.Sort[1].Order)
	})

	t.Run("InvalidQueryType", func(t *testing.T) {
		buf := &bytes.Buffer{}
		util.WriteVarUint(buf, uint64(999))
		util.WriteBytes(buf, []byte("{}"))

		_, err := query.DecodeFindOneQuery(buf.Bytes())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid query type")

		_, err = query.DecodeFindManyQuery(buf.Bytes())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid query type")
	})
}

func TestStableStringify(t *testing.T) {
	t.Run("FindOneQuery", func(t *testing.T) {
		query1 := query.NewFindOneQuery()

		fieldExpr := &qfe.FieldValueExpr{Path: "name"}
		valueExpr := &qfe.ValueExpr{Value: "张三"}
		eqExpr := &qfe.EqExpr{O1: fieldExpr, O2: valueExpr}

		query1.SetFilter(eqExpr)

		str1, err := query.StableStringify(query1)
		assert.NoError(t, err)

		query2 := query.NewFindOneQuery()
		query2.SetFilter(eqExpr)

		str2, err := query.StableStringify(query2)
		assert.NoError(t, err)

		assert.Equal(t, str1, str2)
	})

	t.Run("FindManyQuery", func(t *testing.T) {
		query1 := query.NewFindManyQuery()

		fieldExpr := &qfe.FieldValueExpr{Path: "age"}
		valueExpr := &qfe.ValueExpr{Value: int64(18)}
		gtExpr := &qfe.GtExpr{O1: fieldExpr, O2: valueExpr}

		query1.SetFilter(gtExpr)

		query1.AddSort("age", query.SortOrderDesc)
		query1.AddSort("name", query.SortOrderAsc)

		query1.SetSkip(10)
		query1.SetLimit(20)

		str1, err := query.StableStringify(query1)
		assert.NoError(t, err)

		query2 := query.NewFindManyQuery()
		query2.SetLimit(20)
		query2.AddSort("name", query.SortOrderAsc)
		query2.SetFilter(gtExpr)
		query2.SetSkip(10)
		query2.AddSort("age", query.SortOrderDesc)

		str2, err := query.StableStringify(query2)
		assert.NoError(t, err)

		assert.Equal(t, str1, str2)
	})
}
