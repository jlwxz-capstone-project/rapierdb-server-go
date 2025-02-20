package main

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/transpiler"
)

func TestExecuteCode(t *testing.T) {
	ctx := transpiler.NewContext()

	// 添加一些测试用的函数和变量
	ctx.Vars["add"] = func(a, b float64) float64 { return a + b }
	ctx.Vars["concat"] = func(a, b string) string { return a + b }
	ctx.Vars["console"] = struct {
		Log func(a ...interface{}) (n int, err error)
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
	ctx.PropAccessTransformer = func(chain []transpiler.PropAccessor, obj interface{}) (interface{}, error) {
		// 如果是 console 对象，处理 log -> Log 的转换
		if console, ok := obj.(struct {
			Log func(...interface{}) (int, error)
		}); ok {
			if len(chain) == 1 && chain[0].Prop == "log" {
				return console.Log, nil
			}
		}
		// 其他情况使用默认转换器
		return transpiler.DefaultPropAccessTransformer(chain, obj)
	}

	tests := []struct {
		name    string
		js      string
		want    interface{}
		wantErr bool
	}{
		// 基本运算测试
		{
			name: "数字运算",
			js:   "1 + 2 * 3;",
			want: float64(7),
		},
		{
			name: "字符串拼接",
			js:   "'hello' + ' ' + 'world';",
			want: "hello world",
		},
		{
			name: "函数调用",
			js:   "add(1, 2);",
			want: float64(3),
		},

		// 对象访问测试
		{
			name: "对象属性访问",
			js:   "req.Path;",
			want: "/api/test",
		},
		{
			name: "对象方法调用",
			js:   "req.GetHeader('Content-Type');",
			want: "test-header",
		},
		{
			name: "Map访问",
			js:   "req.Headers['Content-Type'];",
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
			name:    "console.log调用",
			js:      "console.log('test');",
			want:    5,
			wantErr: false,
		},

		// if-else 语句测试
		{
			name: "简单if条件",
			js:   "if (true) 1; else 2;",
			want: float64(1),
		},
		{
			name: "if-else条件",
			js:   "if (false) 1; else 2;",
			want: float64(2),
		},
		{
			name: "比较运算",
			js:   "if (1 < 2) 'yes'; else 'no';",
			want: "yes",
		},
		{
			name: "相等比较",
			js:   "if ('hello' == 'hello') true; else false;",
			want: true,
		},
		{
			name: "不相等比较",
			js:   "if (1 != 2) 'different'; else 'same';",
			want: "different",
		},
		{
			name: "数字条件",
			js:   "if (1) 'truthy'; else 'falsy';",
			want: "truthy",
		},
		{
			name: "零值条件",
			js:   "if (0) 'truthy'; else 'falsy';",
			want: "falsy",
		},
		{
			name: "空字符串条件",
			js:   "if ('') 'truthy'; else 'falsy';",
			want: "falsy",
		},
		{
			name: "null条件",
			js:   "if (null) 'truthy'; else 'falsy';",
			want: "falsy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transpiler.Execute(tt.js, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomPropertyAccess(t *testing.T) {
	// 定义一个测试结构体
	type TestStruct struct {
		data map[string]interface{}
	}

	ctx := transpiler.NewContext()

	// 创建测试对象
	testObj := &TestStruct{
		data: map[string]interface{}{
			"a": &TestStruct{
				data: map[string]interface{}{
					"b": &TestStruct{
						data: map[string]interface{}{
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
	ctx.PropAccessTransformer = func(chain []transpiler.PropAccessor, obj interface{}) (interface{}, error) {
		if ts, ok := obj.(*TestStruct); ok {
			result := ts
			for _, access := range chain {
				if val, ok := result.data[access.Prop]; ok {
					if next, ok := val.(*TestStruct); ok {
						result = next
					} else {
						return val, nil
					}
				} else {
					return nil, fmt.Errorf("属性不存在: %s", access.Prop)
				}
			}
			return result, nil
		}
		return transpiler.DefaultPropAccessTransformer(chain, obj)
	}

	tests := []struct {
		name    string
		js      string
		want    interface{}
		wantErr bool
	}{
		{
			name: "连续属性访问",
			js:   "obj.a.b.c;",
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
			got, err := transpiler.Execute(tt.js, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChainPropertyAccess(t *testing.T) {
	ctx := transpiler.NewContext()

	type Person struct {
		name    string
		age     int
		friends map[string]*Person
	}

	alice := &Person{
		name: "Alice",
		age:  20,
		friends: map[string]*Person{
			"bob": {name: "Bob", age: 22},
		},
	}

	ctx.Vars["person"] = alice

	// 设置自定义属性访问转换器
	ctx.PropAccessTransformer = func(chain []transpiler.PropAccessor, obj interface{}) (interface{}, error) {
		p, ok := obj.(*Person)
		if !ok {
			return nil, fmt.Errorf("不是 Person 对象")
		}

		result := p
		for _, access := range chain {
			switch {
			case access.Prop == "friend" && access.IsCall:
				// friend('name') 转换为 map 访问
				if len(access.Args) != 1 {
					return nil, fmt.Errorf("friend 方法需要一个参数")
				}
				name, ok := access.Args[0].(string)
				if !ok {
					return nil, fmt.Errorf("friend 方法参数必须是字符串")
				}
				result = result.friends[name]
				if result == nil {
					return nil, fmt.Errorf("friend not found: %s", name)
				}
			case access.Prop == "name" || access.Prop == "age":
				// 属性访问转换为字符串
				return fmt.Sprintf("%s: %v", access.Prop, reflect.ValueOf(result).Elem().FieldByName(access.Prop)), nil
			default:
				return nil, fmt.Errorf("不支持的属性访问: %s", access.Prop)
			}
		}
		return result, nil
	}

	tests := []struct {
		name    string
		js      string
		want    interface{}
		wantErr bool
	}{
		{
			name: "访问直接属性",
			js:   "person.name;",
			want: "name: Alice",
		},
		{
			name: "访问朋友属性",
			js:   "person.friend('bob').name;",
			want: "name: Bob",
		},
		{
			name:    "访问不存在的朋友",
			js:      "person.friend('david').name;",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transpiler.Execute(tt.js, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecuteIfStatement(t *testing.T) {
	ctx := transpiler.NewContext()

	tests := []struct {
		name    string
		js      string
		want    interface{}
		wantErr bool
	}{
		{
			name: "简单if条件",
			js:   "if (true) 1; else 2;",
			want: float64(1),
		},
		{
			name: "if-else条件",
			js:   "if (false) 1; else 2;",
			want: float64(2),
		},
		{
			name: "比较运算",
			js:   "if (1 < 2) 'yes'; else 'no';",
			want: "yes",
		},
		{
			name: "相等比较",
			js:   "if ('hello' == 'hello') true; else false;",
			want: true,
		},
		{
			name: "不相等比较",
			js:   "if (1 != 2) 'different'; else 'same';",
			want: "different",
		},
		{
			name: "数字条件",
			js:   "if (1) 'truthy'; else 'falsy';",
			want: "truthy",
		},
		{
			name: "零值条件",
			js:   "if (0) 'truthy'; else 'falsy';",
			want: "falsy",
		},
		{
			name: "空字符串条件",
			js:   "if ('') 'truthy'; else 'falsy';",
			want: "falsy",
		},
		{
			name: "null条件",
			js:   "if (null) 'truthy'; else 'falsy';",
			want: "falsy",
		},
		{
			name: "复杂条件",
			js:   "if (1 < 2 && 'hello' == 'hello') 'both true'; else 'not both true';",
			want: "both true",
		},
		{
			name: "带括号的条件",
			js:   "if ((1 + 2) * 3 > 8) 'greater'; else 'less';",
			want: "greater",
		},
		{
			name: "多重比较",
			js:   "if (1 < 2 && 3 > 2 || false) 'true'; else 'false';",
			want: "true",
		},
		{
			name: "带括号的逻辑运算",
			js:   "if ((true && false) || (true && true)) 'yes'; else 'no';",
			want: "yes",
		},
		{
			name: "混合运算优先级",
			js:   "if (2 + 3 * 4 > 10 + 2) 'yes'; else 'no';",
			want: "yes",
		},
		{
			name: "复杂嵌套条件",
			js:   "if ((1 + 2 > 2) && (3 * 4 <= 12 || true)) 'complex'; else 'simple';",
			want: "complex",
		},
		{
			name: "字符串比较和数字运算",
			js:   "if ('hello'.length > 2 + 1) 'long'; else 'short';",
			want: "long",
		},
		{
			name: "多重括号和运算符",
			js:   "if (((1 + 2) * 3 == 9) && (4 + 5 >= 8 || 2 * 3 < 5)) 'pass'; else 'fail';",
			want: "pass",
		},
		{
			name: "逻辑运算短路",
			js:   "if (false && someUndefinedVar) 'bug'; else 'ok';",
			want: "ok",
		},
		{
			name: "数字和布尔混合运算",
			js:   "if (1 + 1 == 2 && true || false && 5 < 3) 'correct'; else 'wrong';",
			want: "correct",
		},
		{
			name: "复杂的真值判断",
			js:   "if (1 && 'hello' && (2 * 3) && {}) 'truthy'; else 'falsy';",
			want: "truthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transpiler.Execute(tt.js, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecuteBlockStatement(t *testing.T) {
	ctx := transpiler.NewContext()

	// 添加一些测试用的函数和变量
	ctx.Vars["add"] = func(a, b float64) float64 { return a + b }
	ctx.Vars["console"] = struct {
		Log func(a ...interface{}) (n int, err error)
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
	ctx.PropAccessTransformer = func(chain []transpiler.PropAccessor, obj interface{}) (interface{}, error) {
		// 如果是 console 对象，处理 log -> Log 的转换
		if console, ok := obj.(struct {
			Log func(...interface{}) (int, error)
		}); ok {
			if len(chain) == 1 && chain[0].Prop == "log" {
				return console.Log, nil
			}
		}
		// 其他情况使用默认转换器
		return transpiler.DefaultPropAccessTransformer(chain, obj)
	}

	tests := []struct {
		name    string
		js      string
		want    interface{}
		wantErr bool
	}{
		{
			name: "简单代码块",
			js:   "if (true) { 1; 2; 3; } else { 4; 5; 6; }",
			want: float64(3),
		},
		{
			name: "嵌套代码块",
			js:   "if (true) { if (true) { 1; 2; } else { 3; } } else { 4; }",
			want: float64(2),
		},
		{
			name: "空代码块",
			js:   "if (true) {} else { 1; }",
			want: nil,
		},
		{
			name: "代码块中的变量访问",
			js:   "if (true) { req.Path; req.Headers['Content-Type']; }",
			want: "application/json",
		},
		{
			name: "代码块中的函数调用",
			js:   "if (true) { add(1, 2); add(3, 4); }",
			want: float64(7),
		},
		{
			name: "代码块中的复杂表达式",
			js:   "if (true) { 1 + 2; 3 * 4; 5 - 6; }",
			want: float64(-1),
		},
		{
			name: "代码块中的条件语句",
			js:   "if (true) { if (1 < 2) { 'yes'; } else { 'no'; } }",
			want: "yes",
		},
		{
			name: "代码块中的console.log",
			js:   "if (true) { console.log('test1'); console.log('test2'); console.log('test3'); }",
			want: 6,
		},
		{
			name: "多层嵌套代码块",
			js:   "if (true) { if (true) { if (true) { 1; } else { 2; } } else { 3; } } else { 4; }",
			want: float64(1),
		},
		{
			name: "代码块中的混合运算",
			js:   "if (true) { 1 + 2; 'hello' + ' world'; add(3, 4); }",
			want: float64(7),
		},
		{
			name: "代码块中的对象访问链",
			js:   "if (true) { req.Headers['Content-Type']; req.GetHeader('test'); }",
			want: "test-header",
		},
		{
			name: "代码块中的布尔运算",
			js:   "if (true) { true && false; false || true; !false; !true; }",
			want: false,
		},
		{
			name: "代码块中的比较运算",
			js:   "if (true) { 1 < 2; 3 > 4; 5 <= 5; 6 >= 7; }",
			want: false,
		},
		{
			name: "代码块中的字符串操作",
			js:   "if (true) { 'hello'.length; 'world'.length; }",
			want: 5,
		},
		{
			name: "代码块中的对象字面量",
			js:   "if (true) { var empty = {}; var obj = {a: 1}; obj; }",
			want: map[string]any{"a": float64(1)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transpiler.Execute(tt.js, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestObjectLiteralRelated(t *testing.T) {
	ctx := transpiler.NewContext()

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
		want    interface{}
		wantErr bool
	}{
		{
			name: "空对象",
			js:   "({})",
			want: map[string]any{},
		},
		{
			name: "简单对象",
			js:   "({a: 1, b: 2})",
			want: map[string]any{"a": float64(1), "b": float64(2)},
		},
		{
			name: "字符串键",
			js:   "({'a': 1, 'b': 2})",
			want: map[string]any{"a": float64(1), "b": float64(2)},
		},
		{
			name: "表达式作为值",
			js:   "({a: 1 + 2, b: 'hello' + ' world'})",
			want: map[string]any{"a": float64(3), "b": "hello world"},
		},
		{
			name: "变量作为值",
			js:   "({a: x, b: y})",
			want: map[string]any{"a": float64(1), "b": "test"},
		},
		{
			name: "嵌套对象",
			js:   "({a: {b: {c: 1}}})",
			want: map[string]any{"a": map[string]any{"b": map[string]any{"c": float64(1)}}},
		},
		{
			name: "对象作为变量值",
			js:   "var o = {a: 1}; o;",
			want: map[string]any{"a": float64(1)},
		},
		{
			name: "对象属性访问",
			js:   "var o = {a: 1, b: 2}; o.a;",
			want: float64(1),
		},
		{
			name: "计算属性名",
			js:   "({['a' + 'b']: 1})",
			want: map[string]any{"ab": float64(1)},
		},
		{
			name: "对象展开",
			js:   "var base = {a: 1}; ({...base, b: 2})",
			want: map[string]any{"a": float64(1), "b": float64(2)},
		},
		{
			name: "多层对象展开",
			js:   "var a = {x: 1}; var b = {y: 2}; ({...a, ...b, z: 3})",
			want: map[string]any{"x": float64(1), "y": float64(2), "z": float64(3)},
		},
		{
			name: "对象属性覆盖",
			js:   "({a: 1, a: 2})",
			want: map[string]any{"a": float64(2)},
		},
		{
			name: "展开覆盖",
			js:   "var base = {a: 1, b: 1}; ({...base, b: 2})",
			want: map[string]any{"a": float64(1), "b": float64(2)},
		},
		{
			name: "表达式作为键",
			js:   "({[1 + 2]: 'three'})",
			want: map[string]any{"3": "three"},
		},
		{
			name:    "无效的键",
			js:      "({[{x: 1}]: 1})",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transpiler.Execute(tt.js, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Execute() = %v (%T), want %v (%T)", got, got, tt.want, tt.want)

					// 打印详细的值比较
					// if m1, ok1 := got.(map[string]any); ok1 {
					// 	if m2, ok2 := tt.want.(map[string]any); ok2 {
					// 		for k, v1 := range m1 {
					// 			if v2, exists := m2[k]; exists {
					// 				t.Logf("Key %s: got %v (%T), want %v (%T)", k, v1, v1, v2, v2)
					// 			}
					// 		}
					// 	}
					// }
				}
			}
		})
	}
}
