package permissions

import (
	_ "embed"
	"errors"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/ast"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/parser"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/transpiler"
)

var ErrInvalidPermissionDefinition = errors.New("invalid permission definition")

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
func NewPermissionFromJs(js string) (any, error) {
	program, err := parser.ParseFile(js)
	if err != nil {
		return nil, err
	}
	transpiler.PrintProgram(program)

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
	prop0Key, ok := prop0.Key.Expr.(*ast.Identifier)
	if !ok {
		return nil, ErrInvalidPermissionDefinition
	}
	if prop0Key.Name != "version" {
		return nil, ErrInvalidPermissionDefinition
	}
	versionExpr, ok := prop0.Key.Expr.(*ast.StringLiteral)
	if !ok {
		return nil, ErrInvalidPermissionDefinition
	}
	version := versionExpr.Value
	prop1, ok := arg0.Value[1].Prop.(*ast.PropertyKeyed)
	if !ok {
		return nil, ErrInvalidPermissionDefinition
	}
	prop1Key, ok := prop1.Key.Expr.(*ast.Identifier)
	if !ok {
		return nil, ErrInvalidPermissionDefinition
	}
	if prop1Key.Name != "rules" {
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
		propKey, ok := propKeyed.Key.Expr.(*ast.Identifier)
		if !ok {
			return nil, ErrInvalidPermissionDefinition
		}
		collection := propKey.Name
		ruleFuncs, ok := propKeyed.Value.Expr.(*ast.ObjectLiteral)
		if !ok {
			return nil, ErrInvalidPermissionDefinition
		}
		for _, ruleFunc := range ruleFuncs.Value {
			ruleFuncKeyed, ok := ruleFunc.Prop.(*ast.PropertyKeyed)
			if !ok {
				return nil, ErrInvalidPermissionDefinition
			}
			ruleFuncKey, ok := ruleFuncKeyed.Key.Expr.(*ast.Identifier)
			if !ok {
				return nil, ErrInvalidPermissionDefinition
			}
			ruleFuncName := ruleFuncKey.Name
			if ruleFuncName != "canView" && ruleFuncName != "canCreate" && ruleFuncName != "canUpdate" && ruleFuncName != "canDelete" {
				return nil, ErrInvalidPermissionDefinition
			}
			ruleFuncExpr, ok := ruleFuncKeyed.Value.Expr.(*ast.ArrowFunctionLiteral)
			if !ok {
				return nil, ErrInvalidPermissionDefinition
			}

		}
	}

	return nil, nil
}
