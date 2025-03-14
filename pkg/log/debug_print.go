package log

import "fmt"

type DebugPrintable interface {
	DebugPrint() string
}

func DebugPrint(v any) string {
	if dp, ok := v.(DebugPrintable); ok {
		return dp.DebugPrint()
	} else {
		return fmt.Sprintf("%v", v)
	}
}
