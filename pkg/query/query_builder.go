package query

import (
	qfe "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/query_filter_expr"
)

// QueryBuilder 提供链式调用方式来构建查询
type QueryBuilder struct {
	query *Query
}

// NewQueryBuilder 创建一个新的查询构建器
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		query: NewQuery(),
	}
}

// Build 返回构建好的查询对象
func (b *QueryBuilder) Build() *Query {
	return b.query
}

// Filter 设置过滤条件
func (b *QueryBuilder) Filter(filter qfe.QueryFilterExpr) *QueryBuilder {
	b.query.Filter = filter
	return b
}

// Sort 添加排序规则
func (b *QueryBuilder) Sort(field string, order SortOrder) *QueryBuilder {
	b.query.Sort = append(b.query.Sort, SortField{
		Field: field,
		Order: order,
	})
	return b
}

// Skip 设置跳过的文档数量
func (b *QueryBuilder) Skip(skip int64) *QueryBuilder {
	b.query.Skip = skip
	return b
}

// Limit 设置返回的最大文档数量
func (b *QueryBuilder) Limit(limit int64) *QueryBuilder {
	b.query.Limit = limit
	return b
}

// 以下是一些常用的过滤条件构建方法

// Eq 创建等于比较表达式
func Eq(field string, value any) qfe.QueryFilterExpr {
	return &qfe.EqExpr{
		O1: &qfe.FieldValueExpr{Path: field},
		O2: &qfe.ValueExpr{Value: value},
	}
}

// Ne 创建不等于比较表达式
func Ne(field string, value any) qfe.QueryFilterExpr {
	return &qfe.NeExpr{
		O1: &qfe.FieldValueExpr{Path: field},
		O2: &qfe.ValueExpr{Value: value},
	}
}

// Gt 创建大于比较表达式
func Gt(field string, value any) qfe.QueryFilterExpr {
	return &qfe.GtExpr{
		O1: &qfe.FieldValueExpr{Path: field},
		O2: &qfe.ValueExpr{Value: value},
	}
}

// Gte 创建大于等于比较表达式
func Gte(field string, value any) qfe.QueryFilterExpr {
	return &qfe.GteExpr{
		O1: &qfe.FieldValueExpr{Path: field},
		O2: &qfe.ValueExpr{Value: value},
	}
}

// Lt 创建小于比较表达式
func Lt(field string, value any) qfe.QueryFilterExpr {
	return &qfe.LtExpr{
		O1: &qfe.FieldValueExpr{Path: field},
		O2: &qfe.ValueExpr{Value: value},
	}
}

// Lte 创建小于等于比较表达式
func Lte(field string, value any) qfe.QueryFilterExpr {
	return &qfe.LteExpr{
		O1: &qfe.FieldValueExpr{Path: field},
		O2: &qfe.ValueExpr{Value: value},
	}
}

// In 创建包含比较表达式
func In(field string, values ...any) qfe.QueryFilterExpr {
	valueExprs := make([]qfe.QueryFilterExpr, len(values))
	for i, v := range values {
		valueExprs[i] = &qfe.ValueExpr{Value: v}
	}
	return &qfe.InExpr{
		O1: &qfe.FieldValueExpr{Path: field},
		O2: valueExprs,
	}
}

// Nin 创建不包含比较表达式
func Nin(field string, values ...any) qfe.QueryFilterExpr {
	valueExprs := make([]qfe.QueryFilterExpr, len(values))
	for i, v := range values {
		valueExprs[i] = &qfe.ValueExpr{Value: v}
	}
	return &qfe.NinExpr{
		O1: &qfe.FieldValueExpr{Path: field},
		O2: valueExprs,
	}
}

// Exists 创建字段存在检查表达式
func Exists(field string) qfe.QueryFilterExpr {
	return &qfe.ExistsExpr{
		Field: &qfe.FieldValueExpr{Path: field},
	}
}

// Regex 创建正则表达式匹配表达式
func Regex(field string, pattern string) qfe.QueryFilterExpr {
	return &qfe.RegexExpr{
		O1:    &qfe.FieldValueExpr{Path: field},
		Regex: pattern,
	}
}

// And 创建逻辑与表达式
func And(exprs ...qfe.QueryFilterExpr) qfe.QueryFilterExpr {
	return &qfe.AndExpr{Exprs: exprs}
}

// Or 创建逻辑或表达式
func Or(exprs ...qfe.QueryFilterExpr) qfe.QueryFilterExpr {
	return &qfe.OrExpr{Exprs: exprs}
}

// Not 创建逻辑非表达式
func Not(expr qfe.QueryFilterExpr) qfe.QueryFilterExpr {
	return &qfe.NotExpr{Expr: expr}
}
