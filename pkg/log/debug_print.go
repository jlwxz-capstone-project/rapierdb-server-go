package log

import "fmt"

type DebugPrintable interface {
	DebugSprint() string
}

func DebugSprint(v any) string {
	if dp, ok := v.(DebugPrintable); ok {
		return dp.DebugSprint()
	} else {
		return fmt.Sprintf("%v", v)
	}
}
