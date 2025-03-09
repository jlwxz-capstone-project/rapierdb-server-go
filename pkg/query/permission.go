package query

import (
	"errors"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/ast"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/parser"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/transpiler"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

var ErrInvalidPermissionDefinition = errors.New("invalid permission definition")

type Permissions struct {
	// 权限定义的版本
	Version string
	// 权限定义的规则
	Rules map[string]CollectionRule
	// 权限定义的 Js 代码，序列化和反序列化时仅处理 JsDef 即可
	JsDef string
}

type CollectionRuleFunc = func(...any) (any, error)

type CollectionRule struct {
	CanView   CollectionRuleFunc
	CanCreate CollectionRuleFunc
	CanUpdate CollectionRuleFunc
	CanDelete CollectionRuleFunc
}

type CanViewParams struct {
	Collection string
	DocId      string
	Doc        *loro.LoroDoc
	ClientId   string
	Db         *DbWrapper
}

// CanView 检查客户端是否有权限查看指定集合中的指定文档
func (p *Permissions) CanView(params CanViewParams) bool {
	rule, ok := p.Rules[params.Collection]
	if !ok {
		return false
	}
	ret, err := rule.CanView(params)
	if err != nil {
		fmt.Printf("can view error: %v", err)
		return false
	}
	if b, ok := ret.(bool); ok {
		return b
	}
	return false
}

type CanCreateParams struct {
	Collection string
	DocId      string
	NewDoc     *loro.LoroDoc
	ClientId   string
	Db         *DbWrapper
}

// CanCreate 检查客户端是否有权限创建指定集合中的文档
func (p *Permissions) CanCreate(params CanCreateParams) bool {
	rule, ok := p.Rules[params.Collection]
	if !ok {
		return false
	}
	ret, err := rule.CanCreate(params)
	if err != nil {
		fmt.Printf("can create error: %+v\n", err)
		return false
	}
	if b, ok := ret.(bool); ok {
		return b
	}
	return false
}

type CanUpdateParams struct {
	Collection string
	DocId      string
	NewDoc     *loro.LoroDoc
	OldDoc     *loro.LoroDoc
	ClientId   string
	Db         *DbWrapper
}

// CanUpdate 检查客户端是否有权限更新指定集合中的指定文档
func (p *Permissions) CanUpdate(params CanUpdateParams) bool {
	rule, ok := p.Rules[params.Collection]
	if !ok {
		return false
	}
	ret, err := rule.CanUpdate(params)
	if err != nil {
		fmt.Printf("can update error: %v", err)
		return false
	}
	if b, ok := ret.(bool); ok {
		return b
	}
	return false
}

type CanDeleteParams struct {
	Collection string
	DocId      string
	Doc        *loro.LoroDoc
	ClientId   string
	Db         *DbWrapper
}

// CanDelete 检查客户端是否有权限删除指定集合中的指定文档
func (p *Permissions) CanDelete(params CanDeleteParams) bool {
	rule, ok := p.Rules[params.Collection]
	if !ok {
		return false
	}
	ret, err := rule.CanDelete(params)
	if err != nil {
		fmt.Printf("can delete error: %v", err)
		return false
	}
	if b, ok := ret.(bool); ok {
		return b
	}
	return false
}

func (cr *CollectionRule) SetValidator(name string, fn CollectionRuleFunc) {
	switch name {
	case "canView":
		cr.CanView = fn
	case "canCreate":
		cr.CanCreate = fn
	case "canUpdate":
		cr.CanUpdate = fn
	case "canDelete":
		cr.CanDelete = fn
	default:
		panic("invalid rule name")
	}
}

func NewPermissionFuncScope() *transpiler.Scope {
	propGetter := transpiler.NewPropGetter(
		DbWrapperAccessHandler,
		CollectionWrapperAccessHandler,
		LoroDocAccessHandler,
		LoroTextAccessHandler,
		LoroMapAccessHandler,
		LoroListAccessHandler,
		LoroMovableListAccessHandler,
		transpiler.StringPropAccessHandler,
		transpiler.ArrayPropAccessHandler,
		transpiler.DataFieldAccessHandler,
		transpiler.MethodCallHandler,
	)
	propMutator := transpiler.DefaultPropSetter
	scope := transpiler.NewScope(nil, propGetter, propMutator)
	scope.Vars["eq"] = EqWrapper
	scope.Vars["field"] = FieldWrapper
	scope.Vars["asc"] = SortAscWrapper
	scope.Vars["desc"] = SortDescWrapper
	scope.Vars["log"] = LogWrapper // TODO 仅用于测试
	return scope
}

// NewPermissionFromJs 从 Js 权限定义中生成 Go 权限定义
//
// 下面是一个权限定义的例子：
//
//			Permission.create({
//			  version: "1.0.0",
//			  rules: {
//			    users: {
//		        // 仅有管理员和用户自己可以查看自己的信息
//			      canView: (doc, ctx) => ctx.user.role === "admin" || data.userId === ctx.user.userId,
//			      // 仅有管理员可以创建用户
//	         canCreate: (doc, ctx) => ctx.user.role === "admin",
//			      // 仅有管理员可以更新用户
//			      canUpdate: (doc, ctx) => ctx.user.role === "admin",
//			      // 仅有管理员可以删除用户
//			      canDelete: (doc, ctx) => ctx.user.role === "admin",
//			    },
//			  },
//			});
//
// 我们需要先使用 parser 解析出这个 Js 权限定义的 AST，
// 然后手工提取出所有集合对应的 canView, canCreate, canUpdate, canDelete 四个函数
// 然后使用 transpiler 将这四个函数转换为 Go 函数
func NewPermissionFromJs(js string) (*Permissions, error) {
	program, err := parser.ParseFile(js)
	if err != nil {
		return nil, err
	}
	permission := Permissions{
		Rules: make(map[string]CollectionRule),
		JsDef: js,
	}

	exprStmt, ok := program.Body[0].Stmt.(*ast.ExpressionStatement)
	if !ok {
		return nil, ErrInvalidPermissionDefinition
	}
	callExpr, ok := exprStmt.Expression.Expr.(*ast.CallExpression)
	if !ok {
		return nil, ErrInvalidPermissionDefinition
	}
	memberExpr, ok := callExpr.Callee.Expr.(*ast.MemberExpression)
	if !ok {
		return nil, ErrInvalidPermissionDefinition
	}
	objectExpr, ok := memberExpr.Object.Expr.(*ast.Identifier)
	if !ok {
		return nil, ErrInvalidPermissionDefinition
	}
	if objectExpr.Name != "Permission" {
		return nil, ErrInvalidPermissionDefinition
	}
	propertyExpr, ok := memberExpr.Property.Prop.(*ast.Identifier)
	if !ok {
		return nil, ErrInvalidPermissionDefinition
	}
	if propertyExpr.Name != "create" {
		return nil, ErrInvalidPermissionDefinition
	}
	args := callExpr.ArgumentList
	if len(args) != 1 {
		return nil, ErrInvalidPermissionDefinition
	}
	arg0, ok := args[0].Expr.(*ast.ObjectLiteral)
	if !ok {
		return nil, ErrInvalidPermissionDefinition
	}
	if len(arg0.Value) != 2 {
		return nil, ErrInvalidPermissionDefinition
	}
	prop0, ok := arg0.Value[0].Prop.(*ast.PropertyKeyed)
	if !ok {
		return nil, ErrInvalidPermissionDefinition
	}
	prop0Key, ok := prop0.Key.Expr.(*ast.StringLiteral)
	if !ok {
		return nil, ErrInvalidPermissionDefinition
	}
	if prop0Key.Value != "version" {
		return nil, ErrInvalidPermissionDefinition
	}
	versionExpr, ok := prop0.Value.Expr.(*ast.StringLiteral)
	if !ok {
		return nil, ErrInvalidPermissionDefinition
	}
	permission.Version = versionExpr.Value
	prop1, ok := arg0.Value[1].Prop.(*ast.PropertyKeyed)
	if !ok {
		return nil, ErrInvalidPermissionDefinition
	}
	prop1Key, ok := prop1.Key.Expr.(*ast.StringLiteral)
	if !ok {
		return nil, ErrInvalidPermissionDefinition
	}
	if prop1Key.Value != "rules" {
		return nil, ErrInvalidPermissionDefinition
	}
	rulesExpr, ok := prop1.Value.Expr.(*ast.ObjectLiteral)
	if !ok {
		return nil, ErrInvalidPermissionDefinition
	}

	for _, prop := range rulesExpr.Value {
		propKeyed, ok := prop.Prop.(*ast.PropertyKeyed)
		if !ok {
			return nil, ErrInvalidPermissionDefinition
		}
		propKey, ok := propKeyed.Key.Expr.(*ast.StringLiteral)
		if !ok {
			return nil, ErrInvalidPermissionDefinition
		}
		collectionName := propKey.Value
		collectionRule := CollectionRule{}
		ruleFuncs, ok := propKeyed.Value.Expr.(*ast.ObjectLiteral)
		if !ok {
			return nil, ErrInvalidPermissionDefinition
		}
		for _, ruleFunc := range ruleFuncs.Value {
			ruleFuncKeyed, ok := ruleFunc.Prop.(*ast.PropertyKeyed)
			if !ok {
				return nil, ErrInvalidPermissionDefinition
			}
			ruleFuncKey, ok := ruleFuncKeyed.Key.Expr.(*ast.StringLiteral)
			if !ok {
				return nil, ErrInvalidPermissionDefinition
			}
			ruleFuncName := ruleFuncKey.Value
			if ruleFuncName != "canView" && ruleFuncName != "canCreate" && ruleFuncName != "canUpdate" && ruleFuncName != "canDelete" {
				return nil, ErrInvalidPermissionDefinition
			}
			var ruleFuncExpr ast.Expr
			switch ruleFuncKeyed.Value.Expr.(type) {
			case *ast.ArrowFunctionLiteral:
				ruleFuncExpr = ruleFuncKeyed.Value.Expr
			case *ast.FunctionLiteral:
				ruleFuncExpr = ruleFuncKeyed.Value.Expr
			default:
				return nil, ErrInvalidPermissionDefinition
			}
			scope := NewPermissionFuncScope()
			goFunc, err := transpiler.TranspileJsAstToGoFunc(ruleFuncExpr, scope)
			if err != nil {
				return nil, err
			}
			collectionRule.SetValidator(ruleFuncName, goFunc)
		}
		permission.Rules[collectionName] = collectionRule
	}

	return &permission, nil
}
