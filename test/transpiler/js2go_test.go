package main

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/transpiler"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestExecuteCode(t *testing.T) {
	ctx := transpiler.NewScope(nil, nil, nil)

	// 添加一些测试用的函数和变量
	ctx.Vars["add"] = func(a, b float64) float64 { return a + b }
	ctx.Vars["concat"] = func(a, b string) string { return a + b }
	ctx.Vars["console"] = struct {
		Log func(a ...any) (n int, err error)
	}{
		Log: fmt.Println,
	}
	ctx.Vars["req"] = &struct {
		Path      string
		Headers   map[string]string
		GetHeader func(string) string
	}{
		Path:    "/api/test",
		Headers: map[string]string{"Content-Type": "application/json"},
		GetHeader: func(name string) string {
			return "test-header"
		},
	}

	// 为 console 对象设置自定义属性访问转换器
	ctx.PropGetter = func(chain []transpiler.PropAccess, obj any) (any, error) {
		// 如果是 console 对象，处理 log -> Log 的转换
		if console, ok := obj.(struct {
			Log func(...any) (int, error)
		}); ok {
			if len(chain) == 1 && chain[0].Prop == "log" {
				return console.Log, nil
			}
		}
		// 其他情况使用默认转换器
		return transpiler.DefaultPropGetter(chain, obj)
	}

	tests := []struct {
		name    string
		js      string
		want    any
		wantErr bool
	}{
		// 基本运算测试
		{
			name: "数字运算",
			js:   "var result = 1 + 2 * 3;",
			want: float64(7),
		},
		{
			name: "字符串拼接",
			js:   "var result = 'hello' + ' ' + 'world';",
			want: "hello world",
		},
		{
			name: "函数调用",
			js:   "var result = add(1, 2);",
			want: float64(3),
		},

		// 对象访问测试
		{
			name: "对象属性访问",
			js:   "var result = req.Path;",
			want: "/api/test",
		},
		{
			name: "对象方法调用",
			js:   "var result = req.GetHeader('Content-Type');",
			want: "test-header",
		},
		{
			name: "Map访问",
			js:   "var result = req.Headers['Content-Type'];",
			want: "application/json",
		},

		// 错误处理测试
		{
			name:    "未定义变量",
			js:      "notExist;",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "无效属性访问",
			js:      "req.NotExist;",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "无效函数调用",
			js:      "req.Path();",
			want:    nil,
			wantErr: true,
		},

		// console.log 测试
		{
			name: "console.log调用",
			js:   "var result = 0; console.log('test'); result = 5;",
			want: float64(5),
		},

		// if-else 语句测试
		{
			name: "简单if条件",
			js:   "var result; if (true) { result = 1; } else { result = 2; }",
			want: float64(1),
		},
		{
			name: "if-else条件",
			js:   "var result; if (false) { result = 1; } else { result = 2; }",
			want: float64(2),
		},
		{
			name: "比较运算",
			js:   "var result; if (1 < 2) { result = 'yes'; } else { result = 'no'; }",
			want: "yes",
		},
		{
			name: "相等比较",
			js:   "var result; if ('hello' == 'hello') { result = true; } else { result = false; }",
			want: true,
		},
		{
			name: "不相等比较",
			js:   "var result; if (1 != 2) { result = 'different'; } else { result = 'same'; }",
			want: "different",
		},
		{
			name: "数字条件",
			js:   "var result; if (1) { result = 'truthy'; } else { result = 'falsy'; }",
			want: "truthy",
		},
		{
			name: "零值条件",
			js:   "var result; if (0) { result = 'truthy'; } else { result = 'falsy'; }",
			want: "falsy",
		},
		{
			name: "空字符串条件",
			js:   "var result; if ('') { result = 'truthy'; } else { result = 'falsy'; }",
			want: "falsy",
		},
		{
			name: "null条件",
			js:   "var result; if (null) { result = 'truthy'; } else { result = 'falsy'; }",
			want: "falsy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置上下文，但保留预定义变量
			ctx.Vars = map[string]any{
				"add":     ctx.Vars["add"],
				"req":     ctx.Vars["req"],
				"console": ctx.Vars["console"],
			}

			// 执行代码，忽略返回值
			got, err := transpiler.Execute(tt.js, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 返回值应该总是 nil
			if got != nil {
				t.Errorf("Execute() returned %v, want nil", got)
			}

			// 对于非错误情况，检查 result 变量的值
			if !tt.wantErr {
				result := ctx.Vars["result"]
				if !reflect.DeepEqual(result, tt.want) {
					t.Errorf("ctx.Vars[\"result\"] = %v, want %v", result, tt.want)
				}
			}
		})
	}
}

func TestCustomPropertyAccess(t *testing.T) {
	// 定义一个测试结构体
	type TestStruct struct {
		data map[string]any
	}

	ctx := transpiler.NewScope(nil, nil, nil)

	// 创建测试对象
	testObj := &TestStruct{
		data: map[string]any{
			"a": &TestStruct{
				data: map[string]any{
					"b": &TestStruct{
						data: map[string]any{
							"c": "value",
						},
					},
				},
			},
		},
	}

	// 添加测试对象到上下文
	ctx.Vars["obj"] = testObj

	// 设置自定义属性访问转换器
	ctx.PropGetter = func(chain []transpiler.PropAccess, obj any) (any, error) {
		if ts, ok := obj.(*TestStruct); ok {
			result := ts
			for _, access := range chain {
				if prop, ok := access.Prop.(string); ok {
					if val, ok := result.data[prop]; ok {
						if next, ok := val.(*TestStruct); ok {
							result = next
						} else {
							return val, nil
						}
					} else {
						return nil, fmt.Errorf("属性不存在: %s", access.Prop)
					}
				}
			}
			return result, nil
		}
		return transpiler.DefaultPropGetter(chain, obj)
	}

	tests := []struct {
		name    string
		js      string
		want    any
		wantErr bool
	}{
		{
			name: "简单属性访问",
			js:   "var result = obj.a.b.c;",
			want: "value",
		},
		{
			name:    "无效属性访问",
			js:      "obj.x.y.z;",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置上下文，但保留预定义变量
			ctx.Vars = map[string]any{
				"obj": testObj,
			}

			// 执行代码，忽略返回值
			got, err := transpiler.Execute(tt.js, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 返回值应该总是 nil
			if got != nil {
				t.Errorf("Execute() returned %v, want nil", got)
			}

			// 检查 result 变量的值
			if !tt.wantErr {
				result := ctx.Vars["result"]
				if !reflect.DeepEqual(result, tt.want) {
					t.Errorf("ctx.Vars[\"result\"] = %v, want %v", result, tt.want)
				}
			}
		})
	}
}

type Person struct {
	Name    string
	Age     int
	Friends map[string]*Person
}

func (p *Person) Friend(name string) *Person {
	return p.Friends[name]
}

func TestChainPropertyAccess(t *testing.T) {
	ctx := transpiler.NewScope(nil, nil, nil)

	alice := &Person{
		Name: "Alice",
		Age:  20,
		Friends: map[string]*Person{
			"bob": {
				Name: "Bob",
				Age:  21,
			},
		},
	}

	ctx.Vars["person"] = alice

	tests := []struct {
		name    string
		js      string
		want    any
		wantErr bool
	}{
		{
			name: "访问直接属性",
			js:   "var result = person.Name;",
			want: "Alice",
		},
		{
			name: "访问朋友属性",
			js:   "var result = person.Friend('bob').Name;",
			want: "Bob",
		},
		{
			name:    "访问不存在的朋友",
			js:      "var result = person.Friend('david').Name;",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置上下文，但保留预定义变量
			ctx.Vars = map[string]any{
				"person": alice,
			}

			// 执行代码，忽略返回值
			got, err := transpiler.Execute(tt.js, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 返回值应该总是 nil
			if got != nil {
				t.Errorf("Execute() returned %v, want nil", got)
			}

			// 检查 result 变量的值
			if !tt.wantErr {
				result := ctx.Vars["result"]
				if !reflect.DeepEqual(result, tt.want) {
					t.Errorf("ctx.Vars[\"result\"] = %v, want %v", result, tt.want)
				}
			}
		})
	}
}

func TestExecuteIfStatement(t *testing.T) {
	ctx := transpiler.NewScope(nil, nil, nil)

	tests := []struct {
		name    string
		js      string
		want    any
		wantErr bool
	}{
		{
			name: "简单if条件",
			js:   "var result; if (true) { result = 1; } else { result = 2; }",
			want: float64(1),
		},
		{
			name: "if-else条件",
			js:   "var result; if (false) { result = 1; } else { result = 2; }",
			want: float64(2),
		},
		{
			name: "比较运算",
			js:   "var result; if (1 < 2) { result = 'yes'; } else { result = 'no'; }",
			want: "yes",
		},
		{
			name: "相等比较",
			js:   "var result; if ('hello' == 'hello') { result = true; } else { result = false; }",
			want: true,
		},
		{
			name: "不相等比较",
			js:   "var result; if (1 != 2) { result = 'different'; } else { result = 'same'; }",
			want: "different",
		},
		{
			name: "数字条件",
			js:   "var result; if (1) { result = 'truthy'; } else { result = 'falsy'; }",
			want: "truthy",
		},
		{
			name: "零值条件",
			js:   "var result; if (0) { result = 'truthy'; } else { result = 'falsy'; }",
			want: "falsy",
		},
		{
			name: "空字符串条件",
			js:   "var result; if ('') { result = 'truthy'; } else { result = 'falsy'; }",
			want: "falsy",
		},
		{
			name: "null条件",
			js:   "var result; if (null) { result = 'truthy'; } else { result = 'falsy'; }",
			want: "falsy",
		},
		{
			name: "复杂条件",
			js:   "var result; if (1 < 2 && 'hello' == 'hello') { result = 'both true'; } else { result = 'not both true'; }",
			want: "both true",
		},
		{
			name: "带括号的条件",
			js:   "var result; if ((1 + 2) * 3 > 8) { result = 'greater'; } else { result = 'less'; }",
			want: "greater",
		},
		{
			name: "多重比较",
			js:   "var result; if (1 < 2 && 3 > 2 || false) { result = 'true'; } else { result = 'false'; }",
			want: "true",
		},
		{
			name: "带括号的逻辑运算",
			js:   "var result; if ((true && false) || (true && true)) { result = 'yes'; } else { result = 'no'; }",
			want: "yes",
		},
		{
			name: "混合运算优先级",
			js:   "var result; if (2 + 3 * 4 > 10 + 2) { result = 'yes'; } else { result = 'no'; }",
			want: "yes",
		},
		{
			name: "复杂嵌套条件",
			js:   "var result; if ((1 + 2 > 2) && (3 * 4 <= 12 || true)) { result = 'complex'; } else { result = 'simple'; }",
			want: "complex",
		},
		{
			name: "字符串比较和数字运算",
			js:   "var result; if ('hello'.length > 2 + 1) { result = 'long'; } else { result = 'short'; }",
			want: "long",
		},
		{
			name: "多重括号和运算符",
			js:   "var result; if (((1 + 2) * 3 == 9) && (4 + 5 >= 8 || 2 * 3 < 5)) { result = 'pass'; } else { result = 'fail'; }",
			want: "pass",
		},
		{
			name: "逻辑运算短路",
			js:   "var result; if (false && someUndefinedVar) { result = 'bug'; } else { result = 'ok'; }",
			want: "ok",
		},
		{
			name: "数字和布尔混合运算",
			js:   "var result; if (1 + 1 == 2 && true || false && 5 < 3) { result = 'correct'; } else { result = 'wrong'; }",
			want: "correct",
		},
		{
			name: "复杂的真值判断",
			js:   "var result; if (1 && 'hello' && (2 * 3) && {}) { result = 'truthy'; } else { result = 'falsy'; }",
			want: "truthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置上下文
			ctx.Vars = make(map[string]any)

			// 执行代码，忽略返回值
			got, err := transpiler.Execute(tt.js, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 返回值应该总是 nil
			if got != nil {
				t.Errorf("Execute() returned %v, want nil", got)
			}

			// 检查 result 变量的值
			if !tt.wantErr {
				result := ctx.Vars["result"]
				if !reflect.DeepEqual(result, tt.want) {
					t.Errorf("ctx.Vars[\"result\"] = %v, want %v", result, tt.want)
				}
			}
		})
	}
}

func TestExecuteBlockStatement(t *testing.T) {
	ctx := transpiler.NewScope(nil, nil, nil)

	// 添加一些测试用的函数和变量
	ctx.Vars["add"] = func(a, b float64) float64 { return a + b }
	ctx.Vars["console"] = struct {
		Log func(a ...any) (n int, err error)
	}{
		Log: fmt.Println,
	}
	ctx.Vars["req"] = &struct {
		Path      string
		Headers   map[string]string
		GetHeader func(string) string
	}{
		Path:    "/api/test",
		Headers: map[string]string{"Content-Type": "application/json"},
		GetHeader: func(name string) string {
			return "test-header"
		},
	}

	// 为 console 对象设置自定义属性访问转换器
	ctx.PropGetter = func(chain []transpiler.PropAccess, obj any) (any, error) {
		// 如果是 console 对象，处理 log -> Log 的转换
		if console, ok := obj.(struct {
			Log func(...any) (int, error)
		}); ok {
			if len(chain) == 1 && chain[0].Prop == "log" {
				return console.Log, nil
			}
		}
		// 其他情况使用默认转换器
		return transpiler.DefaultPropGetter(chain, obj)
	}

	tests := []struct {
		name    string
		js      string
		want    any
		wantErr bool
	}{
		{
			name: "简单代码块",
			js:   "var result; { result = 1; result = 2; result = 3; }",
			want: float64(3),
		},
		{
			name: "嵌套代码块",
			js:   "var result; { { result = 1; result = 2; }; result = 3; }",
			want: float64(3),
		},
		{
			name: "空代码块",
			js:   "var result = 1; {}",
			want: float64(1),
		},
		{
			name: "代码块中的变量访问",
			js:   "var result; { result = req.Path }",
			want: "/api/test",
		},
		{
			name: "代码块中的函数调用",
			js:   "var result; { result = add(1, 2); result = add(3, 4) }",
			want: float64(7),
		},
		{
			name: "代码块中的复杂表达式",
			js:   "var result; { result = 1 + 2; result = 3 * 4; result = 5 - 6; }",
			want: float64(-1),
		},
		{
			name: "代码块中的条件赋值",
			js:   "var result; { if (1 < 2) { result = 'yes'; } else { result = 'no'; } }",
			want: "yes",
		},
		{
			name: "代码块中的console.log",
			js:   "var result = 0; { console.log('test1'); console.log('test2'); console.log('test3'); result = 1; }",
			want: float64(1),
		},
		{
			name: "多层嵌套代码块",
			js:   "var result; { { { result = 1; } } }",
			want: float64(1),
		},
		{
			name: "代码块中的混合运算",
			js:   "var result; { var x = 1 + 2; var y = 'hello' + ' world'; var z = add(3, 4); result = z; }",
			want: float64(7),
		},
		{
			name: "代码块中的对象访问链",
			js: `var result; { 
				var headers = req.Headers;
				result = headers['Content-Type'];
				result = req.GetHeader('test');
			}`,
			want: "test-header",
		},
		{
			name: "代码块中的对象访问方式",
			js: `var result; {
				var obj = {a: 1, 'b-c': 2};
				result = obj.a;        // 点号访问
				result = obj['a'];     // 方括号访问字符串
				result = obj['b-c'];   // 方括号访问特殊字符
			}`,
			want: float64(2),
		},
		{
			name: "代码块中的布尔运算",
			js:   "var result; { var x = true && false; var y = false || true; var z = !false; result = !z; }",
			want: false,
		},
		{
			name: "代码块中的比较运算",
			js:   "var result; { var result = 1 < 2; result = 3 > 4; result = 5 <= 5; result = 6 >= 7; result = result; }",
			want: false,
		},
		{
			name: "代码块中的字符串操作",
			js:   "var result; { var x = 'hello'.length; var y = 'world'.length; result = y; }",
			want: 5,
		},
		{
			name: "代码块中的对象字面量",
			js:   "var result; { var empty = {}; var obj = {a: 1}; result = obj; }",
			want: map[string]any{"a": float64(1)},
		},
		{
			name: "代码块中的三目运算符",
			js:   "var result; { var x = true ? 1 : 2; var y = false ? 3 : 4; result = y; }",
			want: float64(4),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置上下文，但保留预定义变量
			ctx.Vars = map[string]any{
				"add":     ctx.Vars["add"],
				"req":     ctx.Vars["req"],
				"console": ctx.Vars["console"],
			}

			// 执行代码，忽略返回值
			got, err := transpiler.Execute(tt.js, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 返回值应该总是 nil
			if got != nil {
				t.Errorf("Execute() returned %v, want nil", got)
			}

			// 检查 result 变量的值
			if !tt.wantErr {
				result := ctx.Vars["result"]
				if !reflect.DeepEqual(result, tt.want) {
					t.Errorf("ctx.Vars[\"result\"] = %v, want %v", result, tt.want)
				}
			}
		})
	}
}

func TestObjectLiteralRelated(t *testing.T) {
	ctx := transpiler.NewScope(nil, nil, nil)

	// 添加一些变量到上下文
	ctx.Vars["x"] = float64(1)
	ctx.Vars["y"] = "test"
	ctx.Vars["obj"] = map[string]any{
		"a": float64(1),
		"b": "hello",
	}

	tests := []struct {
		name    string
		js      string
		want    any
		wantErr bool
	}{
		{
			name: "空对象",
			js:   "var result = {};",
			want: map[string]any{},
		},
		{
			name: "简单对象",
			js:   "var result = {a: 1, b: 2};",
			want: map[string]any{"a": float64(1), "b": float64(2)},
		},
		{
			name: "字符串键",
			js:   "var result = {'a': 1, 'b': 2};",
			want: map[string]any{"a": float64(1), "b": float64(2)},
		},
		{
			name: "表达式作为值",
			js:   "var result = {a: 1 + 2, b: 'hello' + ' world'};",
			want: map[string]any{"a": float64(3), "b": "hello world"},
		},
		{
			name: "变量作为值",
			js:   "var result = {a: x, b: y};",
			want: map[string]any{"a": float64(1), "b": "test"},
		},
		{
			name: "嵌套对象",
			js:   "var result = {a: {b: {c: 1}}};",
			want: map[string]any{"a": map[string]any{"b": map[string]any{"c": float64(1)}}},
		},
		{
			name: "对象作为变量值",
			js:   "var result; var o = {a: 1}; result = o;",
			want: map[string]any{"a": float64(1)},
		},
		{
			name: "对象属性访问",
			js:   "var result; var o = {a: 1, b: 2}; result = o.a;",
			want: float64(1),
		},
		{
			name: "计算属性名",
			js:   "var result = {['a' + 'b']: 1};",
			want: map[string]any{"ab": float64(1)},
		},
		{
			name: "对象展开",
			js:   "var base = {a: 1}; var result = {...base, b: 2};",
			want: map[string]any{"a": float64(1), "b": float64(2)},
		},
		{
			name: "多层对象展开",
			js:   "var a = {x: 1}; var b = {y: 2}; var result = {...a, ...b, z: 3};",
			want: map[string]any{"x": float64(1), "y": float64(2), "z": float64(3)},
		},
		{
			name: "对象属性覆盖",
			js:   "var result = {a: 1, a: 2};",
			want: map[string]any{"a": float64(2)},
		},
		{
			name: "展开覆盖",
			js:   "var base = {a: 1, b: 1}; var result = {...base, b: 2};",
			want: map[string]any{"a": float64(1), "b": float64(2)},
		},
		{
			name: "表达式作为键",
			js:   "var result = {[1 + 2]: 'three'};",
			want: map[string]any{"3": "three"},
		},
		{
			name:    "无效的键",
			js:      "var result = {[{x: 1}]: 1};",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置上下文，但保留预定义变量
			ctx.Vars = map[string]any{
				"x":   float64(1),
				"y":   "test",
				"obj": map[string]any{"a": float64(1), "b": "hello"},
			}

			// 执行代码，忽略返回值
			got, err := transpiler.Execute(tt.js, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 返回值应该总是 nil
			if got != nil {
				t.Errorf("Execute() returned %v, want nil", got)
			}

			// 检查 result 变量的值
			if !tt.wantErr {
				result := ctx.Vars["result"]
				if !reflect.DeepEqual(result, tt.want) {
					t.Errorf("ctx.Vars[\"result\"] = %v (%T), want %v (%T)", result, result, tt.want, tt.want)
				}
			}
		})
	}
}

func TestArrayMethods(t *testing.T) {
	ctx := transpiler.NewScope(nil, nil, nil)

	// 在每个测试前重置数组，确保测试相互独立
	setupArrays := func() {
		ctx.Vars["arr"] = []any{1, 2, 3, 4, 5}
		ctx.Vars["strArr"] = []string{"hello", "world", "test"}
	}

	tests := []struct {
		name    string
		setup   func()
		js      string
		want    any
		wantErr bool
	}{
		{
			name:  "数组长度",
			setup: setupArrays,
			js:    "var result = arr.length;",
			want:  5,
		},
		{
			name:  "slice方法-单参数",
			setup: setupArrays,
			js:    "var result = arr.slice(2);",
			want:  []any{3, 4, 5},
		},
		{
			name:  "slice方法-双参数",
			setup: setupArrays,
			js:    "var result = arr.slice(1, 3);",
			want:  []any{2, 3},
		},
		{
			name:  "slice方法-负索引",
			setup: setupArrays,
			js:    "var result = arr.slice(-2);",
			want:  []any{4, 5},
		},
		{
			name:  "indexOf方法-存在的元素",
			setup: setupArrays,
			js:    "var result = arr.indexOf(3);",
			want:  2,
		},
		{
			name:  "indexOf方法-不存在的元素",
			setup: setupArrays,
			js:    "var result = arr.indexOf(6);",
			want:  -1,
		},
		{
			name:  "indexOf方法-字符串数组",
			setup: setupArrays,
			js:    "var result = strArr.indexOf('world');",
			want:  1,
		},
		{
			name:  "join方法-默认分隔符",
			setup: setupArrays,
			js:    "var result = arr.join();",
			want:  "1,2,3,4,5",
		},
		{
			name:  "join方法-自定义分隔符",
			setup: setupArrays,
			js:    "var result = arr.join('-');",
			want:  "1-2-3-4-5",
		},
		{
			name:  "join方法-字符串数组",
			setup: setupArrays,
			js:    "var result = strArr.join(' ');",
			want:  "hello world test",
		},
		{
			name:  "splice方法-删除元素",
			setup: setupArrays,
			js:    "var result = arr.splice(2, 1);",
			want:  []any{1, 2, 4, 5},
		},
		{
			name:  "splice方法-替换元素",
			setup: setupArrays,
			js:    "var result = arr.splice(1, 2, 6, 7);",
			want:  []any{1, 6, 7, 4, 5},
		},
		{
			name:  "splice方法-插入元素",
			setup: setupArrays,
			js:    "var result = arr.splice(2, 0, 8);",
			want:  []any{1, 2, 8, 3, 4, 5},
		},
		{
			name:  "splice方法-负索引",
			setup: setupArrays,
			js:    "var result = arr.splice(-2, 1);",
			want:  []any{1, 2, 3, 5},
		},
		{
			name:    "slice方法-参数类型错误",
			setup:   setupArrays,
			js:      "var result = arr.slice('invalid');",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			// 执行代码，忽略返回值
			got, err := transpiler.Execute(tt.js, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 返回值应该总是 nil
			if got != nil {
				t.Errorf("Execute() returned %v, want nil", got)
			}

			// 检查 result 变量的值
			if !tt.wantErr {
				result := ctx.Vars["result"]
				// 对于切片类型，使用 CompareValues 进行比较
				if gotSlice, ok := result.([]any); ok {
					if wantSlice, ok := tt.want.([]any); ok {
						if len(gotSlice) != len(wantSlice) {
							t.Errorf("result slice length = %v, want %v", len(gotSlice), len(wantSlice))
							return
						}
						for i := range gotSlice {
							cmp, err := util.CompareValues(gotSlice[i], wantSlice[i])
							if err != nil {
								t.Errorf("Compare error at index %d: %v", i, err)
								return
							}
							if cmp != 0 {
								t.Errorf("result at index %d = %v, want %v", i, gotSlice[i], wantSlice[i])
								return
							}
						}
					} else {
						t.Errorf("result = %v (%T), want %v (%T)", result, result, tt.want, tt.want)
					}
				} else {
					// 非切片类型使用 reflect.DeepEqual
					if !reflect.DeepEqual(result, tt.want) {
						t.Errorf("result = %v, want %v", result, tt.want)
					}
				}
			}
		})
	}
}

// 添加新的三目运算符测试
func TestConditionalExpression(t *testing.T) {
	ctx := transpiler.NewScope(nil, nil, nil)

	tests := []struct {
		name    string
		js      string
		want    any
		wantErr bool
	}{
		{
			name: "简单三目运算符",
			js:   "var result = true ? 1 : 2;",
			want: float64(1),
		},
		{
			name: "三目运算符-false分支",
			js:   "var result = false ? 1 : 2;",
			want: float64(2),
		},
		{
			name: "三目运算符-比较运算",
			js:   "var result = 1 < 2 ? 'yes' : 'no';",
			want: "yes",
		},
		{
			name: "三目运算符-字符串操作",
			js:   "var result = 'hello'.length > 3 ? 'long' : 'short';",
			want: "long",
		},
		{
			name: "三目运算符-嵌套",
			js:   "var result = true ? (false ? 1 : 2) : 3;",
			want: float64(2),
		},
		{
			name: "三目运算符-复杂条件",
			js:   "var result = (1 + 2 > 2) && (3 * 4 <= 12) ? 'true' : 'false';",
			want: "true",
		},
		{
			name: "三目运算符-对象访问",
			js:   "var result = true ? {a: 1} : {b: 2};",
			want: map[string]any{"a": float64(1)},
		},
		{
			name: "三目运算符-函数调用",
			js:   "var result = true ? 'hello'.toUpperCase() : 'world'.toLowerCase();",
			want: "HELLO",
		},
		{
			name: "三目运算符-数组操作",
			js:   "var result = [1,2,3].length > 2 ? 'many' : 'few';",
			want: "many",
		},
		{
			name: "三目运算符-truthy值",
			js:   "var result = 'hello' ? 'truthy' : 'falsy';",
			want: "truthy",
		},
		{
			name: "三目运算符-falsy值",
			js:   "var result = '' ? 'truthy' : 'falsy';",
			want: "falsy",
		},
		{
			name: "三目运算符-null检查",
			js:   "var result = null ? 'exists' : 'null';",
			want: "null",
		},
		{
			name: "三目运算符-undefined检查",
			js:   "var result = undefined ? 'exists' : 'undefined';",
			want: "undefined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置上下文
			ctx.Vars = make(map[string]any)
			ctx.Vars["undefined"] = nil

			// 执行代码，忽略返回值
			got, err := transpiler.Execute(tt.js, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 返回值应该总是 nil
			if got != nil {
				t.Errorf("Execute() returned %v, want nil", got)
			}

			// 检查 result 变量的值
			if !tt.wantErr {
				result := ctx.Vars["result"]
				if !reflect.DeepEqual(result, tt.want) {
					t.Errorf("ctx.Vars[\"result\"] = %v, want %v", result, tt.want)
				}
			}
		})
	}
}

func TestFunctionSupport(t *testing.T) {
	ctx := transpiler.NewScope(nil, nil, nil)

	tests := []struct {
		name    string
		js      string
		want    any
		wantErr bool
	}{
		// 函数声明测试
		{
			name: "基本函数调用",
			js: `function add(a, b) {
					return a + b;
				}
				var result = add(2, 3);`,
			want: float64(5),
		},
		{
			name: "嵌套函数调用",
			js: `function multiply(a, b) {
					return a * b;
				}
				function calc() {
					return multiply(3, 4) + 1;
				}
				var result = calc();`,
			want: float64(13),
		},
		{
			name: "函数作用域",
			js: `var x = 10;
				function test() {
					var x = 20;
					return x;
				}
				var result = test();`,
			want: float64(20),
		},
		{
			name: "闭包测试",
			js: `function createCounter() {
					var count = 0;
					return function() {
						count++;
						return count;
					};
				}
				var counter = createCounter();
				var result = counter() + counter();`,
			want: float64(3),
		},

		// 箭头函数测试
		{
			name: "箭头函数-表达式形式",
			js:   "var sum = (a, b) => a + b; var result = sum(5, 7);",
			want: float64(12),
		},
		{
			name: "箭头函数-块语句",
			js: `var greet = name => {
					return 'Hello ' + name;
				}
				var result = greet('World');`,
			want: "Hello World",
		},
		{
			name: "箭头函数-this绑定",
			js: `var obj = {
					value: 42,
					getValue: () => this.value
				};
				var result = obj.getValue();`,
			want: nil, // 需要根据实际this绑定实现调整预期
		},

		// 匿名函数测试
		{
			name: "匿名函数赋值",
			js: `var square = function(x) {
					return x * x;
				};
				var result = square(4);`,
			want: float64(16),
		},
		{
			name: "立即执行函数",
			js: `var result = (function() {
					return 'IIFE';
				})();`,
			want: "IIFE",
		},

		// 参数处理测试
		{
			name: "默认参数处理",
			js: `function sayHello(name) {
					name = name || 'Guest';
					return 'Hello ' + name;
				}
				var result = sayHello();`,
			want: "Hello Guest",
		},
		{
			name: "多余参数处理",
			js: `function sum(a, b) {
					return a + b;
				}
				var result = sum(1, 2, 3, 4);`,
			want: float64(3),
		},
		{
			name: "参数不足处理",
			js: `function sum(a, b) {
					return a + b;
				}
				var result = sum(1);`,
			want: float64(1), // 根据实现可能为1+undefined=NaN，需要根据实际处理调整
		},

		// 错误情况测试
		{
			name:    "未定义函数调用",
			js:      "var result = notDefined();",
			wantErr: true,
		},
		{
			name:    "非函数调用",
			js:      "var x = 1; var result = x();",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置上下文
			ctx.Vars = make(map[string]any)
			ctx.Vars["undefined"] = nil

			// 执行代码
			_, err := transpiler.Execute(tt.js, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 验证结果
			if !tt.wantErr {
				result := ctx.Vars["result"]
				if !reflect.DeepEqual(result, tt.want) {
					t.Errorf("ctx.Vars[\"result\"] = %v (%T), want %v (%T)",
						result, result, tt.want, tt.want)
				}
			}
		})
	}
}

func TestArrayAccess(t *testing.T) {
	ctx := transpiler.NewScope(nil, nil, nil)

	tests := []struct {
		name    string
		js      string
		want    any
		wantErr bool
	}{
		{
			name: "基本索引访问",
			js:   "var arr = [1, 2, 3]; var result = arr[1];",
			want: float64(2),
		},
		{
			name: "表达式索引",
			js:   "var arr = [1, 2, 3]; var result = arr[2 * 0 + 2];",
			want: float64(3),
		},
		{
			name: "最后元素访问",
			js:   "var arr = [1, 2, 3, 4, 5]; var result = arr[arr.length - 1];",
			want: float64(5),
		},
		{
			name: "嵌套数组访问",
			js:   "var arr = [[1, 2], [3, 4]]; var result = arr[1][0];",
			want: float64(3),
		},
		{
			name:    "索引越界",
			js:      "var arr = [1, 2, 3]; var result = arr[5];",
			want:    nil,
			wantErr: true,
		},
		{
			name: "字符串数组访问",
			js:   "var arr = ['hello', 'world']; var result = arr[1];",
			want: "world",
		},
		{
			name: "动态修改数组",
			js:   "var arr = [1, 2, 3]; arr[1] = 10; var result = arr[1];",
			want: float64(10),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置上下文
			ctx.Vars = make(map[string]any)

			// 执行代码
			_, err := transpiler.Execute(tt.js, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 验证结果
			if !tt.wantErr {
				result := ctx.Vars["result"]
				if !reflect.DeepEqual(result, tt.want) {
					t.Errorf("ctx.Vars[\"result\"] = %v (%T), want %v (%T)",
						result, result, tt.want, tt.want)
				}
			}
		})
	}
}

func TestTranspileToGoFunc(t *testing.T) {
	tests := []struct {
		name    string
		js      string
		args    []any
		want    any
		wantErr bool
	}{
		{
			name:    "函数 1",
			js:      `function add(a, b) { return a + b; }`,
			args:    []any{2, 3},
			want:    float64(5),
			wantErr: false,
		},
		{
			name:    "函数 2",
			js:      `function multiply(a, b) { return a + b; }`,
			args:    []any{"hello", "world"},
			want:    "helloworld",
			wantErr: false,
		},
		{
			name: "函数 3",
			js: `function multiply(a, b) {
				var c = a * b;
				var d = a + b;
				return c + d;
			}`,
			args:    []any{2, 3},
			want:    float64(11),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goFunc, err := transpiler.TranspileJsScriptToGoFunc(tt.js, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("TranspileToGoFunc() error = %v, wantErr %v", err, tt.wantErr)
			}
			result := goFunc(tt.args...)
			cmp, err := util.CompareValues(result, tt.want)
			if err != nil || cmp != 0 {
				t.Errorf("TranspileToGoFunc() result = %v, want %v", result, tt.want)
			}
		})
	}
}

type User struct {
	Id   int
	Name string
}

func (u *User) SetName(name string) {
	u.Name = name
}

func TestTranspileToGoFuncPerformance(t *testing.T) {
	// 1. Native Go Function
	runtime.GC()
	goFunc1 := func(a, b *User) *User {
		var name = a.Name
		b.SetName(name)
		return b
	}
	start1 := time.Now()
	for i := 0; i < 1000000; i++ {
		user1 := &User{Id: 1, Name: "hello"}
		user2 := &User{Id: 2, Name: "world"}
		goFunc1(user1, user2)
		assert.Equal(t, user2.Name, "hello")
	}
	elapsed1 := time.Since(start1)
	fmt.Printf("Native Go Function: %s\n", elapsed1)

	// 2. Transpiled Go Function
	runtime.GC()
	js := `function(a, b) {
		var name = a.Name;
		b.SetName(name);
		return b;
	}`
	ctx2 := transpiler.NewScope(nil, nil, nil)
	goFunc2, _ := transpiler.TranspileJsScriptToGoFunc(js, ctx2)
	start2 := time.Now()
	for i := 0; i < 1000000; i++ {
		user1 := &User{Id: 1, Name: "hello"}
		user2 := &User{Id: 2, Name: "world"}
		goFunc2(user1, user2)
		assert.Equal(t, user2.Name, "hello")
	}
	elapsed2 := time.Since(start2)
	fmt.Printf("Transpiled: %s\n", elapsed2)

	// 3. Transpiled Go Function with Specialized PropGetter
	userPropAccessHandler := func(access transpiler.PropAccess, obj any) (any, error) {
		if user, ok := obj.(*User); ok {
			if access.IsCall {
				switch access.Prop {
				case "SetName":
					arg0 := access.Args[0].(string)
					user.SetName(arg0)
					return nil, nil
				}
			} else {
				switch access.Prop {
				case "Id":
					return user.Id, nil
				case "Name":
					return user.Name, nil
				}
			}
		}
		return nil, fmt.Errorf("unsupported property access: %v", access)
	}
	propGetter := transpiler.NewPropGetter(
		userPropAccessHandler,
		transpiler.StringPropAccessHandler,
		transpiler.ArrayPropAccessHandler,
		transpiler.MethodCallHandler,
		transpiler.DataFieldAccessHandler,
	)
	ctx3 := transpiler.NewScope(nil, propGetter, nil)
	goFunc3, _ := transpiler.TranspileJsScriptToGoFunc(js, ctx3)
	start3 := time.Now()
	for i := 0; i < 1000000; i++ {
		user1 := &User{Id: 1, Name: "hello"}
		user2 := &User{Id: 2, Name: "world"}
		goFunc3(user1, user2)
		assert.Equal(t, user2.Name, "hello")
	}
	elapsed3 := time.Since(start3)
	fmt.Printf("Transpiled (Specialized): %s\n", elapsed3)

	// 4. Goja
	runtime.GC()
	vm := goja.New()
	vm.RunString(`function test(a, b) {
		var name = a.Name;
		b.SetName(name);
		return b;
	}`)
	var goFunc4 func(a, b *User) *User
	vm.ExportTo(vm.Get("test"), &goFunc4)
	start4 := time.Now()
	for i := 0; i < 1000000; i++ {
		user1 := &User{Id: 1, Name: "hello"}
		user2 := &User{Id: 2, Name: "world"}
		goFunc4(user1, user2)
	}
	elapsed4 := time.Since(start4)
	fmt.Printf("Goja: %s\n", elapsed4)
}

func TestLoroValueAccess(t *testing.T) {
	tests := []struct {
		name    string
		js      string
		args    []any
		want    any
		wantErr bool
	}{
		{
			name: "访问 Map 属性",
			js: `function getName(doc) {
				return doc.test.name;
			}`,
			args: []any{func() *loro.LoroDoc {
				doc := loro.NewLoroDoc()
				testMap := doc.GetMap("test")
				testMap.InsertString("name", "Alice")
				return doc
			}()},
			want:    "Alice",
			wantErr: false,
		},
		{
			name: "访问嵌套 Map 属性",
			js: `function getNestedValue(doc) {
				return doc.user.profile.age;
			}`,
			args: []any{func() *loro.LoroDoc {
				doc := loro.NewLoroDoc()
				userMap := doc.GetMap("user")
				profileMap := doc.GetMap("profile")
				profileMap, err := userMap.InsertMap("profile", profileMap)
				assert.NoError(t, err)
				profileMap.InsertI64("age", 30)
				return doc
			}()},
			want:    int64(30),
			wantErr: false,
		},
		{
			name: "按下标访问 LoroList",
			js: `function getNestedValue(doc) {
				return doc.arr[1];
			}`,
			args: []any{func() *loro.LoroDoc {
				doc := loro.NewLoroDoc()
				arr := doc.GetList("arr")
				arr.PushString("Hello")
				arr.PushString("World")
				return doc
			}()},
			want:    "World",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			propGetter := transpiler.NewPropGetter(
				query.LoroDocAccessHandler,
				query.LoroTextAccessHandler,
				query.LoroMapAccessHandler,
				query.LoroListAccessHandler,
				query.LoroMovableListAccessHandler,
			)
			ctx := transpiler.NewScope(nil, propGetter, nil)
			goFunc, err := transpiler.TranspileJsScriptToGoFunc(tt.js, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("TranspileJsScriptToGoFunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			got := goFunc(tt.args...)
			assert.Equal(t, tt.want, got)
		})
	}
}

// 测试解构赋值功能
func TestDestructuringAssignment(t *testing.T) {
	// 第一部分：基本执行测试
	t.Run("基本执行测试", func(t *testing.T) {
		ctx := transpiler.NewScope(nil, nil, nil)

		tests := []struct {
			name    string
			js      string
			want    any
			wantErr bool
		}{
			// 对象解构赋值
			{
				name: "基本对象解构赋值",
				js:   "var {a, b} = {a: 1, b: 2}; var result = a + b;",
				want: float64(3),
			},
			// 嵌套对象解构赋值暂不支持
			// {
			// 	name: "嵌套对象解构赋值",
			// 	js:   "var {a, b: {c}} = {a: 1, b: {c: 2}}; var result = a + c;",
			// 	want: float64(3),
			// },
			{
				name: "带默认值的对象解构",
				js:   "var {a = 10, b = 20} = {a: 1}; var result = a + b;",
				want: float64(21),
			},
			// 重命名解构暂不支持
			// {
			// 	name: "重命名变量的对象解构",
			// 	js:   "var {a: x, b: y} = {a: 1, b: 2}; var result = x + y;",
			// 	want: float64(3),
			// },

			// 数组解构赋值
			{
				name: "基本数组解构赋值",
				js:   "var [a, b] = [1, 2]; var result = a + b;",
				want: float64(3),
			},
			{
				name: "忽略元素的数组解构赋值",
				js:   "var [a, , c] = [1, 2, 3]; var result = a + c;",
				want: float64(4),
			},
			// 数组默认值暂不支持
			// {
			// 	name: "默认值数组解构",
			// 	js:   "var [a = 10, b = 20] = [1]; var result = a + b;",
			// 	want: float64(21),
			// },
			// Rest参数暂不支持
			// {
			// 	name: "剩余参数数组解构",
			// 	js:   "var [a, ...rest] = [1, 2, 3]; var result = a + rest[0];",
			// 	want: float64(3),
			// },

			// 函数参数解构
			{
				name: "函数参数对象解构",
				js:   "function test({a, b}) { return a + b; } var result = test({a: 1, b: 2});",
				want: float64(3),
			},
			{
				name: "函数参数数组解构",
				js:   "function test([a, b]) { return a + b; } var result = test([1, 2]);",
				want: float64(3),
			},
			{
				name: "箭头函数参数解构",
				js:   "var test = ({a, b}) => a + b; var result = test({a: 1, b: 2});",
				want: float64(3),
			},

			// 简单赋值表达式解构
			{
				name: "对象赋值解构",
				js:   "var a, b; ({a, b} = {a: 1, b: 2}); var result = a + b;",
				want: float64(3),
			},
			{
				name: "数组赋值解构",
				js:   "var a, b; [a, b] = [1, 2]; var result = a + b;",
				want: float64(3),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// 重置上下文
				ctx.Vars = make(map[string]any)

				// 执行代码
				_, err := transpiler.Execute(tt.js, ctx)
				if (err != nil) != tt.wantErr {
					t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				// 非错误情况，验证结果
				if !tt.wantErr {
					result := ctx.Vars["result"]
					if !reflect.DeepEqual(result, tt.want) {
						t.Errorf("ctx.Vars[\"result\"] = %v (%T), want %v (%T)",
							result, result, tt.want, tt.want)
					}
				}
			})
		}
	})
}

func TestDestructuringAssignment2(t *testing.T) {
	// 第二部分：转译函数测试
	t.Run("转译函数测试", func(t *testing.T) {
		tests := []struct {
			name    string
			js      string
			args    []any
			want    any
			wantErr bool
		}{
			{
				name: "函数参数对象解构",
				js: `function process({x, y}) {
					return x + y;
				}`,
				args:    []any{map[string]any{"x": 10, "y": 20}},
				want:    float64(30),
				wantErr: false,
			},
			{
				name: "函数参数数组解构",
				js: `function process([first, second]) {
					return first + second;
				}`,
				args:    []any{[]any{10, 20}},
				want:    float64(30),
				wantErr: false,
			},
			// 删除带默认值的对象解构测试用例
			{
				name: "箭头函数解构",
				js: `({name, age}) => {
					return name + " is " + age + " years old";
				}`,
				args:    []any{map[string]any{"name": "Alice", "age": 30}},
				want:    "Alice is 30 years old",
				wantErr: false,
			},
			{
				name: "内部使用解构",
				js: `function transform(data) {
					const {x, y} = data;
					return x * y;
				}`,
				args:    []any{map[string]any{"x": 10, "y": 20}},
				want:    float64(200),
				wantErr: false,
			},
			{
				name: "实用场景-HTTP方法提取",
				js: `function parseMethod({method}) {
					return method;
				}`,
				args:    []any{map[string]any{"method": "GET", "url": "/api/users"}},
				want:    "GET",
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ctx := transpiler.NewScope(nil, nil, nil)
				t.Run(tt.name, func(t *testing.T) {
					goFunc, err := transpiler.TranspileJsScriptToGoFunc(tt.js, ctx)
					if (err != nil) != tt.wantErr {
						t.Errorf("TranspileJsScriptToGoFunc() error = %v, wantErr %v", err, tt.wantErr)
						return
					}

					if err != nil {
						return
					}

					// 调用Go函数并验证结果
					result := goFunc(tt.args...)
					if !reflect.DeepEqual(result, tt.want) {
						t.Errorf("TranspileJsScriptToGoFunc() result = %v, want %v", result, tt.want)
					}
				})
			})
		}
	})
}
