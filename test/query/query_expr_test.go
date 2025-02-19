package main

import (
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	qfe "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/query_filter_expr"
	"github.com/stretchr/testify/assert"
)

// 辅助函数：创建一个测试用的 LoroDoc
func createTestDoc() *loro.LoroDoc {
	doc := loro.NewLoroDoc()
	root := doc.GetMap("root")

	// 添加基本类型
	root.InsertNull("nullField")
	root.InsertBool("boolField", true)
	root.InsertI64("intField", 42)
	root.InsertDouble("doubleField", 3.14)
	root.InsertString("stringField", "hello")

	// 添加数组
	list := loro.NewEmptyLoroList()
	list.PushI64(1)
	list.PushI64(2)
	list.PushI64(3)
	root.InsertList("listField", list)

	// 添加嵌套对象
	nestedMap := loro.NewEmptyLoroMap()
	nestedMap.InsertString("nestedKey", "nestedValue")
	root.InsertMap("mapField", nestedMap)

	return doc
}

// 测试基本的相等比较
func TestEqExpr(t *testing.T) {
	doc := createTestDoc()

	tests := []struct {
		name     string
		path     string
		value    any
		expected bool
	}{
		{"比较整数相等", "root/intField", int64(42), true},
		{"比较整数不相等", "root/intField", int64(43), false},
		{"比较浮点数相等", "root/doubleField", 3.14, true},
		{"比较浮点数不相等", "root/doubleField", 3.15, false},
		{"比较字符串相等", "root/stringField", "hello", true},
		{"比较字符串不相等", "root/stringField", "world", false},
		{"比较布尔值相等", "root/boolField", true, true},
		{"比较布尔值不相等", "root/boolField", false, false},
		{"比较null值", "root/nullField", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fieldExpr := &qfe.FieldValueExpr{Path: tt.path}
			valueExpr := &qfe.ValueExpr{Value: tt.value}
			eqExpr := &qfe.EqExpr{O1: fieldExpr, O2: valueExpr}

			result, err := eqExpr.Eval(doc)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Value)
		})
	}
}

// 测试不等比较
func TestNeExpr(t *testing.T) {
	doc := createTestDoc()

	tests := []struct {
		name     string
		path     string
		value    any
		expected bool
	}{
		{"比较整数不等", "root/intField", int64(43), true},
		{"比较整数相等", "root/intField", int64(42), false},
		{"比较浮点数不等", "root/doubleField", 3.15, true},
		{"比较浮点数相等", "root/doubleField", 3.14, false},
		{"比较字符串不等", "root/stringField", "world", true},
		{"比较字符串相等", "root/stringField", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fieldExpr := &qfe.FieldValueExpr{Path: tt.path}
			valueExpr := &qfe.ValueExpr{Value: tt.value}
			neExpr := &qfe.NeExpr{O1: fieldExpr, O2: valueExpr}

			result, err := neExpr.Eval(doc)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Value)
		})
	}
}

// 测试大于比较
func TestGtExpr(t *testing.T) {
	doc := createTestDoc()

	tests := []struct {
		name     string
		path     string
		value    any
		expected bool
	}{
		{"整数大于", "root/intField", int64(41), true},
		{"整数不大于", "root/intField", int64(42), false},
		{"浮点数大于", "root/doubleField", 3.13, true},
		{"浮点数不大于", "root/doubleField", 3.14, false},
		{"字符串大于", "root/stringField", "hella", true},
		{"字符串不大于", "root/stringField", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fieldExpr := &qfe.FieldValueExpr{Path: tt.path}
			valueExpr := &qfe.ValueExpr{Value: tt.value}
			gtExpr := &qfe.GtExpr{O1: fieldExpr, O2: valueExpr}

			result, err := gtExpr.Eval(doc)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Value)
		})
	}
}

// 测试大于等于比较
func TestGteExpr(t *testing.T) {
	doc := createTestDoc()

	tests := []struct {
		name     string
		path     string
		value    any
		expected bool
	}{
		{"整数大于等于（大于）", "root/intField", int64(41), true},
		{"整数大于等于（等于）", "root/intField", int64(42), true},
		{"整数不大于等于", "root/intField", int64(43), false},
		{"浮点数大于等于（大于）", "root/doubleField", 3.13, true},
		{"浮点数大于等于（等于）", "root/doubleField", 3.14, true},
		{"浮点数不大于等于", "root/doubleField", 3.15, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fieldExpr := &qfe.FieldValueExpr{Path: tt.path}
			valueExpr := &qfe.ValueExpr{Value: tt.value}
			gteExpr := &qfe.GteExpr{O1: fieldExpr, O2: valueExpr}

			result, err := gteExpr.Eval(doc)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Value)
		})
	}
}

// 测试小于比较
func TestLtExpr(t *testing.T) {
	doc := createTestDoc()

	tests := []struct {
		name     string
		path     string
		value    any
		expected bool
	}{
		{"整数小于", "root/intField", int64(43), true},
		{"整数不小于", "root/intField", int64(42), false},
		{"浮点数小于", "root/doubleField", 3.15, true},
		{"浮点数不小于", "root/doubleField", 3.14, false},
		{"字符串小于", "root/stringField", "hellz", true},
		{"字符串不小于", "root/stringField", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fieldExpr := &qfe.FieldValueExpr{Path: tt.path}
			valueExpr := &qfe.ValueExpr{Value: tt.value}
			ltExpr := &qfe.LtExpr{O1: fieldExpr, O2: valueExpr}

			result, err := ltExpr.Eval(doc)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Value)
		})
	}
}

// 测试小于等于比较
func TestLteExpr(t *testing.T) {
	doc := createTestDoc()

	tests := []struct {
		name     string
		path     string
		value    any
		expected bool
	}{
		{"整数小于等于（小于）", "root/intField", int64(43), true},
		{"整数小于等于（等于）", "root/intField", int64(42), true},
		{"整数不小于等于", "root/intField", int64(41), false},
		{"浮点数小于等于（小于）", "root/doubleField", 3.15, true},
		{"浮点数小于等于（等于）", "root/doubleField", 3.14, true},
		{"浮点数不小于等于", "root/doubleField", 3.13, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fieldExpr := &qfe.FieldValueExpr{Path: tt.path}
			valueExpr := &qfe.ValueExpr{Value: tt.value}
			lteExpr := &qfe.LteExpr{O1: fieldExpr, O2: valueExpr}

			result, err := lteExpr.Eval(doc)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Value)
		})
	}
}

// 测试 In 表达式
func TestInExpr(t *testing.T) {
	doc := createTestDoc()

	tests := []struct {
		name     string
		path     string
		values   []any
		expected bool
	}{
		{
			"整数在列表中",
			"root/intField",
			[]any{int64(41), int64(42), int64(43)},
			true,
		},
		{
			"整数不在列表中",
			"root/intField",
			[]any{int64(41), int64(43), int64(44)},
			false,
		},
		{
			"字符串在列表中",
			"root/stringField",
			[]any{"hello", "world"},
			true,
		},
		{
			"字符串不在列表中",
			"root/stringField",
			[]any{"world", "golang"},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fieldExpr := &qfe.FieldValueExpr{Path: tt.path}
			valueExprs := make([]qfe.QueryFilterExpr, len(tt.values))
			for i, v := range tt.values {
				valueExprs[i] = &qfe.ValueExpr{Value: v}
			}
			inExpr := &qfe.InExpr{O1: fieldExpr, O2: valueExprs}

			result, err := inExpr.Eval(doc)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Value)
		})
	}
}

// 测试 Not In 表达式
func TestNinExpr(t *testing.T) {
	doc := createTestDoc()

	tests := []struct {
		name     string
		path     string
		values   []any
		expected bool
	}{
		{
			"整数不在列表中",
			"root/intField",
			[]any{int64(41), int64(43), int64(44)},
			true,
		},
		{
			"整数在列表中",
			"root/intField",
			[]any{int64(41), int64(42), int64(43)},
			false,
		},
		{
			"字符串不在列表中",
			"root/stringField",
			[]any{"world", "golang"},
			true,
		},
		{
			"字符串在列表中",
			"root/stringField",
			[]any{"hello", "world"},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fieldExpr := &qfe.FieldValueExpr{Path: tt.path}
			valueExprs := make([]qfe.QueryFilterExpr, len(tt.values))
			for i, v := range tt.values {
				valueExprs[i] = &qfe.ValueExpr{Value: v}
			}
			ninExpr := &qfe.NinExpr{O1: fieldExpr, O2: valueExprs}

			result, err := ninExpr.Eval(doc)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Value)
		})
	}
}

// 测试错误情况
func TestErrorCases(t *testing.T) {
	doc := createTestDoc()

	tests := []struct {
		name        string
		expr        qfe.QueryFilterExpr
		expectedErr string
	}{
		{
			"字段不存在",
			&qfe.EqExpr{
				O1: &qfe.FieldValueExpr{Path: "root/nonexistentField"},
				O2: &qfe.ValueExpr{Value: 42},
			},
			"field error: path=root/nonexistentField",
		},
		{
			"类型不匹配",
			&qfe.EqExpr{
				O1: &qfe.FieldValueExpr{Path: "root/intField"},
				O2: &qfe.ValueExpr{Value: "42"}, // 字符串和整数比较
			},
			"type error: comparing numeric type with string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.expr.Eval(doc)
			assert.Error(t, err)
			assert.ErrorContains(t, err, tt.expectedErr)
		})
	}
}

// 测试复杂的嵌套比较
func TestComplexComparisons(t *testing.T) {
	doc := createTestDoc()

	// 测试嵌套字段访问
	t.Run("嵌套字段访问", func(t *testing.T) {
		expr := &qfe.EqExpr{
			O1: &qfe.FieldValueExpr{Path: "root/mapField/nestedKey"},
			O2: &qfe.ValueExpr{Value: "nestedValue"},
		}
		result, err := expr.Eval(doc)
		assert.NoError(t, err)
		assert.Equal(t, true, result.Value)
	})

	// 测试数组元素访问
	t.Run("数组元素访问", func(t *testing.T) {
		expr := &qfe.EqExpr{
			O1: &qfe.FieldValueExpr{Path: "root/listField/0"},
			O2: &qfe.ValueExpr{Value: int64(1)},
		}
		result, err := expr.Eval(doc)
		assert.NoError(t, err)
		assert.Equal(t, true, result.Value)
	})
}

// 测试 compareValues 函数
func TestCompareValues(t *testing.T) {
	tests := []struct {
		name        string
		v1          any
		v2          any
		expected    int
		expectError bool
		errorMsg    string
	}{
		// nil 值比较
		{
			name:        "nil 和 nil 比较",
			v1:          nil,
			v2:          nil,
			expected:    0,
			expectError: false,
		},
		{
			name:        "nil 和非 nil 比较",
			v1:          nil,
			v2:          42,
			expected:    -1,
			expectError: false,
		},
		{
			name:        "非 nil 和 nil 比较",
			v1:          42,
			v2:          nil,
			expected:    1,
			expectError: false,
		},

		// 布尔值比较
		{
			name:        "相同布尔值比较",
			v1:          true,
			v2:          true,
			expected:    0,
			expectError: false,
		},
		{
			name:        "不同布尔值比较 (false < true)",
			v1:          false,
			v2:          true,
			expected:    -1,
			expectError: false,
		},
		{
			name:        "不同布尔值比较 (true > false)",
			v1:          true,
			v2:          false,
			expected:    1,
			expectError: false,
		},
		{
			name:        "布尔值和其他类型比较",
			v1:          true,
			v2:          42,
			expectError: true,
			errorMsg:    "type error: comparing bool with int",
		},

		// 整数比较
		{
			name:        "相同整数比较",
			v1:          int64(42),
			v2:          int64(42),
			expected:    0,
			expectError: false,
		},
		{
			name:        "不同整数比较 (小于)",
			v1:          int32(41),
			v2:          int64(42),
			expected:    -1,
			expectError: false,
		},
		{
			name:        "不同整数比较 (大于)",
			v1:          uint64(43),
			v2:          int8(42),
			expected:    1,
			expectError: false,
		},
		{
			name:        "整数和其他类型比较",
			v1:          int64(42),
			v2:          "42",
			expectError: true,
			errorMsg:    "type error: comparing numeric type with string",
		},

		// 浮点数比较
		{
			name:        "相同浮点数比较",
			v1:          float64(3.14),
			v2:          float64(3.14),
			expected:    0,
			expectError: false,
		},
		{
			name:        "不同浮点数比较 (小于)",
			v1:          float32(3.13),
			v2:          float64(3.14),
			expected:    -1,
			expectError: false,
		},
		{
			name:        "不同浮点数比较 (大于)",
			v1:          float64(3.15),
			v2:          float32(3.14),
			expected:    1,
			expectError: false,
		},
		{
			name:        "浮点数和其他类型比较",
			v1:          float64(3.14),
			v2:          "3.14",
			expectError: true,
			errorMsg:    "type error: comparing float type with string",
		},

		// 字符串比较
		{
			name:        "相同字符串比较",
			v1:          "hello",
			v2:          "hello",
			expected:    0,
			expectError: false,
		},
		{
			name:        "不同字符串比较 (小于)",
			v1:          "hello",
			v2:          "world",
			expected:    -1,
			expectError: false,
		},
		{
			name:        "不同字符串比较 (大于)",
			v1:          "world",
			v2:          "hello",
			expected:    1,
			expectError: false,
		},
		{
			name:        "字符串和其他类型比较",
			v1:          "42",
			v2:          42,
			expectError: true,
			errorMsg:    "type error: comparing string with int",
		},

		// 切片比较
		{
			name:        "相同切片比较",
			v1:          []any{1, 2, 3},
			v2:          []any{1, 2, 3},
			expected:    0,
			expectError: false,
		},
		{
			name:        "不同长度切片比较",
			v1:          []any{1, 2},
			v2:          []any{1, 2, 3},
			expected:    -1,
			expectError: false,
		},
		{
			name:        "不同内容切片比较",
			v1:          []any{1, 3, 2},
			v2:          []any{1, 2, 3},
			expected:    1,
			expectError: false,
		},
		{
			name:        "切片和其他类型比较",
			v1:          []any{1, 2, 3},
			v2:          42,
			expectError: true,
			errorMsg:    "type error: comparing array with int",
		},

		// map 比较
		{
			name:        "相同 map 比较",
			v1:          map[string]any{"a": 1, "b": 2},
			v2:          map[string]any{"a": 1, "b": 2},
			expected:    0,
			expectError: false,
		},
		{
			name:        "不同大小 map 比较",
			v1:          map[string]any{"a": 1},
			v2:          map[string]any{"a": 1, "b": 2},
			expected:    -1,
			expectError: false,
		},
		{
			name:        "不同内容 map 比较",
			v1:          map[string]any{"a": 2, "b": 2},
			v2:          map[string]any{"a": 1, "b": 2},
			expected:    1,
			expectError: false,
		},
		{
			name:        "map 和其他类型比较",
			v1:          map[string]any{"a": 1},
			v2:          42,
			expectError: true,
			errorMsg:    "type error: comparing map with int",
		},

		// 不支持的类型
		{
			name:        "不支持的类型",
			v1:          make(chan int),
			v2:          make(chan int),
			expectError: true,
			errorMsg:    "type error: unsupported type chan int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := qfe.CompareValues(tt.v1, tt.v2)

			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// 测试逻辑与表达式
func TestAndExpr(t *testing.T) {
	doc := createTestDoc()

	tests := []struct {
		name     string
		exprs    []qfe.QueryFilterExpr
		expected bool
	}{
		{
			"全部为真",
			[]qfe.QueryFilterExpr{
				&qfe.EqExpr{
					O1: &qfe.FieldValueExpr{Path: "root/intField"},
					O2: &qfe.ValueExpr{Value: int64(42)},
				},
				&qfe.EqExpr{
					O1: &qfe.FieldValueExpr{Path: "root/stringField"},
					O2: &qfe.ValueExpr{Value: "hello"},
				},
			},
			true,
		},
		{
			"部分为假",
			[]qfe.QueryFilterExpr{
				&qfe.EqExpr{
					O1: &qfe.FieldValueExpr{Path: "root/intField"},
					O2: &qfe.ValueExpr{Value: int64(42)},
				},
				&qfe.EqExpr{
					O1: &qfe.FieldValueExpr{Path: "root/stringField"},
					O2: &qfe.ValueExpr{Value: "world"},
				},
			},
			false,
		},
		{
			"空表达式列表",
			[]qfe.QueryFilterExpr{},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			andExpr := &qfe.AndExpr{Exprs: tt.exprs}
			result, err := andExpr.Eval(doc)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Value)
		})
	}
}

// 测试逻辑或表达式
func TestOrExpr(t *testing.T) {
	doc := createTestDoc()

	tests := []struct {
		name     string
		exprs    []qfe.QueryFilterExpr
		expected bool
	}{
		{
			"全部为真",
			[]qfe.QueryFilterExpr{
				&qfe.EqExpr{
					O1: &qfe.FieldValueExpr{Path: "root/intField"},
					O2: &qfe.ValueExpr{Value: int64(42)},
				},
				&qfe.EqExpr{
					O1: &qfe.FieldValueExpr{Path: "root/stringField"},
					O2: &qfe.ValueExpr{Value: "hello"},
				},
			},
			true,
		},
		{
			"部分为真",
			[]qfe.QueryFilterExpr{
				&qfe.EqExpr{
					O1: &qfe.FieldValueExpr{Path: "root/intField"},
					O2: &qfe.ValueExpr{Value: int64(43)},
				},
				&qfe.EqExpr{
					O1: &qfe.FieldValueExpr{Path: "root/stringField"},
					O2: &qfe.ValueExpr{Value: "hello"},
				},
			},
			true,
		},
		{
			"全部为假",
			[]qfe.QueryFilterExpr{
				&qfe.EqExpr{
					O1: &qfe.FieldValueExpr{Path: "root/intField"},
					O2: &qfe.ValueExpr{Value: int64(43)},
				},
				&qfe.EqExpr{
					O1: &qfe.FieldValueExpr{Path: "root/stringField"},
					O2: &qfe.ValueExpr{Value: "world"},
				},
			},
			false,
		},
		{
			"空表达式列表",
			[]qfe.QueryFilterExpr{},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orExpr := &qfe.OrExpr{Exprs: tt.exprs}
			result, err := orExpr.Eval(doc)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Value)
		})
	}
}

// 测试逻辑非表达式
func TestNotExpr(t *testing.T) {
	doc := createTestDoc()

	tests := []struct {
		name     string
		expr     qfe.QueryFilterExpr
		expected bool
	}{
		{
			"对真取反",
			&qfe.EqExpr{
				O1: &qfe.FieldValueExpr{Path: "root/intField"},
				O2: &qfe.ValueExpr{Value: int64(42)},
			},
			false,
		},
		{
			"对假取反",
			&qfe.EqExpr{
				O1: &qfe.FieldValueExpr{Path: "root/intField"},
				O2: &qfe.ValueExpr{Value: int64(43)},
			},
			true,
		},
		{
			"嵌套取反",
			&qfe.NotExpr{
				Expr: &qfe.NotExpr{
					Expr: &qfe.EqExpr{
						O1: &qfe.FieldValueExpr{Path: "root/intField"},
						O2: &qfe.ValueExpr{Value: int64(42)},
					},
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notExpr := &qfe.NotExpr{Expr: tt.expr}
			result, err := notExpr.Eval(doc)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Value)
		})
	}
}

// 测试正则表达式匹配
func TestRegexExpr(t *testing.T) {
	doc := createTestDoc()

	tests := []struct {
		name        string
		path        string
		pattern     string
		expected    bool
		expectError bool
		errorMsg    string
	}{
		{
			"简单匹配",
			"root/stringField",
			"^hello$",
			true,
			false,
			"",
		},
		{
			"部分匹配",
			"root/stringField",
			"ell",
			true,
			false,
			"",
		},
		{
			"不匹配",
			"root/stringField",
			"^world$",
			false,
			false,
			"",
		},
		{
			"无效的正则表达式",
			"root/stringField",
			"[",
			false,
			true,
			"syntax error: invalid regex pattern",
		},
		{
			"非字符串字段",
			"root/intField",
			"^42$",
			false,
			true,
			"type error: expected string for regex matching",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regexExpr := &qfe.RegexExpr{
				O1:    &qfe.FieldValueExpr{Path: tt.path},
				Regex: tt.pattern,
			}
			result, err := regexExpr.Eval(doc)
			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result.Value)
			}
		})
	}
}

// 测试表达式的 JSON 序列化和反序列化
func TestExprSerialization(t *testing.T) {
	tests := []struct {
		name string
		expr qfe.QueryFilterExpr
	}{
		{
			"值表达式",
			&qfe.ValueExpr{Value: int64(42)},
		},
		{
			"字段值表达式",
			&qfe.FieldValueExpr{Path: "root/intField"},
		},
		{
			"相等比较表达式",
			&qfe.EqExpr{
				O1: &qfe.FieldValueExpr{Path: "root/intField"},
				O2: &qfe.ValueExpr{Value: int64(42)},
			},
		},
		{
			"不等比较表达式",
			&qfe.NeExpr{
				O1: &qfe.FieldValueExpr{Path: "root/intField"},
				O2: &qfe.ValueExpr{Value: int64(42)},
			},
		},
		{
			"大于比较表达式",
			&qfe.GtExpr{
				O1: &qfe.FieldValueExpr{Path: "root/intField"},
				O2: &qfe.ValueExpr{Value: int64(42)},
			},
		},
		{
			"包含比较表达式",
			&qfe.InExpr{
				O1: &qfe.FieldValueExpr{Path: "root/intField"},
				O2: []qfe.QueryFilterExpr{
					&qfe.ValueExpr{Value: int64(41)},
					&qfe.ValueExpr{Value: int64(42)},
					&qfe.ValueExpr{Value: int64(43)},
				},
			},
		},
		{
			"逻辑与表达式",
			&qfe.AndExpr{
				Exprs: []qfe.QueryFilterExpr{
					&qfe.EqExpr{
						O1: &qfe.FieldValueExpr{Path: "root/intField"},
						O2: &qfe.ValueExpr{Value: int64(42)},
					},
					&qfe.EqExpr{
						O1: &qfe.FieldValueExpr{Path: "root/stringField"},
						O2: &qfe.ValueExpr{Value: "hello"},
					},
				},
			},
		},
		{
			"逻辑非表达式",
			&qfe.NotExpr{
				Expr: &qfe.EqExpr{
					O1: &qfe.FieldValueExpr{Path: "root/intField"},
					O2: &qfe.ValueExpr{Value: int64(42)},
				},
			},
		},
		{
			"正则表达式",
			&qfe.RegexExpr{
				O1:    &qfe.FieldValueExpr{Path: "root/stringField"},
				Regex: "^hello$",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 序列化
			data, err := tt.expr.MarshalJSON()
			assert.NoError(t, err)

			// 反序列化
			expr2, err := qfe.UnmarshalQueryFilterExpr(data)
			assert.NoError(t, err)

			// 再次序列化，确保结果一致
			data2, err := expr2.MarshalJSON()
			assert.NoError(t, err)
			assert.JSONEq(t, string(data), string(data2))

			// 使用测试文档评估两个表达式，确保行为一致
			doc := createTestDoc()
			result1, err1 := tt.expr.Eval(doc)
			if err1 != nil {
				_, err2 := expr2.Eval(doc)
				assert.Error(t, err2)
				assert.Equal(t, err1.Error(), err2.Error())
			} else {
				result2, err2 := expr2.Eval(doc)
				assert.NoError(t, err2)
				if result2 != nil {
					assert.Equal(t, result1.Value, result2.Value)
				}
			}
		})
	}
}

// 测试无效的 JSON 反序列化
func TestInvalidExprDeserialization(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		expectedErr string
	}{
		{
			"无效的 JSON",
			`{"type": "eq", "o1": {`,
			"unexpected end of JSON input",
		},
		{
			"未知的表达式类型",
			`{"type": "unknown"}`,
			"unsupported expression type: unknown",
		},
		{
			"缺少必要字段",
			`{"type": "eq"}`,
			"missing operands for EQ expression",
		},
		{
			"无效的正则表达式",
			`{"type": "regex", "o1": {"type": "field", "path": "root/stringField"}, "regex": "["}`,
			"invalid regex pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := qfe.UnmarshalQueryFilterExpr([]byte(tt.json))
			assert.Error(t, err)
			assert.ErrorContains(t, err, tt.expectedErr)
		})
	}
}

// 测试 Exists 表达式
func TestExistsExpr(t *testing.T) {
	doc := createTestDoc()

	tests := []struct {
		name        string
		path        string
		expected    bool
		expectError bool
		errorMsg    string
	}{
		{
			"存在的字段",
			"root/intField",
			true,
			false,
			"",
		},
		{
			"存在的嵌套字段",
			"root/mapField/nestedKey",
			true,
			false,
			"",
		},
		{
			"不存在的字段",
			"root/nonexistentField",
			false,
			false,
			"",
		},
		{
			"不存在的嵌套字段",
			"root/mapField/nonexistentKey",
			false,
			false,
			"",
		},
		{
			"空路径",
			"",
			false,
			true,
			"field error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existsExpr := &qfe.ExistsExpr{
				Field: &qfe.FieldValueExpr{Path: tt.path},
			}
			result, err := existsExpr.Eval(doc)
			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result.Value)
			}
		})
	}
}
