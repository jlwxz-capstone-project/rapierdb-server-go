package query_filter_expr

import "fmt"

var (
	ErrTypeError   = fmt.Errorf("type error")
	ErrFieldError  = fmt.Errorf("field error")
	ErrEvalError   = fmt.Errorf("eval error")
	ErrSyntaxError = fmt.Errorf("syntax error")
)
