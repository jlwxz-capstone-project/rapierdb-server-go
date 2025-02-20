package transpiler

import (
	"fmt"
	"reflect"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/ast"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/parser"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/token"
)

// Context 定义转译上下文
type Context struct {
	// 变量和函数映射表，让 js 代码可以访问这些变量和函数
	Vars map[string]any
	// 属性访问转换器
	PropAccessTransformer func(chain []PropAccessor, obj any) (any, error)
}

// PropAccessor 表示一次属性访问
//
// 例如：obj.method(arg1, arg2) 对应的 PropAccessor 为：
//
//	PropAccessor{
//		Prop: "method",
//		Args: []any{"arg1", "arg2"},
//		IsCall: true,
//	}
//
// obj.name 对应的 PropAccessor 为：
//
//	PropAccessor{
//		Prop: "name",
//	}
type PropAccessor struct {
	// 属性名
	Prop string
	// 如果是函数调用，这里是参数
	Args []any
	// 是否是函数调用
	IsCall bool
}

// NewContext 创建新的转译上下文
func NewContext() *Context {
	return &Context{
		Vars:                  make(map[string]any),
		PropAccessTransformer: DefaultPropAccessTransformer,
	}
}

// DefaultPropAccessTransformer 默认的属性访问转换器
//
// chain 是属性访问链，obj 是根对象。比如 obj.name.slice(1, 2).toUpperCase() 对应的 chain 为：
//
//	chain = []PropAccessor{
//		{Prop: "name"},
//		{Prop: "slice", Args: []any{1, 2}, IsCall: true},
//		{Prop: "toUpperCase", IsCall: true},
//	}
func DefaultPropAccessTransformer(chain []PropAccessor, obj any) (any, error) {
	result := obj
	for _, access := range chain {
		val := reflect.ValueOf(result)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if access.IsCall {
			// 处理方法调用
			method := val.MethodByName(access.Prop)
			if !method.IsValid() {
				return nil, fmt.Errorf("方法不存在: %s", access.Prop)
			}
			args := make([]reflect.Value, len(access.Args))
			for i, arg := range access.Args {
				args[i] = reflect.ValueOf(arg)
			}
			results := method.Call(args)
			if len(results) == 0 {
				return nil, nil
			}
			result = results[0].Interface()
		} else {
			// 处理属性访问
			if val.Kind() == reflect.Struct {
				field := val.FieldByName(access.Prop)
				if !field.IsValid() {
					return nil, fmt.Errorf("属性不存在: %s", access.Prop)
				}
				result = field.Interface()
			} else if val.Kind() == reflect.Map {
				mapVal := val.MapIndex(reflect.ValueOf(access.Prop))
				if !mapVal.IsValid() {
					return nil, fmt.Errorf("属性不存在: %s", access.Prop)
				}
				result = mapVal.Interface()
			} else {
				return nil, fmt.Errorf("不支持的对象类型: %v", val.Kind())
			}
		}
	}
	return result, nil
}

// Execute 执行 JavaScript 代码并返回结果
func Execute(js string, ctx *Context) (any, error) {
	if ctx == nil {
		ctx = NewContext()
	}

	program, err := parser.ParseFile(js)
	if err != nil {
		return nil, fmt.Errorf("解析错误: %v", err)
	}

	var result any
	for _, stmt := range program.Body {
		result, err = executeStatement(stmt.Stmt, ctx)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func executeStatement(stmt ast.Stmt, ctx *Context) (any, error) {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		return executeExpression(s.Expression.Expr, ctx)
	default:
		return nil, fmt.Errorf("暂不支持的语句类型: %T", stmt)
	}
}

func executeExpression(expr ast.Expr, ctx *Context) (any, error) {
	switch e := expr.(type) {
	case *ast.NumberLiteral:
		return e.Value, nil
	case *ast.StringLiteral:
		return e.Value, nil
	case *ast.Identifier:
		if val, ok := ctx.Vars[e.Name]; ok {
			return val, nil
		}
		return nil, fmt.Errorf("未定义的标识符: %s", e.Name)

	case *ast.MemberExpression:
		// 收集属性访问链
		chain := []PropAccessor{}
		current := expr
		var rootExpr ast.Expr

		// 构建访问链
		for {
			if member, ok := current.(*ast.MemberExpression); ok {
				access := PropAccessor{}

				// 处理属性名
				if id, ok := member.Property.Prop.(*ast.Identifier); ok {
					access.Prop = id.Name
				} else if computed, ok := member.Property.Prop.(*ast.ComputedProperty); ok {
					prop, err := executeExpression(computed.Expr.Expr, ctx)
					if err != nil {
						return nil, err
					}
					access.Prop = fmt.Sprint(prop)
				}

				// 处理函数调用
				if call, ok := expr.(*ast.CallExpression); ok && call.Callee.Expr == current {
					access.IsCall = true
					access.Args = make([]any, len(call.ArgumentList))
					for i, arg := range call.ArgumentList {
						val, err := executeExpression(arg.Expr, ctx)
						if err != nil {
							return nil, err
						}
						access.Args[i] = val
					}
				}

				chain = append([]PropAccessor{access}, chain...)
				current = member.Object.Expr
			} else {
				rootExpr = current
				break
			}
		}

		// 获取根对象
		obj, err := executeExpression(rootExpr, ctx)
		if err != nil {
			return nil, err
		}

		// 应用转换器
		return ctx.PropAccessTransformer(chain, obj)

	case *ast.CallExpression:
		callee, err := executeExpression(e.Callee.Expr, ctx)
		if err != nil {
			return nil, err
		}

		args := make([]reflect.Value, len(e.ArgumentList))
		for i, arg := range e.ArgumentList {
			val, err := executeExpression(arg.Expr, ctx)
			if err != nil {
				return nil, err
			}
			args[i] = reflect.ValueOf(val)
		}

		// 调用函数
		fn := reflect.ValueOf(callee)
		if !fn.IsValid() || fn.Kind() != reflect.Func {
			return nil, fmt.Errorf("不是可调用的函数: %v", callee)
		}

		results := fn.Call(args)
		if len(results) == 0 {
			return nil, nil
		}
		// 对于多返回值的函数，只返回第一个值
		return results[0].Interface(), nil

	case *ast.BinaryExpression:
		left, err := executeExpression(e.Left.Expr, ctx)
		if err != nil {
			return nil, err
		}
		right, err := executeExpression(e.Right.Expr, ctx)
		if err != nil {
			return nil, err
		}

		// 执行运算
		switch e.Operator {
		case token.Plus:
			// 字符串拼接
			if ls, ok := left.(string); ok {
				if rs, ok := right.(string); ok {
					return ls + rs, nil
				}
			}
			// 数字相加
			return toFloat64(left) + toFloat64(right), nil
		case token.Minus:
			return toFloat64(left) - toFloat64(right), nil
		case token.Multiply:
			return toFloat64(left) * toFloat64(right), nil
		case token.Slash:
			return toFloat64(left) / toFloat64(right), nil
		case token.Remainder:
			return float64(int64(toFloat64(left)) % int64(toFloat64(right))), nil
		default:
			return nil, fmt.Errorf("暂不支持的运算符: %v", e.Operator)
		}

	default:
		return nil, fmt.Errorf("暂不支持的表达式类型: %T", expr)
	}
}

// toFloat64 将值转换为float64
func toFloat64(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	default:
		return 0
	}
}

// callMethod 通过反射调用方法
func callMethod(obj any, methodName string) (any, error) {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	method := val.MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("方法不存在: %s", methodName)
	}
	results := method.Call(nil)
	if len(results) == 0 {
		return nil, nil
	}
	return results[0].Interface(), nil
}

// defaultAccess 默认的属性访问逻辑
func defaultAccess(obj any, prop string) (any, error) {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		field := val.FieldByName(prop)
		if field.IsValid() {
			return field.Interface(), nil
		}
	} else if val.Kind() == reflect.Map {
		mapVal := val.MapIndex(reflect.ValueOf(prop))
		if mapVal.IsValid() {
			return mapVal.Interface(), nil
		}
	}

	return nil, fmt.Errorf("属性不存在: %s", prop)
}
