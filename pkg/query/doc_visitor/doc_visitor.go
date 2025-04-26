package doc_visitor

import (
	"errors"
	"regexp"
	"strconv"
	"unsafe"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	pe "github.com/pkg/errors"
)

const DATA_MAP_NAME = "data"

var (
	PathNotFoundError = errors.New("path not found")
	InvalidPathError  = errors.New("invalid path")
)

func isNil(v any) bool {
	return (*[2]uintptr)(unsafe.Pointer(&v))[1] == 0
}

// VisitDocByPath 查询中，给定路径，返回文档中对应路径的值
// 注意：默认文档中有一个根 Map，路径从根 Map 开始，根 Map 的 key 为 DATA_MAP_NAME
// 支持的语法参见：test/query/extract_segment_test.go
func VisitDocByPath(doc *loro.LoroDoc, path string) (any, error) {
	dataMap := doc.GetMap(DATA_MAP_NAME)
	var curr any = dataMap

	segments, err := ExtractSegments(path)
	if err != nil {
		return nil, err
	}

	for _, segment := range segments {
		// 访问不存在的路径返回 PathNotFoundError
		if isNil(curr) {
			return nil, PathNotFoundError
		}

		index, isIndex := segment.(int)
		key, _ := segment.(string)
		switch v := curr.(type) {
		case *loro.LoroMap:
			if isIndex {
				return nil, pe.Wrapf(InvalidPathError, path)
			} else {
				val, err := v.Get(key)
				if err != nil {
					if pe.Is(err, loro.ErrLoroGetNull) {
						return nil, PathNotFoundError
					}
					return nil, err
				}
				curr = val
			}
		case *loro.LoroList:
			if isIndex {
				vlen := int(v.GetLen())
				if index < 0 || index >= vlen {
					return nil, PathNotFoundError
				}
				val, err := v.Get(uint32(index))
				if err != nil {
					return nil, err
				}
				curr = val
			} else {
				return nil, pe.Wrapf(InvalidPathError, path)
			}
		case *loro.LoroMovableList:
			if isIndex {
				vlen := int(v.GetLen())
				if index < 0 || index >= vlen {
					return nil, PathNotFoundError
				}
				val, err := v.Get(uint32(index))
				if err != nil {
					return nil, err
				}
				curr = val
			} else {
				return nil, pe.Wrapf(InvalidPathError, path)
			}
		case map[string]loro.LoroValue:
			if isIndex {
				return nil, pe.Wrapf(InvalidPathError, path)
			} else {
				curr = v[key]
			}
		case []loro.LoroValue:
			if isIndex {
				if index < 0 || index >= len(v) {
					return nil, PathNotFoundError
				}
				curr = v[index]
			} else {
				return nil, pe.Wrapf(InvalidPathError, path)
			}
		}
	}

	return curr, nil
}

// ExtractSegments 将路径字符串解析为段列表
func ExtractSegments(path string) ([]any, error) {
	if path == "" {
		return nil, pe.Wrapf(InvalidPathError, path)
	}

	segments := []interface{}{}
	currentPos := 0

	for currentPos < len(path) {
		// 处理数组索引开头形式 [0]
		if path[currentPos] == '[' {
			closeBracketIndex, err := findClosingBracket(path, currentPos)
			if err != nil {
				return nil, err
			}

			bracketContent := path[currentPos+1 : closeBracketIndex]

			// 处理数字索引 [42]
			digitRegex := regexp.MustCompile(`^\d+$`)
			if digitRegex.MatchString(bracketContent) {
				num, _ := strconv.Atoi(bracketContent)
				segments = append(segments, num)
			} else if (len(bracketContent) >= 2 && bracketContent[0] == '"' && bracketContent[len(bracketContent)-1] == '"') ||
				(len(bracketContent) >= 2 && bracketContent[0] == '\'' && bracketContent[len(bracketContent)-1] == '\'') {
				// 处理字符串属性 ["key"] 或 ['key']
				quoteChar := bracketContent[0]
				content := bracketContent[1 : len(bracketContent)-1]
				unescaped := unescapeString(content, quoteChar)
				segments = append(segments, unescaped)
			} else {
				return nil, pe.Wrapf(InvalidPathError, path)
			}

			currentPos = closeBracketIndex + 1

			// 跳过点号
			if currentPos < len(path) && path[currentPos] == '.' {
				currentPos++
			}
		} else {
			// 处理普通属性名
			var endPos int

			// 查找当前段结束位置（点号或方括号前）
			nextDot := indexOf(path[currentPos:], '.')
			if nextDot != -1 {
				nextDot += currentPos
			}

			nextBracket := indexOf(path[currentPos:], '[')
			if nextBracket != -1 {
				nextBracket += currentPos
			}

			if nextDot == -1 && nextBracket == -1 {
				// 最后一个段
				endPos = len(path)
			} else if nextDot == -1 {
				// 没有点号但有方括号
				endPos = nextBracket
			} else if nextBracket == -1 {
				// 有点号但没有方括号
				endPos = nextDot
			} else {
				// 两者都有，取最近的
				if nextDot < nextBracket {
					endPos = nextDot
				} else {
					endPos = nextBracket
				}
			}

			segment := path[currentPos:endPos]

			// 检查无效路径
			if len(segment) == 0 || (currentPos > 0 && path[currentPos-1] == '.' && len(segment) == 0) {
				return nil, pe.Wrapf(InvalidPathError, path)
			}

			segments = append(segments, segment)
			currentPos = endPos

			// 跳过点号
			if currentPos < len(path) && path[currentPos] == '.' {
				currentPos++
			}
		}
	}

	return segments, nil
}

// indexOf 返回字符在字符串中第一次出现的位置，不存在则返回-1
func indexOf(s string, char byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == char {
			return i
		}
	}
	return -1
}

// findClosingBracket 查找对应的闭合方括号
func findClosingBracket(str string, openBracketPos int) (int, error) {
	if str[openBracketPos] != '[' {
		return -1, pe.Wrapf(InvalidPathError, str)
	}

	depth := 1
	inQuote := false
	var quoteChar byte
	escaped := false

	for i := openBracketPos + 1; i < len(str); i++ {
		// 处理转义字符
		if inQuote && str[i] == '\\' && !escaped {
			escaped = true
			continue
		}

		// 处理引号
		if (str[i] == '"' || str[i] == '\'') && !escaped {
			if !inQuote {
				inQuote = true
				quoteChar = str[i]
			} else if str[i] == quoteChar {
				inQuote = false
			}
		}

		// 只在非引号状态下计算方括号
		if !inQuote {
			if str[i] == '[' {
				depth++
			} else if str[i] == ']' {
				depth--
				if depth == 0 {
					return i, nil
				}
			}
		}

		escaped = false
	}

	return -1, pe.Wrapf(InvalidPathError, str)
}

// unescapeString 处理转义字符
func unescapeString(str string, quoteChar byte) string {
	var result []byte
	escaped := false

	for i := 0; i < len(str); i++ {
		if str[i] == '\\' && !escaped {
			escaped = true
		} else {
			result = append(result, str[i])
			escaped = false
		}
	}

	return string(result)
}
