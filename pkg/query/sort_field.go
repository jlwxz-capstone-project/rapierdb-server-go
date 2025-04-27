package query

import "fmt"

// SortOrder 表示排序顺序
type SortOrder int

const (
	SortOrderAsc  SortOrder = 1  // 升序
	SortOrderDesc SortOrder = -1 // 降序
)

// SortField 表示排序字段
type SortField struct {
	Field string    `json:"field"` // 字段路径
	Order SortOrder `json:"order"` // 排序顺序
}

func (s *SortField) DebugSprint() string {
	orderStr := "asc"
	if s.Order == SortOrderDesc {
		orderStr = "desc"
	}
	return fmt.Sprintf("SortField{Field: %s, Order: %s}", s.Field, orderStr)
}
