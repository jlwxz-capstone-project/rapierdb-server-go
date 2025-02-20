package token

import (
	"strconv"
)

// ECMA5 Token
type Token int

// String 返回 token 的字符串表示
func (t Token) String() string {
	if t == 0 {
		return "UNKNOWN"
	}
	if t < Token(len(token2string)) {
		return token2string[t]
	}
	return "token(" + strconv.Itoa(int(t)) + ")"
}

// Precedence 返回 token 的优先级
func (t Token) Precedence(in bool) int {
	switch t {
	case LogicalOr:
		return 1
	case LogicalAnd:
		return 2
	case Or, OrAssign:
		return 3
	case ExclusiveOr:
		return 4
	case And, AndAssign:
		return 5
	case Equal,
		NotEqual,
		StrictEqual,
		StrictNotEqual:
		return 6
	case Less, Greater, LessOrEqual, GreaterOrEqual, InstanceOf:
		return 7
	case In:
		if in {
			return 7
		}
		return 0
	case ShiftLeft, ShiftRight, UnsignedShiftRight:
		fallthrough
	case ShiftLeftAssign, ShiftRightAssign, UnsignedShiftRightAssign:
		return 8
	case Plus, Minus, AddAssign, SubtractAssign:
		return 9
	case Multiply, Slash, Remainder, MultiplyAssign, QuotientAssign, RemainderAssign:
		return 11
	}
	return 0
}

// keyword 表示关键字的结构体
type keyword struct {
	token         Token // 关键字对应的 token
	futureKeyword bool  // 是否是未来的关键字 (如 const, let 等)
	strict        bool  // 是否是严格模式下的关键字
}

// LiteralKeyword 判断一个字面量是否是关键字
// 如果是普通关键字, 返回对应的 token 和 false
// 如果是未来关键字, 返回 Keyword token 和 strict 标志
// 如果不是关键字, 返回 0 和 false
func LiteralKeyword(literal string) (Token, bool) {
	if k, exists := keywordTable[literal]; exists {
		if k.futureKeyword {
			return Keyword, k.strict
		}
		return k.token, false
	}
	return 0, false
}

// ID 判断 token 是否是标识符类型 (包括关键字、标识符等)
func ID(token Token) bool {
	return token >= Identifier
}

// UnreservedWord 判断 token 是否是非保留字
func UnreservedWord(token Token) bool {
	return token > EscapedReservedWord
}
