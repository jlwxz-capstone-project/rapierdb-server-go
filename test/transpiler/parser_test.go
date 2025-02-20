package main

import (
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/parser"
)

func TestParseSimpleExpression(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "数字字面量",
			input:   "42;",
			wantErr: false,
		},
		{
			name:    "字符串字面量",
			input:   "'hello world';",
			wantErr: false,
		},
		{
			name:    "简单二元表达式",
			input:   "1 + 2;",
			wantErr: false,
		},
		{
			name:    "变量声明",
			input:   "let x = 10;",
			wantErr: false,
		},
		{
			name:    "函数调用",
			input:   "console.log('hello');",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := parser.ParseFile(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && program == nil {
				t.Error("ParseFile() returned nil program without error")
			}
		})
	}
}

func TestParseComplexExpression(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "对象字面量",
			input: `const obj = {
				name: 'John',
				age: 30,
				greet: function() {
					console.log('Hello, ' + this.name);
				}
			};`,
			wantErr: false,
		},
		{
			name: "箭头函数",
			input: `const add = (a, b) => {
				return a + b;
			};`,
			wantErr: false,
		},
		{
			name:    "模板字符串",
			input:   "const message = `Hello, ${name}!`;",
			wantErr: false,
		},
		{
			name:    "解构赋值",
			input:   "const { name, age } = person;",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := parser.ParseFile(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && program == nil {
				t.Error("ParseFile() returned nil program without error")
			}
		})
	}
}

func TestParseStatements(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "if语句",
			input: `if (x > 0) {
				console.log('positive');
			} else {
				console.log('non-positive');
			}`,
			wantErr: false,
		},
		{
			name: "for循环",
			input: `for (let i = 0; i < 10; i++) {
				console.log(i);
			}`,
			wantErr: false,
		},
		{
			name: "try-catch",
			input: `try {
				riskyOperation();
			} catch (err) {
				console.error(err);
			}`,
			wantErr: false,
		},
		{
			name: "switch语句",
			input: `switch (value) {
				case 1:
					console.log('one');
					break;
				case 2:
					console.log('two');
					break;
				default:
					console.log('other');
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := parser.ParseFile(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && program == nil {
				t.Error("ParseFile() returned nil program without error")
			}
		})
	}
}

func TestParseInvalidSyntax(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "未闭合的字符串",
			input:   "'unclosed string;",
			wantErr: true,
		},
		{
			name:    "无效的标识符",
			input:   "123abc = 456;",
			wantErr: true,
		},
		{
			name:    "缺少分号",
			input:   "let x = 10\nlet y = 20",
			wantErr: true,
		},
		{
			name:    "未闭合的括号",
			input:   "if (x > 0 {",
			wantErr: true,
		},
		{
			name:    "无效的运算符组合",
			input:   "x +* y;",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.ParseFile(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseProgram(t *testing.T) {
	input := `
		// 这是一个完整的程序示例
		const MAX_VALUE = 100;

		function calculateSum(numbers) {
			let sum = 0;
			for (const num of numbers) {
				if (num > MAX_VALUE) {
					throw new Error('Number too large');
				}
				sum += num;
			}
			return sum;
		}

		try {
			const result = calculateSum([1, 2, 3, 4, 5]);
			console.log('Sum is: ' + result);
		} catch (err) {
			console.error('Calculation failed:', err.message);
		}
	`

	program, err := parser.ParseFile(input)
	if err != nil {
		t.Errorf("ParseFile() failed to parse valid program: %v", err)
		return
	}

	if program == nil {
		t.Error("ParseFile() returned nil program without error")
		return
	}

	// 验证程序结构
	if len(program.Body) == 0 {
		t.Error("Program body is empty")
	}
}

func TestParseClass(t *testing.T) {
	input := `
		class Person {
			constructor(name, age) {
				this.name = name;
				this.age = age;
			}

			greet() {
				return 'Hello, my name is ' + this.name;
			}

			static create(name, age) {
				return new Person(name, age);
			}
		}

		const person = new Person('John', 30);
	`

	program, err := parser.ParseFile(input)
	if err != nil {
		t.Errorf("ParseFile() failed to parse valid class: %v", err)
		return
	}

	if program == nil {
		t.Error("ParseFile() returned nil program without error")
		return
	}

	// 验证程序结构
	if len(program.Body) == 0 {
		t.Error("Program body is empty")
	}
}

func TestTemplateString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "简单模板字符串",
			input:   "const msg = `Hello, world!`;",
			wantErr: false,
		},
		{
			name:    "带表达式的模板字符串",
			input:   "const msg = `The sum of 1 + 1 is ${1 + 1}`;",
			wantErr: false,
		},
		{
			name:    "多行模板字符串",
			input:   "const msg = `Hello, ${name}! How are you?`;",
			wantErr: false,
		},
		{
			name:    "嵌套模板字符串",
			input:   "const msg = `Outer ${`Inner ${value}`}`;",
			wantErr: false,
		},
		{
			name:    "带函数调用的模板字符串",
			input:   "const msg = `Result: ${calculate(x, y)}`;",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := parser.ParseFile(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && program == nil {
				t.Error("ParseFile() returned nil program without error")
			}
		})
	}
}
