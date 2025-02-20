package main

import (
	"fmt"
	"reflect"
	"strings"
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

	// 为 console 对象添加特殊的属性访问规则
	consoleType := reflect.TypeOf(ctx.Vars["console"])
	ctx.AddPropAccessRule(consoleType, func(obj interface{}, prop string) (interface{}, error) {
		// 将 log 转换为 Log
		if prop == "log" {
			val := reflect.ValueOf(obj)
			method := val.FieldByName("Log")
			if method.IsValid() {
				return method.Interface(), nil
			}
		}
		return nil, fmt.Errorf("属性不存在: %s", prop)
	})

	tests := []struct {
		name    string
		js      string
		want    interface{}
		wantErr bool
	}{
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
	}

	// 添加 console.log 测试用例
	tests = append(tests, struct {
		name    string
		js      string
		want    interface{}
		wantErr bool
	}{
		name:    "console.log调用",
		js:      "console.log('test');",
		want:    5,
		wantErr: false,
	})

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

	// 添加自定义属性访问规则
	ctx.AddPropAccessRule(reflect.TypeOf(TestStruct{}), func(obj interface{}, prop string) (interface{}, error) {
		ts := obj.(*TestStruct)
		if val, ok := ts.data[prop]; ok {
			return val, nil
		}
		// 转换为 Get 方法调用
		method := reflect.ValueOf(ts).MethodByName("Get" + strings.Title(prop))
		if method.IsValid() {
			results := method.Call(nil)
			if len(results) > 0 {
				return results[0].Interface(), nil
			}
		}
		return nil, fmt.Errorf("属性不存在: %s", prop)
	})

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
