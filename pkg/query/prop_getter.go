package query

import (
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/js2go_transpiler/transpiler"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/pkg/errors"
)

func LoroDocAccessHandler(access transpiler.PropAccess, obj any) (any, error) {
	if doc, ok := obj.(*loro.LoroDoc); ok {
		if !access.IsCall {
			if prop, ok := access.Prop.(string); ok {
				a, err := doc.GetByPath(prop).Unwrap()
				if err != nil {
					return nil, err
				}
				return a, nil
			}
			return nil, transpiler.ErrPropNotSupport
		}
	}
	return nil, transpiler.ErrPropNotSupport
}

// LoroTextAccessHandler 处理 LoroText 的属性访问
//
//	text.length // 返回文本对应字符串的长度
//	text.toString() // 转为字符串
//	text[index] // 返回字符串的第 index 个字符（自动转为字符串）
func LoroTextAccessHandler(access transpiler.PropAccess, obj any) (any, error) {
	if text, ok := obj.(*loro.LoroText); ok {
		switch prop := access.Prop.(type) {
		case string:
			if !access.IsCall {
				switch prop {
				case "length":
					s, err := text.ToString()
					if err != nil {
						return nil, err
					}
					return len(s), nil
				}
			} else {
				if access.Prop == "toString" {
					s, err := text.ToString()
					if err != nil {
						return nil, err
					}
					return s, nil
				}
			}
		case int:
			index := prop
			s, err := text.ToString()
			if err != nil {
				return nil, err
			}
			if index < 0 || index >= len(s) {
				return nil, errors.WithStack(fmt.Errorf("index out of range: %d", index))
			}
			return string(s[index]), nil
		}
	}
	return nil, transpiler.ErrPropNotSupport
}

// LoroMapAccessHandler 处理 LoroMap 的属性访问
func LoroMapAccessHandler(access transpiler.PropAccess, obj any) (any, error) {
	if lm, ok := obj.(*loro.LoroMap); ok {
		if !access.IsCall {
			if prop, ok := access.Prop.(string); ok {
				a, err := lm.Get(prop).Unwrap()
				if err != nil {
					return nil, err
				}
				return a, nil
			}
		}
	}
	return nil, transpiler.ErrPropNotSupport
}

func LoroListAccessHandler(access transpiler.PropAccess, obj any) (any, error) {
	if ll, ok := obj.(*loro.LoroList); ok {
		if !access.IsCall {
			switch prop := access.Prop.(type) {
			case string:
				if prop == "length" {
					return int(ll.GetLen()), nil
				}
			case int:
				len := ll.GetLen()
				if prop < 0 || prop >= int(len) {
					return nil, errors.WithStack(fmt.Errorf("index out of range: %d", prop))
				}
				a, err := ll.Get(uint32(prop)).Unwrap()
				if err != nil {
					return nil, err
				}
				return a, nil
			}
		} else {
			// if prop, ok := access.Prop.(string); ok {
			// 	if prop == "push" {
			// 		if len(access.Args) != 1 {
			// 			return nil, errors.WithStack(fmt.Errorf("push method requires 1 argument"))
			// 		}
			// 		arg0 := access.Args[0]
			// 		return ll.Push(arg0)
			// 	}
			// }
		}
	}
	return nil, transpiler.ErrPropNotSupport
}

func LoroMovableListAccessHandler(access transpiler.PropAccess, obj any) (any, error) {
	if lm, ok := obj.(*loro.LoroMovableList); ok {
		if !access.IsCall {
			switch prop := access.Prop.(type) {
			case string:
				if prop == "length" {
					return int(lm.GetLen()), nil
				}
			case int:
				len := lm.GetLen()
				if prop < 0 || prop >= int(len) {
					return nil, errors.WithStack(fmt.Errorf("index out of range: %d", prop))
				}
				a, err := lm.Get(len).Unwrap()
				if err != nil {
					return nil, err
				}
				return a, nil
			}
		}
	}
	return nil, transpiler.ErrPropNotSupport
}
