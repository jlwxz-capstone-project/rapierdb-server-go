package query_filter_expr

import "errors"

var (
	ErrTypeError   = errors.New("type error")
	ErrFieldError  = errors.New("field error")
	ErrEvalError   = errors.New("eval error")
	ErrSyntaxError = errors.New("syntax error")
)
