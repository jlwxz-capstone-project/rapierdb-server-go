package transpiler

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/ast"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/parser"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/token"
)

// Context 定义转译上下文
type Context struct {
	// 变量和函数映射表
	Vars map[string]interface{}
	// 属性访问转译规则
	PropAccessRules []PropAccessRule
}

// PropAccessRule 定义属性访问的转译规则
type PropAccessRule struct {
	// 匹配对象类型
	ObjectType reflect.Type
	// 属性访问转译函数
	Transform func(obj interface{}, prop string) (interface{}, error)
}

// defaultPropAccessRule 返回默认的属性访问规则
func defaultPropAccessRule() PropAccessRule {
	return PropAccessRule{
		ObjectType: reflect.TypeOf(struct{}{}),
		Transform: func(obj interface{}, prop string) (interface{}, error) {
			val := reflect.ValueOf(obj)
			if val.Kind() == reflect.Ptr {
				val = val.Elem()
			}

			// 先尝试原始属性名
			field := val.FieldByName(prop)
			if field.IsValid() {
				return field.Interface(), nil
			}

			// 尝试首字母大写的属性名
			field = val.FieldByName(strings.Title(prop))
			if field.IsValid() {
				return field.Interface(), nil
			}

			// 尝试方法调用
			method := val.MethodByName(prop)
			if method.IsValid() {
				return method.Interface(), nil
			}
			method = val.MethodByName(strings.Title(prop))
			if method.IsValid() {
				return method.Interface(), nil
			}

			return nil, fmt.Errorf("属性不存在: %s", prop)
		},
	}
}

// NewContext 创建新的转译上下文
func NewContext() *Context {
	ctx := &Context{
		Vars:            make(map[string]interface{}),
		PropAccessRules: make([]PropAccessRule, 0),
	}
	// 添加默认规则
	ctx.AddPropAccessRule(defaultPropAccessRule().ObjectType, defaultPropAccessRule().Transform)
	return ctx
}

// AddPropAccessRule 添加属性访问转译规则
func (c *Context) AddPropAccessRule(objType reflect.Type, transform func(obj interface{}, prop string) (interface{}, error)) {
	c.PropAccessRules = append(c.PropAccessRules, PropAccessRule{
		ObjectType: objType,
		Transform:  transform,
	})
}

// Execute 执行 JavaScript 代码并返回结果
func Execute(js string, ctx *Context) (interface{}, error) {
	if ctx == nil {
		ctx = NewContext()
	}

	program, err := parser.ParseFile(js)
	if err != nil {
		return nil, fmt.Errorf("解析错误: %v", err)
	}

	var result interface{}
	for _, stmt := range program.Body {
		result, err = executeStatement(stmt.Stmt, ctx)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func executeStatement(stmt ast.Stmt, ctx *Context) (interface{}, error) {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		return executeExpression(s.Expression.Expr, ctx)
	default:
		return nil, fmt.Errorf("暂不支持的语句类型: %T", stmt)
	}
}

func executeExpression(expr ast.Expr, ctx *Context) (interface{}, error) {
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
		obj, err := executeExpression(e.Object.Expr, ctx)
		if err != nil {
			return nil, err
		}

		var propName string
		switch p := e.Property.Prop.(type) {
		case *ast.Identifier:
			propName = p.Name
		case *ast.ComputedProperty:
			prop, err := executeExpression(p.Expr.Expr, ctx)
			if err != nil {
				return nil, err
			}
			propName = fmt.Sprint(prop)
		default:
			return nil, fmt.Errorf("不支持的属性访问类型: %T", p)
		}

		// 应用属性访问规则
		objType := reflect.TypeOf(obj)
		for _, rule := range ctx.PropAccessRules {
			if objType == rule.ObjectType ||
				(objType.Kind() == reflect.Ptr && objType.Elem() == rule.ObjectType) {
				return rule.Transform(obj, propName)
			}
		}

		// 默认的反射访问逻辑保持不变
		val := reflect.ValueOf(obj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if val.Kind() == reflect.Struct {
			field := val.FieldByName(propName)
			if field.IsValid() {
				return field.Interface(), nil
			}
			// 尝试调用方法
			method := val.MethodByName(propName)
			if method.IsValid() {
				return method.Interface(), nil
			}
		} else if val.Kind() == reflect.Map {
			mapVal := val.MapIndex(reflect.ValueOf(propName))
			if mapVal.IsValid() {
				return mapVal.Interface(), nil
			}
		}

		return nil, fmt.Errorf("属性不存在: %s", propName)

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
func toFloat64(v interface{}) float64 {
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
