package ast

import (
	"fmt"
	"strings"
)

func DebugPrintProgram(program *Program) string {
	var sb strings.Builder
	sb.WriteString("Program:\n")
	for i, stmt := range program.Body {
		sb.WriteString(DebugPrintNode(stmt.Stmt, 1, fmt.Sprintf("[%d]", i)))
	}
	return sb.String()
}

func DebugPrintNode(node Node, depth int, prefix string) string {
	indent := strings.Repeat("  ", depth)
	var sb strings.Builder

	switch n := node.(type) {
	case *ExpressionStatement:
		sb.WriteString(fmt.Sprintf("%s%s ExpressionStatement:\n", indent, prefix))
		sb.WriteString(DebugPrintNode(n.Expression.Expr, depth+1, "Expression"))

	case *MemberExpression:
		sb.WriteString(fmt.Sprintf("%s%s MemberExpression:\n", indent, prefix))
		sb.WriteString(DebugPrintNode(n.Object.Expr, depth+1, "Object"))
		// MemberProperty 包含 Prop 字段，可以是 Identifier 或 ComputedProperty
		switch prop := n.Property.Prop.(type) {
		case *Identifier:
			sb.WriteString(fmt.Sprintf("%s  Property (Identifier): %s\n", indent, prop.Name))
		case *ComputedProperty:
			sb.WriteString(fmt.Sprintf("%s  Property (Computed):\n", indent))
			sb.WriteString(DebugPrintNode(prop.Expr.Expr, depth+2, "Expression"))
		}

	case *CallExpression:
		sb.WriteString(fmt.Sprintf("%s%s CallExpression:\n", indent, prefix))
		sb.WriteString(DebugPrintNode(n.Callee.Expr, depth+1, "Callee"))
		for i, arg := range n.ArgumentList {
			sb.WriteString(DebugPrintNode(arg.Expr, depth+1, fmt.Sprintf("Arg[%d]", i)))
		}

	case *Identifier:
		sb.WriteString(fmt.Sprintf("%s%s Identifier: %s\n", indent, prefix, n.Name))

	case *StringLiteral:
		sb.WriteString(fmt.Sprintf("%s%s StringLiteral: %q\n", indent, prefix, n.Value))

	case *NumberLiteral:
		sb.WriteString(fmt.Sprintf("%s%s NumberLiteral: %v\n", indent, prefix, n.Value))

	case *BooleanLiteral:
		sb.WriteString(fmt.Sprintf("%s%s BooleanLiteral: %v\n", indent, prefix, n.Value))

	case *NullLiteral:
		sb.WriteString(fmt.Sprintf("%s%s NullLiteral\n", indent, prefix))

	case *ObjectLiteral:
		sb.WriteString(fmt.Sprintf("%s%s ObjectLiteral:\n", indent, prefix))
		for i, prop := range n.Value {
			sb.WriteString(DebugPrintNode(prop.Prop, depth+1, fmt.Sprintf("Prop[%d]", i)))
		}

	case *PropertyKeyed:
		sb.WriteString(fmt.Sprintf("%s%s PropertyKeyed:\n", indent, prefix))
		sb.WriteString(DebugPrintNode(n.Key.Expr, depth+1, "Key"))
		sb.WriteString(DebugPrintNode(n.Value.Expr, depth+1, "Value"))

	case *BlockStatement:
		sb.WriteString(fmt.Sprintf("%s%s BlockStatement:\n", indent, prefix))
		for i, stmt := range n.List {
			sb.WriteString(DebugPrintNode(stmt.Stmt, depth+1, fmt.Sprintf("Stmt[%d]", i)))
		}

	case *VariableDeclaration:
		sb.WriteString(fmt.Sprintf("%s%s VariableDeclaration:\n", indent, prefix))
		for i, decl := range n.List {
			sb.WriteString(DebugPrintNode(decl.Target.Target, depth+1, fmt.Sprintf("Target[%d]", i)))
			if decl.Initializer != nil {
				sb.WriteString(DebugPrintNode(decl.Initializer.Expr, depth+1, fmt.Sprintf("Init[%d]", i)))
			}
		}

	case *AssignExpression:
		sb.WriteString(fmt.Sprintf("%s%s AssignExpression:\n", indent, prefix))
		sb.WriteString(DebugPrintNode(n.Left.Expr, depth+1, "Left"))
		sb.WriteString(DebugPrintNode(n.Right.Expr, depth+1, "Right"))

	case *IfStatement:
		sb.WriteString(fmt.Sprintf("%s%s IfStatement:\n", indent, prefix))
		sb.WriteString(DebugPrintNode(n.Test.Expr, depth+1, "Test"))
		sb.WriteString(DebugPrintNode(n.Consequent.Stmt, depth+1, "Consequent"))
		if n.Alternate != nil {
			sb.WriteString(DebugPrintNode(n.Alternate.Stmt, depth+1, "Alternate"))
		}

	case *ConditionalExpression:
		sb.WriteString(fmt.Sprintf("%s%s ConditionalExpression:\n", indent, prefix))
		sb.WriteString(DebugPrintNode(n.Test.Expr, depth+1, "Test"))
		sb.WriteString(DebugPrintNode(n.Consequent.Expr, depth+1, "Consequent"))
		sb.WriteString(DebugPrintNode(n.Alternate.Expr, depth+1, "Alternate"))

	case *BinaryExpression:
		sb.WriteString(fmt.Sprintf("%s%s BinaryExpression: %s\n", indent, prefix, n.Operator))
		sb.WriteString(DebugPrintNode(n.Left.Expr, depth+1, "Left"))
		sb.WriteString(DebugPrintNode(n.Right.Expr, depth+1, "Right"))

	case *UnaryExpression:
		sb.WriteString(fmt.Sprintf("%s%s UnaryExpression: %s\n", indent, prefix, n.Operator))
		sb.WriteString(DebugPrintNode(n.Operand.Expr, depth+1, "Operand"))

	case *EmptyStatement:
		sb.WriteString(fmt.Sprintf("%s%s EmptyStatement\n", indent, prefix))

	case *FunctionDeclaration:
		sb.WriteString(fmt.Sprintf("%s%s FunctionDeclaration: %s\n", indent, prefix, n.Function.Name.Name))
		sb.WriteString(fmt.Sprintf("%s  Parameters: %s\n", indent, formatParams(n.Function.ParameterList)))
		sb.WriteString(fmt.Sprintf("%s  Body:\n", indent))
		sb.WriteString(DebugPrintNode(n.Function.Body, depth+2, "Body"))

	case *ArrowFunctionLiteral:
		sb.WriteString(fmt.Sprintf("%s%s ArrowFunction:\n", indent, prefix))
		sb.WriteString(fmt.Sprintf("%s  Parameters: %s\n", indent, formatParams(n.ParameterList)))
		sb.WriteString(fmt.Sprintf("%s  Body:\n", indent))
		switch body := n.Body.Body.(type) {
		case *BlockStatement:
			sb.WriteString(DebugPrintNode(body, depth+2, "Block"))
		case *Expression:
			sb.WriteString(DebugPrintNode(body.Expr, depth+2, "Expression"))
		}

	case *FunctionLiteral:
		name := "anonymous"
		if n.Name != nil {
			name = n.Name.Name
		}
		sb.WriteString(fmt.Sprintf("%s%s FunctionLiteral: %s\n", indent, prefix, name))
		sb.WriteString(fmt.Sprintf("%s  Parameters: %s\n", indent, formatParams(n.ParameterList)))
		sb.WriteString(fmt.Sprintf("%s  Body:\n", indent))
		sb.WriteString(DebugPrintNode(n.Body, depth+2, "Block"))

	case *ReturnStatement:
		sb.WriteString(fmt.Sprintf("%s%s ReturnStatement:\n", indent, prefix))
		if n.Argument != nil {
			sb.WriteString(DebugPrintNode(n.Argument.Expr, depth+1, "Argument"))
		} else {
			sb.WriteString(fmt.Sprintf("%s  (no argument)\n", indent))
		}

	default:
		sb.WriteString(fmt.Sprintf("%s%s Unknown node type: %T\n", indent, prefix, n))
	}

	return sb.String()
}

func formatParams(pl ParameterList) string {
	params := make([]string, len(pl.List))
	for i, p := range pl.List {
		switch target := p.Target.Target.(type) {
		case *Identifier:
			// 简单标识符参数
			params[i] = target.Name
		case *ObjectPattern:
			// 对象解构赋值
			properties := make([]string, len(target.Properties))
			for j, prop := range target.Properties {
				switch pp := prop.Prop.(type) {
				case *PropertyShort:
					properties[j] = pp.Name.Name
				default:
					properties[j] = "unknown_prop"
				}
			}
			params[i] = "{" + strings.Join(properties, ", ") + "}"
		case *ArrayPattern:
			// 数组解构赋值
			elements := make([]string, 0, len(target.Elements))
			for _, elem := range target.Elements {
				if elem.Expr == nil {
					elements = append(elements, "_") // 跳过的元素
				} else if id, ok := elem.Expr.(*Identifier); ok {
					elements = append(elements, id.Name)
				} else {
					elements = append(elements, "complex_elem")
				}
			}
			params[i] = "[" + strings.Join(elements, ", ") + "]"
		default:
			params[i] = "complex_param"
		}

		// 添加默认值信息
		if p.Initializer != nil {
			params[i] += " = ..."
		}
	}
	return strings.Join(params, ", ")
}
