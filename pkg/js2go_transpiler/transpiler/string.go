package transpiler

import (
	"fmt"
	"strings"
)

func StringPropAccessHandler(access PropAccess, obj any) (any, error) {
	if str, ok := obj.(string); ok {
		switch access.Prop {
		case "length":
			return len(str), nil

		case "toLowerCase":
			if access.IsCall {
				return strings.ToLower(str), nil
			}
			return strings.ToLower, nil

		case "toUpperCase":
			if access.IsCall {
				return strings.ToUpper(str), nil
			}
			return strings.ToUpper, nil

		case "trim":
			if access.IsCall {
				return strings.TrimSpace(str), nil
			}
			return strings.TrimSpace, nil

		case "substring":
			if access.IsCall {
				if len(access.Args) < 1 {
					return nil, fmt.Errorf("%w: substring method requires 1 argument", ErrInCall)
				}
				start, ok := access.Args[0].(int)
				if !ok {
					return nil, fmt.Errorf("%w: first argument of substring must be a number", ErrInCall)
				}
				if start < 0 {
					start = 0
				}

				if len(access.Args) > 1 {
					end, ok := access.Args[1].(int)
					if !ok {
						return nil, fmt.Errorf("%w: second argument of substring must be a number", ErrInCall)
					}
					if end > len(str) {
						end = len(str)
					}
					obj = str[start:end]
				} else {
					obj = str[start:]
				}
				return obj, nil
			}

		case "indexOf":
			// 查找子串位置
			if access.IsCall {
				if len(access.Args) < 1 {
					return nil, fmt.Errorf("%w: indexOf method requires 1 argument", ErrInCall)
				}
				substr, ok := access.Args[0].(string)
				if !ok {
					return nil, fmt.Errorf("%w: argument of indexOf must be a string", ErrInCall)
				}
				return strings.Index(str, substr), nil
			}

		case "replace":
			// 替换字符串
			if access.IsCall {
				if len(access.Args) < 2 {
					return nil, fmt.Errorf("%w: replace method requires 2 arguments", ErrInCall)
				}
				old, ok1 := access.Args[0].(string)
				new, ok2 := access.Args[1].(string)
				if !ok1 || !ok2 {
					return nil, fmt.Errorf("%w: arguments of replace must be strings", ErrInCall)
				}
				return strings.Replace(str, old, new, 1), nil
			}
		}
	}
	return nil, ErrPropNotSupport
}
