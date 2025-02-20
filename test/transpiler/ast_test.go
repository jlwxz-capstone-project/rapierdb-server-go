package main

import (
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/ast"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/token"
)

func TestBasicNodes(t *testing.T) {
	tests := []struct {
		name     string
		node     ast.Node
		wantIdx0 ast.Idx
		wantIdx1 ast.Idx
	}{
		{
			name:     "StringLiteral",
			node:     &ast.StringLiteral{Idx: 10, Raw: stringPtr("'hello'")},
			wantIdx0: 10,
			wantIdx1: 17, // 10 + len("'hello'")
		},
		{
			name:     "NumberLiteral",
			node:     &ast.NumberLiteral{Idx: 20, Raw: stringPtr("42")},
			wantIdx0: 20,
			wantIdx1: 22, // 20 + len("42")
		},
		{
			name:     "BooleanLiteral",
			node:     &ast.BooleanLiteral{Idx: 30},
			wantIdx0: 30,
			wantIdx1: 34, // 30 + len("true")
		},
		{
			name:     "NullLiteral",
			node:     &ast.NullLiteral{Idx: 40},
			wantIdx0: 40,
			wantIdx1: 44, // 40 + len("null")
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.Idx0(); got != tt.wantIdx0 {
				t.Errorf("Node.Idx0() = %v, want %v", got, tt.wantIdx0)
			}
			if got := tt.node.Idx1(); got != tt.wantIdx1 {
				t.Errorf("Node.Idx1() = %v, want %v", got, tt.wantIdx1)
			}
		})
	}
}

func TestIdentifier(t *testing.T) {
	ident := &ast.Identifier{
		Name: "testVar",
		Idx:  100,
	}

	if got := ident.Idx0(); got != 100 {
		t.Errorf("Identifier.Idx0() = %v, want %v", got, 100)
	}

	wantIdx1 := ast.Idx(100 + len("testVar"))
	if got := ident.Idx1(); got != wantIdx1 {
		t.Errorf("Identifier.Idx1() = %v, want %v", got, wantIdx1)
	}
}

func TestProgram(t *testing.T) {
	stmt := &ast.ExpressionStatement{
		Expression: &ast.Expression{
			Expr: &ast.StringLiteral{
				Idx: 1,
				Raw: stringPtr("'test'"),
			},
		},
	}

	program := &ast.Program{
		Body: []ast.Statement{{Stmt: stmt}},
	}

	if got := program.Idx0(); got != 1 {
		t.Errorf("Program.Idx0() = %v, want %v", got, 1)
	}

	wantIdx1 := ast.Idx(1 + len("'test'"))
	if got := program.Idx1(); got != wantIdx1 {
		t.Errorf("Program.Idx1() = %v, want %v", got, wantIdx1)
	}
}

func TestBinaryExpression(t *testing.T) {
	tests := []struct {
		name     string
		expr     *ast.BinaryExpression
		wantIdx0 ast.Idx
		wantIdx1 ast.Idx
	}{
		{
			name: "Addition",
			expr: &ast.BinaryExpression{
				Operator: token.Plus,
				Left: &ast.Expression{
					Expr: &ast.NumberLiteral{Idx: 1, Raw: stringPtr("5")},
				},
				Right: &ast.Expression{
					Expr: &ast.NumberLiteral{Idx: 3, Raw: stringPtr("3")},
				},
			},
			wantIdx0: 1,
			wantIdx1: 4,
		},
		{
			name: "Comparison",
			expr: &ast.BinaryExpression{
				Operator: token.Greater,
				Left: &ast.Expression{
					Expr: &ast.Identifier{Idx: 10, Name: "x"},
				},
				Right: &ast.Expression{
					Expr: &ast.NumberLiteral{Idx: 14, Raw: stringPtr("100")},
				},
			},
			wantIdx0: 10,
			wantIdx1: 17,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.expr.Idx0(); got != tt.wantIdx0 {
				t.Errorf("BinaryExpression.Idx0() = %v, want %v", got, tt.wantIdx0)
			}
			if got := tt.expr.Idx1(); got != tt.wantIdx1 {
				t.Errorf("BinaryExpression.Idx1() = %v, want %v", got, tt.wantIdx1)
			}
		})
	}
}

func TestCallExpression(t *testing.T) {
	expr := &ast.CallExpression{
		Callee: &ast.Expression{
			Expr: &ast.Identifier{Idx: 1, Name: "console"},
		},
		LeftParenthesis: 8,
		ArgumentList: []ast.Expression{
			{Expr: &ast.StringLiteral{Idx: 9, Raw: stringPtr("'Hello'")}},
		},
		RightParenthesis: 16,
	}

	if got := expr.Idx0(); got != 1 {
		t.Errorf("CallExpression.Idx0() = %v, want %v", got, 1)
	}
	if got := expr.Idx1(); got != 16 {
		t.Errorf("CallExpression.Idx1() = %v, want %v", got, 16)
	}
}

func TestIfStatement(t *testing.T) {
	stmt := &ast.IfStatement{
		If: 1,
		Test: &ast.Expression{
			Expr: &ast.BinaryExpression{
				Operator: token.StrictEqual,
				Left: &ast.Expression{
					Expr: &ast.Identifier{Idx: 5, Name: "x"},
				},
				Right: &ast.Expression{
					Expr: &ast.NumberLiteral{Idx: 10, Raw: stringPtr("10")},
				},
			},
		},
		Consequent: &ast.Statement{
			Stmt: &ast.BlockStatement{
				LeftBrace: 14,
				List: []ast.Statement{
					{
						Stmt: &ast.ReturnStatement{
							Return: 16,
							Argument: &ast.Expression{
								Expr: &ast.BooleanLiteral{Idx: 23},
							},
						},
					},
				},
				RightBrace: 29,
			},
		},
	}

	if got := stmt.Idx0(); got != 1 {
		t.Errorf("IfStatement.Idx0() = %v, want %v", got, 1)
	}
	if got := stmt.Idx1(); got != 29 {
		t.Errorf("IfStatement.Idx1() = %v, want %v", got, 29)
	}
}

func TestForStatement(t *testing.T) {
	stmt := &ast.ForStatement{
		For: 1,
		Initializer: &ast.ForLoopInitializer{
			Initializer: &ast.Expression{
				Expr: &ast.AssignExpression{
					Left: &ast.Expression{
						Expr: &ast.Identifier{Idx: 5, Name: "i"},
					},
					Operator: token.Assign,
					Right: &ast.Expression{
						Expr: &ast.NumberLiteral{Idx: 9, Raw: stringPtr("0")},
					},
				},
			},
		},
		Test: &ast.Expression{
			Expr: &ast.BinaryExpression{
				Left: &ast.Expression{
					Expr: &ast.Identifier{Idx: 12, Name: "i"},
				},
				Operator: token.Less,
				Right: &ast.Expression{
					Expr: &ast.NumberLiteral{Idx: 16, Raw: stringPtr("10")},
				},
			},
		},
		Update: &ast.Expression{
			Expr: &ast.UpdateExpression{
				Operator: token.Increment,
				Operand: &ast.Expression{
					Expr: &ast.Identifier{Idx: 20, Name: "i"},
				},
				Postfix: true,
			},
		},
		Body: &ast.Statement{
			Stmt: &ast.BlockStatement{
				LeftBrace:  24,
				List:       []ast.Statement{},
				RightBrace: 25,
			},
		},
	}

	if got := stmt.Idx0(); got != 1 {
		t.Errorf("ForStatement.Idx0() = %v, want %v", got, 1)
	}
	if got := stmt.Idx1(); got != 25 {
		t.Errorf("ForStatement.Idx1() = %v, want %v", got, 25)
	}
}

func TestTemplateLiteral(t *testing.T) {
	expr := &ast.TemplateLiteral{
		OpenQuote:  1,
		CloseQuote: 20,
		Elements: []ast.TemplateElement{
			{Idx: 2, Literal: "Hello ", Parsed: "Hello ", Valid: true},
			{Idx: 15, Literal: "!", Parsed: "!", Valid: true},
		},
		Expressions: []ast.Expression{
			{Expr: &ast.Identifier{Idx: 10, Name: "name"}},
		},
	}

	if got := expr.Idx0(); got != 1 {
		t.Errorf("TemplateLiteral.Idx0() = %v, want %v", got, 1)
	}
	if got := expr.Idx1(); got != 20 {
		t.Errorf("TemplateLiteral.Idx1() = %v, want %v", got, 20)
	}
}

func TestObjectLiteral(t *testing.T) {
	expr := &ast.ObjectLiteral{
		LeftBrace: 1,
		Value: []ast.Property{
			{
				Prop: &ast.PropertyKeyed{
					Key: &ast.Expression{
						Expr: &ast.Identifier{Idx: 3, Name: "name"},
					},
					Kind: ast.PropertyKindValue,
					Value: &ast.Expression{
						Expr: &ast.StringLiteral{Idx: 10, Raw: stringPtr("'John'")},
					},
				},
			},
			{
				Prop: &ast.PropertyKeyed{
					Key: &ast.Expression{
						Expr: &ast.Identifier{Idx: 18, Name: "age"},
					},
					Kind: ast.PropertyKindValue,
					Value: &ast.Expression{
						Expr: &ast.NumberLiteral{Idx: 24, Raw: stringPtr("30")},
					},
				},
			},
		},
		RightBrace: 27,
	}

	if got := expr.Idx0(); got != 1 {
		t.Errorf("ObjectLiteral.Idx0() = %v, want %v", got, 1)
	}
	if got := expr.Idx1(); got != 27 {
		t.Errorf("ObjectLiteral.Idx1() = %v, want %v", got, 27)
	}
}

func TestTryStatement(t *testing.T) {
	stmt := &ast.TryStatement{
		Try: 1,
		Body: &ast.BlockStatement{
			LeftBrace: 5,
			List: []ast.Statement{
				{
					Stmt: &ast.ExpressionStatement{
						Expression: &ast.Expression{
							Expr: &ast.CallExpression{
								Callee: &ast.Expression{
									Expr: &ast.Identifier{Idx: 7, Name: "riskyOperation"},
								},
								LeftParenthesis:  19,
								ArgumentList:     []ast.Expression{},
								RightParenthesis: 20,
							},
						},
					},
				},
			},
			RightBrace: 22,
		},
		Catch: &ast.CatchStatement{
			Catch: 24,
			Parameter: &ast.BindingTarget{
				Target: &ast.Identifier{Idx: 30, Name: "err"},
			},
			Body: &ast.BlockStatement{
				LeftBrace: 35,
				List: []ast.Statement{
					{
						Stmt: &ast.ExpressionStatement{
							Expression: &ast.Expression{
								Expr: &ast.CallExpression{
									Callee: &ast.Expression{
										Expr: &ast.MemberExpression{
											Object: &ast.Expression{
												Expr: &ast.Identifier{Idx: 37, Name: "console"},
											},
											Property: &ast.MemberProperty{
												Prop: &ast.Identifier{Idx: 45, Name: "error"},
											},
										},
									},
									LeftParenthesis: 50,
									ArgumentList: []ast.Expression{
										{Expr: &ast.Identifier{Idx: 51, Name: "err"}},
									},
									RightParenthesis: 54,
								},
							},
						},
					},
				},
				RightBrace: 56,
			},
		},
	}

	if got := stmt.Idx0(); got != 1 {
		t.Errorf("TryStatement.Idx0() = %v, want %v", got, 1)
	}
	if got := stmt.Idx1(); got != 56 {
		t.Errorf("TryStatement.Idx1() = %v, want %v", got, 56)
	}
}

func TestArrayPattern(t *testing.T) {
	pattern := &ast.ArrayPattern{
		LeftBracket:  1,
		RightBracket: 15,
		Elements: []ast.Expression{
			{Expr: &ast.Identifier{Idx: 2, Name: "first"}},
			{Expr: &ast.Identifier{Idx: 9, Name: "second"}},
		},
	}

	if got := pattern.Idx0(); got != 1 {
		t.Errorf("ArrayPattern.Idx0() = %v, want %v", got, 1)
	}
	if got := pattern.Idx1(); got != 15 {
		t.Errorf("ArrayPattern.Idx1() = %v, want %v", got, 15)
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
