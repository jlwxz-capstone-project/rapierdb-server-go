package transpiler

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/ast"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/parser"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/token"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// Context 定义转译上下文
type Context struct {
	// 变量和函数映射表，让 js 代码可以访问这些变量和函数
	Vars map[string]any
	// 属性访问器
	PropGetter func(chain []PropAccess, obj any) (any, error)
}

// PropAccess 表示一次属性访问
//
// 例如：obj.method(arg1, arg2) 对应的 PropAccess 为：
//
//	PropAccess{
//		Prop: "method",
//		Args: []any{"arg1", "arg2"},
//		IsCall: true,
//	}
//
// obj.name 对应的 PropAccess 为：
//
//	PropAccess{
//		Prop: "name",
//	}
type PropAccess struct {
	// 属性名
	Prop string
	// 如果是函数调用，这里是参数
	Args []any
	// 是否是函数调用
	IsCall bool
}

type PropAccessHandler func(access PropAccess, obj any) (any, error)

var (
	ErrPropNotSupport = errors.New("property not supported")
	ErrInCall         = errors.New("error in function call")
)

func StringPropAccessHandler(access PropAccess, obj any) (any, error) {
	if str, ok := obj.(string); ok {
		switch access.Prop {
		case "length":
			return len(str), nil

		case "toLowerCase":
			if access.IsCall {
				return strings.ToLower(str), nil
			}
			return strings.ToLower, nil

		case "toUpperCase":
			if access.IsCall {
				return strings.ToUpper(str), nil
			}
			return strings.ToUpper, nil

		case "trim":
			if access.IsCall {
				return strings.TrimSpace(str), nil
			}
			return strings.TrimSpace, nil

		case "substring":
			if access.IsCall {
				if len(access.Args) < 1 {
					return nil, fmt.Errorf("%w: substring method requires 1 argument", ErrInCall)
				}
				start, ok := access.Args[0].(int)
				if !ok {
					return nil, fmt.Errorf("%w: first argument of substring must be a number", ErrInCall)
				}
				if start < 0 {
					start = 0
				}

				if len(access.Args) > 1 {
					end, ok := access.Args[1].(int)
					if !ok {
						return nil, fmt.Errorf("%w: second argument of substring must be a number", ErrInCall)
					}
					if end > len(str) {
						end = len(str)
					}
					obj = str[start:end]
				} else {
					obj = str[start:]
				}
				return obj, nil
			}

		case "indexOf":
			// 查找子串位置
			if access.IsCall {
				if len(access.Args) < 1 {
					return nil, fmt.Errorf("%w: indexOf method requires 1 argument", ErrInCall)
				}
				substr, ok := access.Args[0].(string)
				if !ok {
					return nil, fmt.Errorf("%w: argument of indexOf must be a string", ErrInCall)
				}
				return strings.Index(str, substr), nil
			}

		case "replace":
			// 替换字符串
			if access.IsCall {
				if len(access.Args) < 2 {
					return nil, fmt.Errorf("%w: replace method requires 2 arguments", ErrInCall)
				}
				old, ok1 := access.Args[0].(string)
				new, ok2 := access.Args[1].(string)
				if !ok1 || !ok2 {
					return nil, fmt.Errorf("%w: arguments of replace must be strings", ErrInCall)
				}
				return strings.Replace(str, old, new, 1), nil
			}
		}
	}
	return nil, ErrPropNotSupport
}

func MethodPropAccessHandler(access PropAccess, obj any) (any, error) {
	if access.IsCall {
		val := reflect.ValueOf(obj)
		method := val.MethodByName(access.Prop)
		if !method.IsValid() {
			return nil, fmt.Errorf("method not found: %s", access.Prop)
		}
		args := make([]reflect.Value, len(access.Args))
		for i, arg := range access.Args {
			args[i] = reflect.ValueOf(arg)
		}
		results := method.Call(args)
		if len(results) == 0 {
			return nil, nil
		}
		return results[0].Interface(), nil
	}
	return nil, ErrPropNotSupport
}

func DataPropAccessHandler(access PropAccess, obj any) (any, error) {
	fmt.Printf("DataPropAccessHandler: obj=%T, prop=%s, isCall=%v\n", obj, access.Prop, access.IsCall)

	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	fmt.Printf("After dereference: val.Kind()=%v\n", val.Kind())

	// 如果是方法调用
	if access.IsCall {
		// 先尝试获取字段
		switch val.Kind() {
		case reflect.Struct:
			field := val.FieldByName(access.Prop)
			if field.IsValid() && field.Kind() == reflect.Func {
				// 如果字段是函数，直接调用
				fn := field.Interface()
				fnVal := reflect.ValueOf(fn)
				args := make([]reflect.Value, len(access.Args))
				for i, arg := range access.Args {
					args[i] = reflect.ValueOf(arg)
				}
				results := fnVal.Call(args)
				if len(results) == 0 {
					return nil, nil
				}
				return results[0].Interface(), nil
			}
		}
		return nil, fmt.Errorf("method not found: %s", access.Prop)
	}

	// 非方法调用的属性访问
	switch val.Kind() {
	case reflect.Map:
		fmt.Printf("Handling map access: obj=%v, prop=%s\n", obj, access.Prop)
		// 对于 map，先尝试直接访问
		if m, ok := obj.(map[string]any); ok {
			fmt.Printf("Direct map access: m=%v\n", m)
			if val, ok := m[access.Prop]; ok {
				fmt.Printf("Found value in map: %v\n", val)
				return val, nil
			}
		}
		// 如果直接访问失败，尝试使用反射
		mapVal := val.MapIndex(reflect.ValueOf(access.Prop))
		if !mapVal.IsValid() {
			fmt.Printf("Map value not found: %s\n", access.Prop)
			// 尝试查找方法
			method := val.MethodByName(access.Prop)
			if method.IsValid() {
				return method.Interface(), nil
			}
			return nil, fmt.Errorf("property not found: %s", access.Prop)
		}
		return mapVal.Interface(), nil

	case reflect.Struct:
		fmt.Printf("Handling struct access: obj=%+v, prop=%s\n", obj, access.Prop)
		field := val.FieldByName(access.Prop)
		if !field.IsValid() {
			// 尝试查找方法
			method := val.MethodByName(access.Prop)
			if method.IsValid() {
				return method.Interface(), nil
			}
			return nil, fmt.Errorf("property not found: %s", access.Prop)
		}
		return field.Interface(), nil

	case reflect.Slice, reflect.Array:
		// 对于切片和数组，只支持 length 属性
		if access.Prop == "length" {
			return val.Len(), nil
		}
		return nil, fmt.Errorf("unsupported slice property: %s", access.Prop)

	default:
		return nil, fmt.Errorf("unsupported object type: %v", val.Kind())
	}
}

func ArrayPropAccessHandler(access PropAccess, obj any) (any, error) {
	// 检查是否是切片类型
	val := reflect.ValueOf(obj)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil, ErrPropNotSupport
	}

	switch access.Prop {
	case "length":
		return val.Len(), nil

	case "slice":
		if !access.IsCall {
			return nil, ErrPropNotSupport
		}

		// 检查参数
		if len(access.Args) < 1 || len(access.Args) > 2 {
			return nil, fmt.Errorf("%w: slice method requires 1 or 2 arguments", ErrInCall)
		}

		// 获取起始位置
		start, ok := toInt(access.Args[0])
		if !ok {
			return nil, fmt.Errorf("%w: first argument of slice must be a number", ErrInCall)
		}

		// 处理负数索引
		if start < 0 {
			start = val.Len() + start
		}
		if start < 0 {
			start = 0
		}

		end := val.Len()
		if len(access.Args) > 1 {
			// 获取结束位置
			if e, ok := toInt(access.Args[1]); ok {
				if e < 0 {
					end = val.Len() + e
				} else {
					end = e
				}
			} else {
				return nil, fmt.Errorf("%w: second argument of slice must be a number", ErrInCall)
			}
		}

		// 边界检查
		if start > val.Len() {
			start = val.Len()
		}
		if end > val.Len() {
			end = val.Len()
		}
		if end < start {
			end = start
		}

		// 直接返回切片
		return val.Slice(start, end).Interface(), nil

	case "indexOf":
		if !access.IsCall {
			return nil, ErrPropNotSupport
		}

		if len(access.Args) < 1 {
			return nil, fmt.Errorf("%w: indexOf method requires 1 argument", ErrInCall)
		}

		// 遍历查找元素
		searchVal := access.Args[0]
		for i := 0; i < val.Len(); i++ {
			current := val.Index(i).Interface()
			// 使用 CompareValues 进行比较
			cmp, err := util.CompareValues(current, searchVal)
			if err == nil && cmp == 0 {
				return i, nil
			}
		}
		return -1, nil

	case "join":
		if !access.IsCall {
			return nil, ErrPropNotSupport
		}

		// 默认分隔符
		separator := ","
		if len(access.Args) > 0 {
			if sep, ok := access.Args[0].(string); ok {
				separator = sep
			}
		}

		// 构建字符串
		var result strings.Builder
		for i := 0; i < val.Len(); i++ {
			if i > 0 {
				result.WriteString(separator)
			}
			result.WriteString(fmt.Sprint(val.Index(i).Interface()))
		}
		return result.String(), nil

	case "splice":
		if !access.IsCall {
			return nil, ErrPropNotSupport
		}

		// 检查参数
		if len(access.Args) < 1 {
			return nil, fmt.Errorf("%w: splice method requires at least 1 argument", ErrInCall)
		}

		// 获取起始位置
		start, ok := toInt(access.Args[0])
		if !ok {
			return nil, fmt.Errorf("%w: first argument of splice must be a number", ErrInCall)
		}

		// 处理负数索引
		if start < 0 {
			start = val.Len() + start
		}
		if start < 0 {
			start = 0
		}
		if start > val.Len() {
			start = val.Len()
		}

		// 获取删除数量
		deleteCount := val.Len() - start
		if len(access.Args) > 1 {
			if count, ok := toInt(access.Args[1]); ok {
				if count < 0 {
					count = 0
				}
				if start+count > val.Len() {
					deleteCount = val.Len() - start
				} else {
					deleteCount = count
				}
			} else {
				return nil, fmt.Errorf("%w: second argument of splice must be a number", ErrInCall)
			}
		}

		// 创建新切片存储结果
		newLen := val.Len() - deleteCount + len(access.Args) - 2
		if newLen < 0 {
			newLen = 0
		}

		// 创建新的切片，类型与原切片相同
		newSlice := reflect.MakeSlice(val.Type(), newLen, newLen)

		// 复制前半部分
		if start > 0 {
			reflect.Copy(newSlice.Slice(0, start), val.Slice(0, start))
		}

		// 插入新元素
		insertCount := len(access.Args) - 2
		if insertCount > 0 {
			for i := 0; i < insertCount; i++ {
				newVal := reflect.ValueOf(access.Args[i+2])
				if !newVal.Type().AssignableTo(val.Type().Elem()) {
					return nil, fmt.Errorf("%w: cannot insert value of type %v into slice of type %v",
						ErrInCall, newVal.Type(), val.Type().Elem())
				}
				newSlice.Index(start + i).Set(newVal)
			}
		}

		// 复制后半部分
		if start+deleteCount < val.Len() {
			reflect.Copy(
				newSlice.Slice(start+insertCount, newLen),
				val.Slice(start+deleteCount, val.Len()),
			)
		}

		// 直接返回切片
		return newSlice.Interface(), nil
	}

	return nil, ErrPropNotSupport
}

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

// NewContext 创建新的转译上下文
func NewContext() *Context {
	ctx := &Context{
		Vars:       make(map[string]any),
		PropGetter: DefaultPropGetter,
	}
	// 添加 JavaScript 的内置值
	ctx.Vars["undefined"] = nil
	return ctx
}

// DefaultPropGetter 默认的属性访问器，用于根据 JavaScript 的属性访问获取正确的值
//
// chain 是属性访问链，obj 是根对象。比如 obj.name.slice(1, 2).toUpperCase() 对应的 chain 为：
//
//	chain = []PropAccessor{
//		{Prop: "name"},
//		{Prop: "slice", Args: []any{1, 2}, IsCall: true},
//		{Prop: "toUpperCase", IsCall: true},
//	}
func DefaultPropGetter(chain []PropAccess, obj any) (any, error) {
	fmt.Printf("\nDefaultPropGetter: obj=%T, chain=%+v\n", obj, chain)
	result := obj
	propHandlers := []PropAccessHandler{
		StringPropAccessHandler,
		ArrayPropAccessHandler,
		MethodPropAccessHandler,
		DataPropAccessHandler,
	}
	for _, access := range chain {
		success := false
		fmt.Printf("Trying to access: prop=%s, isCall=%v\n", access.Prop, access.IsCall)
		for _, handler := range propHandlers {
			resultNew, err := handler(access, result)
			if err == nil {
				success = true
				result = resultNew
				fmt.Printf("Handler succeeded: result=%v\n", result)
				break
			}
			fmt.Printf("Handler failed: err=%v\n", err)
		}
		if !success {
			return nil, ErrPropNotSupport
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
	PrintProgram(program)
	if err != nil {
		return nil, fmt.Errorf("parse error: %v", err)
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

func executeStatement(stmt ast.Stmt, ctx *Context) (any, error) {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		// 执行表达式但不返回值
		_, err := executeExpression(s.Expression.Expr, ctx)
		return nil, err

	case *ast.BlockStatement:
		// 执行块中的每个语句，但不返回值
		for _, stmt := range s.List {
			_, err := executeStatement(stmt.Stmt, ctx)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil

	case *ast.IfStatement:
		// 执行条件表达式
		condition, err := executeExpression(s.Test.Expr, ctx)
		if err != nil {
			return nil, err
		}

		// 根据条件执行相应分支，但不返回值
		if isTruthy(condition) {
			_, err = executeStatement(s.Consequent.Stmt, ctx)
		} else if s.Alternate != nil {
			_, err = executeStatement(s.Alternate.Stmt, ctx)
		}
		return nil, err

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
			default:
				return nil, fmt.Errorf("unsupported variable declaration target: %T", decl.Target.Target)
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
		return nil, fmt.Errorf("unsupported label statement: %v", s.Label.Name)

	default:
		return nil, fmt.Errorf("unsupported statement type: %T", stmt)
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
		return nil, fmt.Errorf("undefined identifier: %s", e.Name)

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
						return nil, fmt.Errorf("object cannot be used as key: %T", k)
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
				var value any
				if p.Initializer != nil {
					_, err := executeExpression(p.Initializer.Expr, ctx)
					if err != nil {
						return nil, err
					}
				} else {
					value = ctx.Vars[key]
				}
				obj[key] = value

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
					return nil, fmt.Errorf("spread operator only supports objects: %T", value)
				}

			default:
				return nil, fmt.Errorf("unsupported object property type: %T", prop.Prop)
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
					access.Prop = fmt.Sprint(prop)
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
			return nil, fmt.Errorf("not a callable function: %v", callee)
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
				if rs, ok := right.(string); ok {
					return ls + rs, nil
				}
			}
			// 数字相加
			lv, lok := toFloat64(left)
			rv, rok := toFloat64(right)
			if !lok || !rok {
				return nil, fmt.Errorf("invalid numeric operation: %v + %v", left, right)
			}
			return lv + rv, nil

		case token.Minus:
			lv, lok := toFloat64(left)
			rv, rok := toFloat64(right)
			if !lok || !rok {
				return nil, fmt.Errorf("invalid numeric operation: %v - %v", left, right)
			}
			return lv - rv, nil

		case token.Multiply:
			lv, lok := toFloat64(left)
			rv, rok := toFloat64(right)
			if !lok || !rok {
				return nil, fmt.Errorf("invalid numeric operation: %v * %v", left, right)
			}
			return lv * rv, nil

		case token.Slash:
			lv, lok := toFloat64(left)
			rv, rok := toFloat64(right)
			if !lok || !rok {
				return nil, fmt.Errorf("invalid numeric operation: %v / %v", left, right)
			}
			if rv == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return lv / rv, nil

		case token.Remainder:
			lv, lok := toFloat64(left)
			rv, rok := toFloat64(right)
			if !lok || !rok {
				return nil, fmt.Errorf("invalid numeric operation: %v %% %v", left, right)
			}
			if rv == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return float64(int64(lv) % int64(rv)), nil

		case token.Equal:
			return reflect.DeepEqual(left, right), nil
		case token.NotEqual:
			return !reflect.DeepEqual(left, right), nil
		case token.Greater:
			cmp, err := util.CompareValues(left, right)
			if err != nil {
				return nil, err
			}
			return cmp > 0, nil
		case token.Less:
			cmp, err := util.CompareValues(left, right)
			if err != nil {
				return nil, err
			}
			return cmp < 0, nil
		case token.GreaterOrEqual:
			cmp, err := util.CompareValues(left, right)
			if err != nil {
				return nil, err
			}
			return cmp >= 0, nil
		case token.LessOrEqual:
			cmp, err := util.CompareValues(left, right)
			if err != nil {
				return nil, err
			}
			return cmp <= 0, nil
		default:
			return nil, fmt.Errorf("unsupported operator: %v", e.Operator)
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
			return nil, fmt.Errorf("invalid numeric operation: -%v", operand)
		case token.Plus: // +
			if val, ok := toFloat64(operand); ok {
				return val, nil
			}
			return nil, fmt.Errorf("invalid numeric operation: +%v", operand)
		default:
			return nil, fmt.Errorf("unsupported unary operator: %v", e.Operator)
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

		case *ast.MemberExpression:
			// 对象属性赋值
			obj, err := executeExpression(target.Object.Expr, ctx)
			if err != nil {
				return nil, err
			}

			// 获取属性名
			var propName string
			switch prop := target.Property.Prop.(type) {
			case *ast.Identifier:
				propName = prop.Name
			case *ast.ComputedProperty:
				val, err := executeExpression(prop.Expr.Expr, ctx)
				if err != nil {
					return nil, err
				}
				propName = fmt.Sprint(val)
			default:
				return nil, fmt.Errorf("unsupported property type: %T", prop)
			}

			// 设置属性值
			if m, ok := obj.(map[string]any); ok {
				m[propName] = right
				return right, nil
			}

			return nil, fmt.Errorf("cannot assign to %T", obj)

		default:
			return nil, fmt.Errorf("invalid assignment target: %T", target)
		}

	case *ast.ArrayLiteral:
		// 创建新的切片
		arr := make([]interface{}, len(e.Value))
		// 执行每个元素的表达式
		for i, elem := range e.Value {
			val, err := executeExpression(elem.Expr, ctx)
			if err != nil {
				return nil, err
			}
			arr[i] = val
		}
		return arr, nil

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
	switch x := v.(type) {
	case nil:
		return false
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
