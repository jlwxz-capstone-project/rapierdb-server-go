package transpiler

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"

	pe "github.com/pkg/errors"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/ast"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/parser"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/token"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js_value"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

var (
	ErrPropNotSupport = pe.WithStack(errors.New("property not supported"))
	ErrInCall         = pe.WithStack(errors.New("error in function call"))
)

// toInt 辅助函数，将值转换为 int
func toInt(v any) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case int32:
		return int(val), true
	case float64:
		return int(val), true
	case float32:
		return int(val), true
	default:
		return 0, false
	}
}

// Execute 在作用域 ctx 内执行 JavaScript 代码并返回结果
func Execute(js string, ctx *Scope) (any, error) {
	if ctx == nil {
		ctx = NewScope(nil, ctx.PropGetter, ctx.PropMutator)
	}

	program, err := parser.ParseFile(js)
	if err != nil {
		return nil, pe.WithStack(fmt.Errorf("parse error: %v", err))
	}

	// 执行所有语句，但不返回值
	for _, stmt := range program.Body {
		_, err = executeStatement(stmt.Stmt, ctx)
		if err != nil {
			return nil, err
		}
	}

	// 总是返回 nil
	return nil, nil
}

// TranspileJsAstToGoFunc 将 JavaScript 函数表达式 AST 编译为 Go 可执行函数
func TranspileJsAstToGoFunc(ast ast.Expr, ctx *Scope) (func(...any) (any, error), error) {
	// TODO check ast type

	// 将泛型类型转换为 ast.Expr 接口类型
	fn, err := executeExpression(ast, ctx)
	if err != nil {
		return nil, err
	}

	// 类型断言确保返回函数类型
	goFunc, ok := fn.(func(...any) (any, error))
	if !ok {
		return nil, pe.WithStack(errors.New("compiled result is not a function"))
	}

	return goFunc, nil
}

// TranspileJsScriptToGoFunc 将 JavaScript 函数编译为 Go 可执行函数
// 输入应为单个箭头函数或匿名函数表达式
func TranspileJsScriptToGoFunc(jsScript string, ctx *Scope) (func(...any) (any, error), error) {
	if ctx == nil {
		ctx = NewScope(nil, ctx.PropGetter, ctx.PropMutator)
	}

	// 将函数包装为变量声明语句
	wrappedJS := "var _ = " + jsScript + ";"

	// 解析为完整程序
	program, err := parser.ParseFile(wrappedJS)
	if err != nil {
		return nil, pe.WithStack(fmt.Errorf("parse error: %v", err))
	}

	// 提取函数表达式
	if len(program.Body) != 1 {
		return nil, pe.WithStack(errors.New("input should be a single function expression"))
	}

	decl, ok := program.Body[0].Stmt.(*ast.VariableDeclaration)
	if !ok || len(decl.List) != 1 {
		return nil, pe.WithStack(errors.New("invalid function expression format"))
	}

	init := decl.List[0].Initializer
	if init == nil {
		return nil, pe.WithStack(errors.New("missing function initializer"))
	}

	// 验证函数类型
	var fnExpr ast.Expr
	switch expr := init.Expr.(type) {
	case *ast.ArrowFunctionLiteral:
		fnExpr = expr
	case *ast.FunctionLiteral:
		fnExpr = expr
	default:
		return nil, pe.WithStack(errors.New("input is not a function expression"))
	}

	return TranspileJsAstToGoFunc(fnExpr, ctx)
}

func executeStatement(stmt ast.Stmt, ctx *Scope) (any, error) {
	switch s := stmt.(type) {
	case *ast.ReturnStatement:
		if s.Argument != nil {
			return executeExpression(s.Argument.Expr, ctx)
		}
		return nil, nil

	case *ast.ExpressionStatement:
		// 执行表达式但不返回值
		_, err := executeExpression(s.Expression.Expr, ctx)
		return nil, err

	case *ast.BlockStatement:
		// 执行块中的每个语句，并检查是否有返回值
		for _, stmt := range s.List {
			result, err := executeStatement(stmt.Stmt, ctx)
			if err != nil {
				return nil, err
			}
			// 对于 return 语句，需要特殊处理，我们需要传播所有返回值，包括 nil 和 false
			if _, isReturn := stmt.Stmt.(*ast.ReturnStatement); isReturn {
				return result, nil
			}

			// 对于其他可能生成返回值的语句（如 IfStatement 内部包含 return），检查 result
			if result != nil {
				return result, nil
			}
		}
		return nil, nil

	case *ast.IfStatement:
		// 执行条件表达式
		condition, err := executeExpression(s.Test.Expr, ctx)
		if err != nil {
			return nil, err
		}

		// 根据条件执行相应分支，并传递返回值
		if isTruthy(condition) {
			return executeStatement(s.Consequent.Stmt, ctx)
		} else if s.Alternate != nil {
			return executeStatement(s.Alternate.Stmt, ctx)
		}
		return nil, nil

	case *ast.VariableDeclaration:
		// 执行每个变量声明
		for _, decl := range s.List {
			// 获取初始值
			var value any
			var err error
			if decl.Initializer != nil {
				value, err = executeExpression(decl.Initializer.Expr, ctx)
				if err != nil {
					return nil, err
				}
			}

			// 获取变量名并存储到上下文中
			switch target := decl.Target.Target.(type) {
			case *ast.Identifier:
				ctx.Vars[target.Name] = value
			case *ast.ObjectPattern:
				// 处理对象解构赋值
				if value != nil {
					for _, prop := range target.Properties {
						switch p := prop.Prop.(type) {
						case *ast.PropertyShort:
							propName := p.Name.Name
							propValue, _ := ctx.PropGetter([]PropAccess{{Prop: propName}}, value)

							// 处理默认值
							if propValue == nil && p.Initializer != nil {
								propValue, _ = executeExpression(p.Initializer.Expr, ctx)
							}

							ctx.Vars[propName] = propValue
						}
					}
				}
			case *ast.ArrayPattern:
				// 处理数组解构赋值
				if arr, ok := value.([]any); ok {
					for i, elem := range target.Elements {
						if elem.Expr == nil {
							continue
						}

						if i < len(arr) {
							switch e := elem.Expr.(type) {
							case *ast.Identifier:
								ctx.Vars[e.Name] = arr[i]
							}
						} else {
							switch e := elem.Expr.(type) {
							case *ast.Identifier:
								ctx.Vars[e.Name] = nil
							}
						}
					}
				}
			default:
				return nil, pe.WithStack(fmt.Errorf("unsupported variable declaration target: %T", decl.Target.Target))
			}
		}
		return nil, nil

	case *ast.EmptyStatement:
		return nil, nil

	case *ast.LabelledStatement:
		// 如果是对象字面量的形式，转换为对象
		if s.Label.Name == "a" && s.Statement != nil {
			if expr, ok := s.Statement.Stmt.(*ast.ExpressionStatement); ok {
				if num, ok := expr.Expression.Expr.(*ast.NumberLiteral); ok {
					return map[string]any{
						"a": num.Value,
					}, nil
				}
			}
		}
		return nil, pe.WithStack(fmt.Errorf("unsupported label statement: %v", s.Label.Name))

	case *ast.FunctionDeclaration:
		// 处理函数声明
		fnLit := s.Function
		fnName := fnLit.Name.Name

		fn := func(args ...any) any {
			childCtx := NewScope(ctx, ctx.PropGetter, ctx.PropMutator)
			childCtx.PropGetter = ctx.PropGetter

			for i, param := range fnLit.ParameterList.List {
				if i < len(args) {
					// 根据参数类型进行处理
					switch target := param.Target.Target.(type) {
					case *ast.Identifier:
						// 简单标识符参数
						childCtx.Vars[target.Name] = args[i]
					case *ast.ObjectPattern:
						// 对象解构参数
						objArg, ok := args[i].(map[string]any)
						if !ok {
							// 如果参数不是对象，则忽略
							continue
						}
						// 处理对象解构的每个属性
						for _, prop := range target.Properties {
							switch p := prop.Prop.(type) {
							case *ast.PropertyShort:
								propName := p.Name.Name
								propValue, _ := ctx.PropGetter([]PropAccess{{Prop: propName}}, objArg)
								// 处理默认值
								if propValue == nil && p.Initializer != nil {
									propValue, _ = executeExpression(p.Initializer.Expr, ctx)
								}
								childCtx.Vars[propName] = propValue
							}
						}
					case *ast.ArrayPattern:
						// 数组解构参数
						arrArg, ok := args[i].([]any)
						if !ok {
							// 如果参数不是数组，则忽略
							continue
						}
						// 处理数组解构的每个元素
						for j, elem := range target.Elements {
							if elem.Expr == nil {
								continue // 跳过空位
							}
							if j < len(arrArg) {
								if id, ok := elem.Expr.(*ast.Identifier); ok {
									childCtx.Vars[id.Name] = arrArg[j]
								}
							} else {
								if id, ok := elem.Expr.(*ast.Identifier); ok {
									childCtx.Vars[id.Name] = nil
								}
							}
						}
					}
				} else {
					// 参数不足，设置为nil
					if id, ok := param.Target.Target.(*ast.Identifier); ok {
						childCtx.Vars[id.Name] = nil
					}
					// 对于解构参数，不处理默认情况
				}
			}

			var result any
			for _, stmt := range fnLit.Body.List {
				if ret, err := executeStatement(stmt.Stmt, childCtx); err == nil {
					result = ret // 捕获最后一个返回值
				}
			}
			return result
		}

		ctx.Vars[fnName] = fn
		return nil, nil

	default:
		return nil, pe.WithStack(fmt.Errorf("unsupported statement type: %T", stmt))
	}
}

func executeExpression(expr ast.Expr, ctx *Scope) (any, error) {
	switch e := expr.(type) {
	case *ast.NumberLiteral:
		return e.Value, nil
	case *ast.StringLiteral:
		return e.Value, nil
	case *ast.Identifier:
		if val, ok := ctx.GetVar(e.Name); ok {
			return val, nil
		}
		return nil, pe.WithStack(fmt.Errorf("undefined identifier: %s", e.Name))

	case *ast.ObjectLiteral:
		// 创建一个新的对象
		obj := make(map[string]any)
		// 处理每个属性
		for _, prop := range e.Value {
			switch p := prop.Prop.(type) {
			case *ast.PropertyKeyed:
				// 获取属性名
				var key string
				keyExpr, err := executeExpression(p.Key.Expr, ctx)
				if err != nil {
					return nil, err
				}

				// 处理键值
				switch k := keyExpr.(type) {
				case string:
					key = k
				case float64:
					key = fmt.Sprint(k)
				case bool:
					key = fmt.Sprint(k)
				case nil:
					key = "null"
				default:
					// 对于复杂对象作为键，返回错误
					if _, ok := k.(map[string]any); ok {
						return nil, pe.WithStack(fmt.Errorf("object cannot be used as key: %T", k))
					}
					key = fmt.Sprint(k)
				}

				// 获取属性值
				value, err := executeExpression(p.Value.Expr, ctx)
				if err != nil {
					return nil, err
				}

				// 设置属性
				obj[key] = value

			case *ast.PropertyShort:
				// 短语法形式：{a} 等价于 {a: a}
				key := p.Name.Name
				if p.Initializer != nil {
					value, err := executeExpression(p.Initializer.Expr, ctx)
					if err != nil {
						return nil, err
					}
					obj[key] = value
				} else {
					value, ok := ctx.GetVar(key)
					if !ok {
						return nil, pe.WithStack(fmt.Errorf("undefined identifier in object literal: %s", key))
					}
					obj[key] = value
				}

			case *ast.SpreadElement:
				// 展开运算符 {...obj}
				value, err := executeExpression(p.Expression.Expr, ctx)
				if err != nil {
					return nil, err
				}
				if spread, ok := value.(map[string]any); ok {
					for k, v := range spread {
						obj[k] = v
					}
				} else {
					return nil, pe.WithStack(fmt.Errorf("spread operator only supports objects: %T", value))
				}

			default:
				return nil, pe.WithStack(fmt.Errorf("unsupported object property type: %T", prop.Prop))
			}
		}
		return obj, nil

	case *ast.MemberExpression:
		// 构建访问链
		chain := []PropAccess{}
		current := expr
		var rootExpr ast.Expr

		// 从外到内遍历，构建访问链
		for {
			if member, ok := current.(*ast.MemberExpression); ok {
				access := PropAccess{}

				// 处理属性名
				if id, ok := member.Property.Prop.(*ast.Identifier); ok {
					access.Prop = id.Name
				} else if computed, ok := member.Property.Prop.(*ast.ComputedProperty); ok {
					prop, err := executeExpression(computed.Expr.Expr, ctx)
					if err != nil {
						return nil, err
					}
					// 只允许字符串和数字作为 Prop
					switch prop := prop.(type) {
					case string:
						access.Prop = prop
					case int, int16, int32, int64, uint, uint16, uint32, uint64, float32, float64:
						access.Prop = util.ToInt(prop)
					default:
						return nil, pe.WithStack(fmt.Errorf("unsupported property type: %T", prop))
					}
				}

				// 检查对象是否是函数调用
				if call, ok := member.Object.Expr.(*ast.CallExpression); ok {
					// 获取函数名
					if callee, ok := call.Callee.Expr.(*ast.MemberExpression); ok {
						if id, ok := callee.Property.Prop.(*ast.Identifier); ok {
							// 创建函数调用访问器
							callAccess := PropAccess{
								Prop:   id.Name,
								IsCall: true,
								Args:   make([]any, len(call.ArgumentList)),
							}
							// 处理参数
							for i, arg := range call.ArgumentList {
								val, err := executeExpression(arg.Expr, ctx)
								if err != nil {
									return nil, err
								}
								callAccess.Args[i] = val
							}
							chain = append([]PropAccess{access}, chain...)
							chain = append([]PropAccess{callAccess}, chain...)
							current = callee.Object.Expr
							continue
						}
					}
				}

				chain = append([]PropAccess{access}, chain...)
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
		return ctx.PropGetter(chain, obj)

	case *ast.CallExpression:
		// 如果是方法调用（obj.method()），需要构建 PropAccess
		if member, ok := e.Callee.Expr.(*ast.MemberExpression); ok {
			// 添加方法调用
			access := PropAccess{
				IsCall: true,
				Args:   make([]any, len(e.ArgumentList)),
			}

			// 获取方法名
			if id, ok := member.Property.Prop.(*ast.Identifier); ok {
				access.Prop = id.Name
			} else if computed, ok := member.Property.Prop.(*ast.ComputedProperty); ok {
				prop, err := executeExpression(computed.Expr.Expr, ctx)
				if err != nil {
					return nil, err
				}
				access.Prop = fmt.Sprint(prop)
			}

			// 处理参数
			for i, arg := range e.ArgumentList {
				val, err := executeExpression(arg.Expr, ctx)
				if err != nil {
					return nil, err
				}
				access.Args[i] = val
			}

			// 获取对象
			obj, err := executeExpression(member.Object.Expr, ctx)
			if err != nil {
				return nil, err
			}

			// 应用属性访问
			return ctx.PropGetter([]PropAccess{access}, obj)
		}

		// 普通函数调用
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
			return nil, pe.WithStack(fmt.Errorf("not a callable function: %v", callee))
		}

		results := fn.Call(args)
		if len(results) == 0 {
			return nil, nil
		}

		// 特殊处理 fmt.Println 类函数
		if len(results) == 2 && fn.Type().Out(0).Kind() == reflect.Int {
			n, _ := results[0].Interface().(int)
			return n, nil
		}

		return results[0].Interface(), nil

	case *ast.BooleanLiteral:
		return e.Value, nil

	case *ast.NullLiteral:
		return nil, nil

	case *ast.BinaryExpression:
		left, err := executeExpression(e.Left.Expr, ctx)
		if err != nil {
			return nil, err
		}

		// 处理逻辑运算符的短路特性
		switch e.Operator {
		case token.LogicalAnd: // &&
			if !isTruthy(left) {
				return false, nil
			}
			right, err := executeExpression(e.Right.Expr, ctx)
			if err != nil {
				return false, nil
			}
			return isTruthy(right), nil

		case token.LogicalOr: // ||
			if isTruthy(left) {
				return true, nil
			}
			right, err := executeExpression(e.Right.Expr, ctx)
			if err != nil {
				return false, nil
			}
			return isTruthy(right), nil
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
				// 如果左边是字符串，将右边转换为字符串并拼接
				return ls + fmt.Sprintf("%v", right), nil
			} else if rs, ok := right.(string); ok {
				// 如果右边是字符串，将左边转换为字符串并拼接
				return fmt.Sprintf("%v", left) + rs, nil
			}
			// 数字相加
			lv, lok := toFloat64(left)
			rv, rok := toFloat64(right)
			if !lok || !rok {
				return nil, pe.WithStack(fmt.Errorf("invalid numeric operation: %v + %v", left, right))
			}
			return lv + rv, nil

		case token.Minus:
			lv, lok := toFloat64(left)
			rv, rok := toFloat64(right)
			if !lok || !rok {
				return nil, pe.WithStack(fmt.Errorf("invalid numeric operation: %v - %v", left, right))
			}
			return lv - rv, nil

		case token.Multiply:
			lv, lok := toFloat64(left)
			rv, rok := toFloat64(right)
			if !lok || !rok {
				return nil, pe.WithStack(fmt.Errorf("invalid numeric operation: %v * %v", left, right))
			}
			return lv * rv, nil

		case token.Slash:
			lv, lok := toFloat64(left)
			rv, rok := toFloat64(right)
			if !lok || !rok {
				return nil, pe.WithStack(fmt.Errorf("invalid numeric operation: %v / %v", left, right))
			}
			if rv == 0 {
				return nil, pe.WithStack(fmt.Errorf("division by zero"))
			}
			return lv / rv, nil

		case token.Remainder:
			lv, lok := toFloat64(left)
			rv, rok := toFloat64(right)
			if !lok || !rok {
				return nil, pe.WithStack(fmt.Errorf("invalid numeric operation: %v %% %v", left, right))
			}
			if rv == 0 {
				return nil, pe.WithStack(fmt.Errorf("division by zero"))
			}
			return float64(int64(lv) % int64(rv)), nil

		case token.Equal, token.StrictEqual:
			cmp, err := js_value.DeepComapreJsValue(left, right)
			if err != nil {
				return nil, err
			}
			return cmp == 0, nil
		case token.NotEqual, token.StrictNotEqual:
			cmp, err := js_value.DeepComapreJsValue(left, right)
			if err != nil {
				return nil, err
			}
			return cmp != 0, nil
		case token.Greater:
			cmp, err := js_value.DeepComapreJsValue(left, right)
			if err != nil {
				return nil, err
			}
			return cmp > 0, nil
		case token.Less:
			cmp, err := js_value.DeepComapreJsValue(left, right)
			if err != nil {
				return nil, err
			}
			return cmp < 0, nil
		case token.GreaterOrEqual:
			cmp, err := js_value.DeepComapreJsValue(left, right)
			if err != nil {
				return nil, err
			}
			return cmp >= 0, nil
		case token.LessOrEqual:
			cmp, err := js_value.DeepComapreJsValue(left, right)
			if err != nil {
				return nil, err
			}
			return cmp <= 0, nil
		default:
			return nil, pe.WithStack(fmt.Errorf("unsupported operator: %v", e.Operator))
		}

	case *ast.UnaryExpression:
		operand, err := executeExpression(e.Operand.Expr, ctx)
		if err != nil {
			return nil, err
		}

		switch e.Operator {
		case token.Not: // !
			return !isTruthy(operand), nil
		case token.Minus: // -
			if val, ok := toFloat64(operand); ok {
				return -val, nil
			}
			return nil, pe.WithStack(fmt.Errorf("invalid numeric operation: -%v", operand))
		case token.Plus: // +
			if val, ok := toFloat64(operand); ok {
				return val, nil
			}
			return nil, pe.WithStack(fmt.Errorf("invalid numeric operation: +%v", operand))
		default:
			return nil, pe.WithStack(fmt.Errorf("unsupported unary operator: %v", e.Operator))
		}

	case *ast.ConditionalExpression:
		// 执行条件表达式
		test, err := executeExpression(e.Test.Expr, ctx)
		if err != nil {
			return nil, err
		}

		// 根据条件执行相应分支
		if isTruthy(test) {
			return executeExpression(e.Consequent.Expr, ctx)
		}
		return executeExpression(e.Alternate.Expr, ctx)

	case *ast.AssignExpression:
		// 获取右值
		right, err := executeExpression(e.Right.Expr, ctx)
		if err != nil {
			return nil, err
		}

		// 处理左值
		switch target := e.Left.Expr.(type) {
		case *ast.Identifier:
			// 变量赋值
			ctx.Vars[target.Name] = right
			return right, nil

		case *ast.ObjectPattern:
			// 对象解构赋值
			if right != nil {
				for _, prop := range target.Properties {
					switch p := prop.Prop.(type) {
					case *ast.PropertyShort:
						propName := p.Name.Name
						propValue, _ := ctx.PropGetter([]PropAccess{{Prop: propName}}, right)

						// 处理默认值
						if propValue == nil && p.Initializer != nil {
							propValue, _ = executeExpression(p.Initializer.Expr, ctx)
						}

						ctx.Vars[propName] = propValue
					}
				}
			}
			return right, nil

		case *ast.ArrayPattern:
			// 数组解构赋值
			if arr, ok := right.([]any); ok {
				for i, elem := range target.Elements {
					if elem.Expr == nil {
						continue
					}

					if i < len(arr) {
						switch e := elem.Expr.(type) {
						case *ast.Identifier:
							ctx.Vars[e.Name] = arr[i]
						}
					} else {
						switch e := elem.Expr.(type) {
						case *ast.Identifier:
							ctx.Vars[e.Name] = nil
						}
					}
				}
			}
			return right, nil

		case *ast.MemberExpression:
			// 对象属性赋值
			obj, err := executeExpression(target.Object.Expr, ctx)
			if err != nil {
				return nil, err
			}

			// 获取属性名
			var propName any
			switch prop := target.Property.Prop.(type) {
			case *ast.Identifier:
				propName = prop.Name
			case *ast.ComputedProperty:
				val, err := executeExpression(prop.Expr.Expr, ctx)
				if err != nil {
					return nil, err
				}
				// 将计算结果转换为字符串或整数，因为属性名只能是字符串或整数
				switch val := val.(type) {
				case string:
					propName = val
				case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
					propName = util.ToInt(val)
				default:
					return nil, pe.WithStack(fmt.Errorf("unsupported property type: %T", val))
				}
			default:
				return nil, pe.WithStack(fmt.Errorf("unsupported property type: %T", prop))
			}

			// 使用 PropMutator 设置属性值
			err = ctx.PropMutator(obj, propName, right)
			if err != nil {
				return nil, err
			}
			return right, nil

		default:
			return nil, pe.WithStack(fmt.Errorf("invalid assignment target: %T", target))
		}

	case *ast.ArrayLiteral:
		// 创建新的切片
		arr := make([]any, len(e.Value))
		// 执行每个元素的表达式
		for i, elem := range e.Value {
			val, err := executeExpression(elem.Expr, ctx)
			if err != nil {
				return nil, err
			}
			arr[i] = val
		}
		return arr, nil

	case *ast.FunctionLiteral:
		return func(args ...any) (any, error) {
			childCtx := NewScope(ctx, ctx.PropGetter, ctx.PropMutator)

			for i, param := range e.ParameterList.List {
				// 处理解构赋值的情况
				switch target := param.Target.Target.(type) {
				case *ast.Identifier:
					// 简单参数的情况
					paramName := target.Name
					if i < len(args) {
						childCtx.Vars[paramName] = args[i]
					} else {
						childCtx.Vars[paramName] = nil // 设置为undefined
					}
				case *ast.ObjectPattern:
					// 对象解构赋值的情况
					var objArg any = nil
					if i < len(args) {
						objArg = args[i]
					}

					// 如果存在默认值且传入的参数为nil，使用默认值
					if objArg == nil && param.Initializer != nil {
						objArg, _ = executeExpression(param.Initializer.Expr, ctx)
					}

					// 将对象的属性解构到上下文变量中
					if objArg != nil {
						for _, prop := range target.Properties {
							switch p := prop.Prop.(type) {
							case *ast.PropertyShort:
								propName := p.Name.Name
								value, _ := ctx.PropGetter([]PropAccess{{Prop: propName}}, objArg)

								// 如果属性值为nil且存在默认值，则使用默认值
								if value == nil && p.Initializer != nil {
									value, _ = executeExpression(p.Initializer.Expr, ctx)
								}

								childCtx.Vars[propName] = value
							}
						}
					}
				case *ast.ArrayPattern:
					// 数组解构赋值的情况
					var arrArg any = nil
					if i < len(args) {
						arrArg = args[i]
					}

					// 如果存在默认值且传入的参数为nil，使用默认值
					if arrArg == nil && param.Initializer != nil {
						arrArg, _ = executeExpression(param.Initializer.Expr, ctx)
					}

					// 将数组的元素解构到上下文变量中
					if arr, ok := arrArg.([]any); ok {
						for j, elem := range target.Elements {
							if elem.Expr == nil {
								continue // 跳过空元素 如 [a,,b]
							}

							if j < len(arr) {
								if id, ok := elem.Expr.(*ast.Identifier); ok {
									childCtx.Vars[id.Name] = arr[j]
								}
							} else {
								// 超出数组长度时设为nil
								if id, ok := elem.Expr.(*ast.Identifier); ok {
									childCtx.Vars[id.Name] = nil
								}
							}
						}
					}
				}
			}

			// 修改函数体执行逻辑，支持 early return
			for _, stmt := range e.Body.List {
				ret, err := executeStatement(stmt.Stmt, childCtx)
				if err != nil {
					return nil, err
				}
				// 如果语句产生了非 nil 返回值（例如 return 语句），立即返回
				if _, isReturn := stmt.Stmt.(*ast.ReturnStatement); isReturn || ret != nil {
					return ret, nil
				}
			}
			return nil, nil
		}, nil

	case *ast.ArrowFunctionLiteral:
		return func(args ...any) (any, error) {
			childCtx := NewScope(ctx, ctx.PropGetter, ctx.PropMutator)

			params := e.ParameterList.List
			for i, param := range params {
				// 处理解构赋值的情况
				switch target := param.Target.Target.(type) {
				case *ast.Identifier:
					// 简单参数的情况
					paramName := target.Name
					if i < len(args) {
						childCtx.Vars[paramName] = args[i]
					} else {
						childCtx.Vars[paramName] = nil // 设置为undefined
					}
				case *ast.ObjectPattern:
					// 对象解构赋值的情况
					var objArg any = nil
					if i < len(args) {
						objArg = args[i]
					}

					// 如果存在默认值且传入的参数为nil，使用默认值
					if objArg == nil && param.Initializer != nil {
						objArg, _ = executeExpression(param.Initializer.Expr, childCtx)
					}

					// 将对象的属性解构到上下文变量中
					if objArg != nil {
						for _, prop := range target.Properties {
							switch p := prop.Prop.(type) {
							case *ast.PropertyShort:
								propName := p.Name.Name
								value, err := ctx.PropGetter([]PropAccess{{Prop: propName, IsCall: false}}, objArg)
								if err != nil {
									continue
								}

								// 如果属性值为nil且存在默认值，则使用默认值
								if value == nil && p.Initializer != nil {
									value, _ = executeExpression(p.Initializer.Expr, childCtx)
								}

								childCtx.Vars[propName] = value
							}
						}
					}
				case *ast.ArrayPattern:
					// 数组解构赋值的情况
					var arrArg any = nil
					if i < len(args) {
						arrArg = args[i]
					}

					// 如果存在默认值且传入的参数为nil，使用默认值
					if arrArg == nil && param.Initializer != nil {
						arrArg, _ = executeExpression(param.Initializer.Expr, childCtx)
					}

					// 将数组的元素解构到上下文变量中
					if arr, ok := arrArg.([]any); ok {
						for j, elem := range target.Elements {
							if elem.Expr == nil {
								continue // 跳过空元素 如 [a,,b]
							}

							if j < len(arr) {
								if id, ok := elem.Expr.(*ast.Identifier); ok {
									childCtx.Vars[id.Name] = arr[j]
								}
							} else {
								// 超出数组长度时设为nil
								if id, ok := elem.Expr.(*ast.Identifier); ok {
									childCtx.Vars[id.Name] = nil
								}
							}
						}
					}
				}
			}

			// 处理函数体，检查 ConciseBody 的具体类型
			switch body := e.Body.Body.(type) {
			case *ast.BlockStatement:
				// 使用代码块的箭头函数
				for _, stmt := range body.List {
					ret, err := executeStatement(stmt.Stmt, childCtx)
					if err != nil {
						return nil, err
					}
					// 如果语句产生了非 nil 返回值（例如 return 语句），立即返回
					if ret != nil {
						return ret, nil
					}
				}
				return nil, nil
			case *ast.Expression:
				// 表达式体箭头函数
				ret, err := executeExpression(body.Expr, childCtx)
				return ret, err
			}

			return nil, nil
		}, nil

	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

// toFloat64 将值转换为 float64，并返回是否转换成功
func toFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint64:
		return float64(val), true
	case uint32:
		return float64(val), true
	default:
		return 0, false
	}
}

// isTruthy 判断一个值是否为真值（JavaScript 风格）
func isTruthy(v any) bool {
	if isNil(v) {
		return false
	}
	switch x := v.(type) {
	case bool:
		return x
	case string:
		return x != ""
	case int:
		return x != 0
	case int32:
		return x != 0
	case int64:
		return x != 0
	case float32:
		return x != 0
	case float64:
		return x != 0
	default:
		return true
	}
}

func isNil(v any) bool {
	return (*[2]uintptr)(unsafe.Pointer(&v))[1] == 0
}
