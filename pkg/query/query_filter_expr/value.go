package query_filter_expr

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// ValueExpr 表示一个值
//
//	  type Value =
//		  | bool
//		  | int64
//		  | float64
//		  | string
//		  | []Value
//		  | map[string]Value
//		  | nil
type ValueExpr struct {
	Value any
}

func (e *ValueExpr) DebugPrint() string {
	return fmt.Sprintf("ValueExpr{Value: %v}", e.Value)
}

func (e *ValueExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	return e, nil
}

func (e *ValueExpr) MarshalJSON() ([]byte, error) {
	// 对于整数类型，确保使用 int64
	switch v := e.Value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return json.Marshal(SerializedQueryFilterExpr{
			Type:  ExprTypeValue,
			Value: util.ToInt64(v),
		})
	default:
		return json.Marshal(SerializedQueryFilterExpr{
			Type:  ExprTypeValue,
			Value: e.Value,
		})
	}
}

func (e *ValueExpr) UnmarshalJSON(data []byte) error {
	var s SerializedQueryFilterExpr
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s.Type != ExprTypeValue {
		return fmt.Errorf("expected value expression, got %s", s.Type)
	}

	// 处理数值类型
	if num, ok := s.Value.(float64); ok {
		// 检查是否是整数
		if num == float64(int64(num)) {
			e.Value = int64(num)
		} else {
			e.Value = num
		}
	} else {
		e.Value = s.Value
	}
	return nil
}
