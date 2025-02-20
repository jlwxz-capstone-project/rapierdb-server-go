package transpiler

import (
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/ast"
)

func PrintProgram(program *ast.Program) {
	fmt.Println("Program:")
	for i, stmt := range program.Body {
		printNode(stmt.Stmt, 1, fmt.Sprintf("[%d]", i))
	}
}

func printNode(node ast.Node, depth int, prefix string) {
	indent := strings.Repeat("  ", depth)

	switch n := node.(type) {
	case *ast.ExpressionStatement:
		fmt.Printf("%s%s ExpressionStatement:\n", indent, prefix)
		printNode(n.Expression.Expr, depth+1, "Expression")

	case *ast.MemberExpression:
		fmt.Printf("%s%s MemberExpression:\n", indent, prefix)
		printNode(n.Object.Expr, depth+1, "Object")
		// MemberProperty 包含 Prop 字段，可以是 Identifier 或 ComputedProperty
		switch prop := n.Property.Prop.(type) {
		case *ast.Identifier:
			fmt.Printf("%s  Property (Identifier): %s\n", indent, prop.Name)
		case *ast.ComputedProperty:
			fmt.Printf("%s  Property (Computed):\n", indent)
			printNode(prop.Expr.Expr, depth+2, "Expression")
		}

	case *ast.CallExpression:
		fmt.Printf("%s%s CallExpression:\n", indent, prefix)
		printNode(n.Callee.Expr, depth+1, "Callee")
		for i, arg := range n.ArgumentList {
			printNode(arg.Expr, depth+1, fmt.Sprintf("Arg[%d]", i))
		}

	case *ast.Identifier:
		fmt.Printf("%s%s Identifier: %s\n", indent, prefix, n.Name)

	case *ast.StringLiteral:
		fmt.Printf("%s%s StringLiteral: %q\n", indent, prefix, n.Value)

	case *ast.NumberLiteral:
		fmt.Printf("%s%s NumberLiteral: %v\n", indent, prefix, n.Value)

	case *ast.BooleanLiteral:
		fmt.Printf("%s%s BooleanLiteral: %v\n", indent, prefix, n.Value)

	case *ast.NullLiteral:
		fmt.Printf("%s%s NullLiteral\n", indent, prefix)

	case *ast.ObjectLiteral:
		fmt.Printf("%s%s ObjectLiteral:\n", indent, prefix)
		for i, prop := range n.Value {
			printNode(prop.Prop, depth+1, fmt.Sprintf("Prop[%d]", i))
		}

	case *ast.PropertyKeyed:
		fmt.Printf("%s%s PropertyKeyed:\n", indent, prefix)
		printNode(n.Key.Expr, depth+1, "Key")
		printNode(n.Value.Expr, depth+1, "Value")

	default:
		fmt.Printf("%s%s Unknown node type: %T\n", indent, prefix, n)
	}
}
