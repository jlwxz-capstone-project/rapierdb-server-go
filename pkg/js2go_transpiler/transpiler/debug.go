package transpiler

import (
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/ast"
)

func PrintProgram(program *ast.Program) {
	fmt.Println("Program:")
	for i, stmt := range program.Body {
		PrintNode(stmt.Stmt, 1, fmt.Sprintf("[%d]", i))
	}
}

func PrintNode(node ast.Node, depth int, prefix string) {
	indent := strings.Repeat("  ", depth)

	switch n := node.(type) {
	case *ast.ExpressionStatement:
		fmt.Printf("%s%s ExpressionStatement:\n", indent, prefix)
		PrintNode(n.Expression.Expr, depth+1, "Expression")

	case *ast.MemberExpression:
		fmt.Printf("%s%s MemberExpression:\n", indent, prefix)
		PrintNode(n.Object.Expr, depth+1, "Object")
		// MemberProperty 包含 Prop 字段，可以是 Identifier 或 ComputedProperty
		switch prop := n.Property.Prop.(type) {
		case *ast.Identifier:
			fmt.Printf("%s  Property (Identifier): %s\n", indent, prop.Name)
		case *ast.ComputedProperty:
			fmt.Printf("%s  Property (Computed):\n", indent)
			PrintNode(prop.Expr.Expr, depth+2, "Expression")
		}

	case *ast.CallExpression:
		fmt.Printf("%s%s CallExpression:\n", indent, prefix)
		PrintNode(n.Callee.Expr, depth+1, "Callee")
		for i, arg := range n.ArgumentList {
			PrintNode(arg.Expr, depth+1, fmt.Sprintf("Arg[%d]", i))
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
			PrintNode(prop.Prop, depth+1, fmt.Sprintf("Prop[%d]", i))
		}

	case *ast.PropertyKeyed:
		fmt.Printf("%s%s PropertyKeyed:\n", indent, prefix)
		PrintNode(n.Key.Expr, depth+1, "Key")
		PrintNode(n.Value.Expr, depth+1, "Value")

	case *ast.BlockStatement:
		fmt.Printf("%s%s BlockStatement:\n", indent, prefix)
		for i, stmt := range n.List {
			PrintNode(stmt.Stmt, depth+1, fmt.Sprintf("Stmt[%d]", i))
		}

	case *ast.VariableDeclaration:
		fmt.Printf("%s%s VariableDeclaration:\n", indent, prefix)
		for i, decl := range n.List {
			PrintNode(decl.Target.Target, depth+1, fmt.Sprintf("Target[%d]", i))
			if decl.Initializer != nil {
				PrintNode(decl.Initializer.Expr, depth+1, fmt.Sprintf("Init[%d]", i))
			}
		}

	case *ast.AssignExpression:
		fmt.Printf("%s%s AssignExpression:\n", indent, prefix)
		PrintNode(n.Left.Expr, depth+1, "Left")
		PrintNode(n.Right.Expr, depth+1, "Right")

	case *ast.IfStatement:
		fmt.Printf("%s%s IfStatement:\n", indent, prefix)
		PrintNode(n.Test.Expr, depth+1, "Test")
		PrintNode(n.Consequent.Stmt, depth+1, "Consequent")
		if n.Alternate != nil {
			PrintNode(n.Alternate.Stmt, depth+1, "Alternate")
		}

	case *ast.ConditionalExpression:
		fmt.Printf("%s%s ConditionalExpression:\n", indent, prefix)
		PrintNode(n.Test.Expr, depth+1, "Test")
		PrintNode(n.Consequent.Expr, depth+1, "Consequent")
		PrintNode(n.Alternate.Expr, depth+1, "Alternate")

	case *ast.BinaryExpression:
		fmt.Printf("%s%s BinaryExpression: %s\n", indent, prefix, n.Operator)
		PrintNode(n.Left.Expr, depth+1, "Left")
		PrintNode(n.Right.Expr, depth+1, "Right")

	case *ast.UnaryExpression:
		fmt.Printf("%s%s UnaryExpression: %s\n", indent, prefix, n.Operator)
		PrintNode(n.Operand.Expr, depth+1, "Operand")

	case *ast.EmptyStatement:
		fmt.Printf("%s%s EmptyStatement\n", indent, prefix)

	case *ast.FunctionDeclaration:
		fmt.Printf("%s%s FunctionDeclaration: %s\n", indent, prefix, n.Function.Name.Name)
		fmt.Printf("%s  Parameters: %s\n", indent, formatParams(n.Function.ParameterList))
		fmt.Printf("%s  Body:\n", indent)
		PrintNode(n.Function.Body, depth+2, "Body")

	case *ast.ArrowFunctionLiteral:
		fmt.Printf("%s%s ArrowFunction:\n", indent, prefix)
		fmt.Printf("%s  Parameters: %s\n", indent, formatParams(n.ParameterList))
		fmt.Printf("%s  Body:\n", indent)
		switch body := n.Body.Body.(type) {
		case *ast.BlockStatement:
			PrintNode(body, depth+2, "Block")
		case *ast.Expression:
			PrintNode(body.Expr, depth+2, "Expression")
		}

	case *ast.FunctionLiteral:
		name := "anonymous"
		if n.Name != nil {
			name = n.Name.Name
		}
		fmt.Printf("%s%s FunctionLiteral: %s\n", indent, prefix, name)
		fmt.Printf("%s  Parameters: %s\n", indent, formatParams(n.ParameterList))
		fmt.Printf("%s  Body:\n", indent)
		PrintNode(n.Body, depth+2, "Block")

	case *ast.ReturnStatement:
		fmt.Printf("%s%s ReturnStatement:\n", indent, prefix)
		if n.Argument != nil {
			PrintNode(n.Argument.Expr, depth+1, "Argument")
		} else {
			fmt.Printf("%s  (no argument)\n", indent)
		}

	default:
		fmt.Printf("%s%s Unknown node type: %T\n", indent, prefix, n)
	}
}

func formatParams(pl ast.ParameterList) string {
	params := make([]string, len(pl.List))
	for i, p := range pl.List {
		if id, ok := p.Target.Target.(*ast.Identifier); ok {
			params[i] = id.Name
		} else {
			params[i] = "complex_param"
		}
	}
	return strings.Join(params, ", ")
}
