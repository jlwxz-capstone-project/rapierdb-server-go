package main

import (
	"reflect"
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/doc_visitor"
)

func TestExtractSegments(t *testing.T) {
	// 普通点号分隔
	t.Run("解析点号表示法", func(t *testing.T) {
		cases := []struct {
			path     string
			expected []any
		}{
			{"foo.bar", []any{"foo", "bar"}},
			{"foo.bar.baz", []any{"foo", "bar", "baz"}},
			{"foo.0", []any{"foo", "0"}},
		}

		for _, c := range cases {
			result, err := doc_visitor.ExtractSegments(c.path)
			if err != nil {
				t.Fatalf("解析路径 %s 失败: %v", c.path, err)
			}
			if !reflect.DeepEqual(result, c.expected) {
				t.Errorf("路径 %s: 期望 %v, 得到 %v", c.path, c.expected, result)
			}
		}
	})

	// 数字索引（数组下标）
	t.Run("解析数字方括号为数字", func(t *testing.T) {
		cases := []struct {
			path     string
			expected []any
		}{
			{"[0]", []any{0}},
			{"[12].foo", []any{12, "foo"}},
			{"foo[42].bar", []any{"foo", 42, "bar"}},
		}

		for _, c := range cases {
			result, err := doc_visitor.ExtractSegments(c.path)
			if err != nil {
				t.Fatalf("解析路径 %s 失败: %v", c.path, err)
			}
			if !reflect.DeepEqual(result, c.expected) {
				t.Errorf("路径 %s: 期望 %v, 得到 %v", c.path, c.expected, result)
			}
		}
	})

	// 字符串属性（带引号）
	t.Run("解析带引号的方括号表示法", func(t *testing.T) {
		cases := []struct {
			path     string
			expected []any
		}{
			{`obj["key"]`, []any{"obj", "key"}},
			{`obj['key']`, []any{"obj", "key"}},
			{`foo["bar"].baz`, []any{"foo", "bar", "baz"}},
			{`foo['bar.baz']`, []any{"foo", "bar.baz"}},
			{`foo["bar.baz"].qux`, []any{"foo", "bar.baz", "qux"}},
		}

		for _, c := range cases {
			result, err := doc_visitor.ExtractSegments(c.path)
			if err != nil {
				t.Fatalf("解析路径 %s 失败: %v", c.path, err)
			}
			if !reflect.DeepEqual(result, c.expected) {
				t.Errorf("路径 %s: 期望 %v, 得到 %v", c.path, c.expected, result)
			}
		}
	})

	// 带转义的字符串属性
	t.Run("解析带转义的引号方括号表示法", func(t *testing.T) {
		cases := []struct {
			path     string
			expected []any
		}{
			{`obj["b\"az"]`, []any{"obj", `b"az`}},
			{`obj['b\'az']`, []any{"obj", `b'az`}},
			{`foo["b\\az"]`, []any{"foo", `b\az`}},
		}

		for _, c := range cases {
			result, err := doc_visitor.ExtractSegments(c.path)
			if err != nil {
				t.Fatalf("解析路径 %s 失败: %v", c.path, err)
			}
			if !reflect.DeepEqual(result, c.expected) {
				t.Errorf("路径 %s: 期望 %v, 得到 %v", c.path, c.expected, result)
			}
		}
	})

	// 属性名后跟数组下标
	t.Run("解析带数组索引的属性", func(t *testing.T) {
		cases := []struct {
			path     string
			expected []any
		}{
			{"foo[1]", []any{"foo", 1}},
			{"foo.bar[2]", []any{"foo", "bar", 2}},
		}

		for _, c := range cases {
			result, err := doc_visitor.ExtractSegments(c.path)
			if err != nil {
				t.Fatalf("解析路径 %s 失败: %v", c.path, err)
			}
			if !reflect.DeepEqual(result, c.expected) {
				t.Errorf("路径 %s: 期望 %v, 得到 %v", c.path, c.expected, result)
			}
		}
	})

	// 只有一个 segment
	t.Run("解析单个段", func(t *testing.T) {
		cases := []struct {
			path     string
			expected []any
		}{
			{"foo", []any{"foo"}},
		}

		for _, c := range cases {
			result, err := doc_visitor.ExtractSegments(c.path)
			if err != nil {
				t.Fatalf("解析路径 %s 失败: %v", c.path, err)
			}
			if !reflect.DeepEqual(result, c.expected) {
				t.Errorf("路径 %s: 期望 %v, 得到 %v", c.path, c.expected, result)
			}
		}
	})

	// 非法路径
	t.Run("非法路径应抛出错误", func(t *testing.T) {
		invalidPaths := []string{
			"foo..bar",
			"foo[abc",
			`foo["bar]`,
			"foo[1a]",
		}

		for _, path := range invalidPaths {
			_, err := doc_visitor.ExtractSegments(path)
			if err == nil {
				t.Errorf("路径 %s 应该失败但成功了", path)
			}
		}
	})
}
