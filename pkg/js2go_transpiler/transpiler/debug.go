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

	case *ast.BlockStatement:
		fmt.Printf("%s%s BlockStatement:\n", indent, prefix)
		for i, stmt := range n.List {
			printNode(stmt.Stmt, depth+1, fmt.Sprintf("Stmt[%d]", i))
		}

	case *ast.VariableDeclaration:
		fmt.Printf("%s%s VariableDeclaration:\n", indent, prefix)
		for i, decl := range n.List {
			printNode(decl.Target.Target, depth+1, fmt.Sprintf("Target[%d]", i))
			if decl.Initializer != nil {
				printNode(decl.Initializer.Expr, depth+1, fmt.Sprintf("Init[%d]", i))
			}
		}

	case *ast.AssignExpression:
		fmt.Printf("%s%s AssignExpression:\n", indent, prefix)
		printNode(n.Left.Expr, depth+1, "Left")
		printNode(n.Right.Expr, depth+1, "Right")

	case *ast.IfStatement:
		fmt.Printf("%s%s IfStatement:\n", indent, prefix)
		printNode(n.Test.Expr, depth+1, "Test")
		printNode(n.Consequent.Stmt, depth+1, "Consequent")
		if n.Alternate != nil {
			printNode(n.Alternate.Stmt, depth+1, "Alternate")
		}

	case *ast.ConditionalExpression:
		fmt.Printf("%s%s ConditionalExpression:\n", indent, prefix)
		printNode(n.Test.Expr, depth+1, "Test")
		printNode(n.Consequent.Expr, depth+1, "Consequent")
		printNode(n.Alternate.Expr, depth+1, "Alternate")

	case *ast.BinaryExpression:
		fmt.Printf("%s%s BinaryExpression: %s\n", indent, prefix, n.Operator)
		printNode(n.Left.Expr, depth+1, "Left")
		printNode(n.Right.Expr, depth+1, "Right")

	case *ast.UnaryExpression:
		fmt.Printf("%s%s UnaryExpression: %s\n", indent, prefix, n.Operator)
		printNode(n.Operand.Expr, depth+1, "Operand")

	case *ast.EmptyStatement:
		fmt.Printf("%s%s EmptyStatement\n", indent, prefix)

	default:
		fmt.Printf("%s%s Unknown node type: %T\n", indent, prefix, n)
	}
}
