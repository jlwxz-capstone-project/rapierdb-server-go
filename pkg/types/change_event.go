package types

// WriteOperation 表示写入操作类型
type WriteOperation string

const (
	// 操作类型常量
	OperationInsert WriteOperation = "INSERT"
	OperationUpdate WriteOperation = "UPDATE"
	OperationDelete WriteOperation = "DELETE"
)

// ChangeEventBase 表示变更事件的基本结构
type ChangeEventBase struct {
	// 操作类型
	Operation WriteOperation `json:"operation"`
	// 文档ID，主键的值
	ID string `json:"id"`
}

// ChangeEventInsert 表示插入操作的变更事件
type ChangeEventInsert struct {
	ChangeEventBase
	Doc      map[string]interface{} `json:"doc"`
	Previous interface{}            `json:"previous"` // 插入操作时，Previous 为 nil
}

// ChangeEventUpdate 表示更新操作的变更事件
type ChangeEventUpdate struct {
	ChangeEventBase
	Doc      map[string]interface{} `json:"doc"`
	Previous map[string]interface{} `json:"previous"`
}

// ChangeEventDelete 表示删除操作的变更事件
type ChangeEventDelete struct {
	ChangeEventBase
	Doc      interface{}            `json:"doc"` // 删除操作时，Doc 为 nil
	Previous map[string]interface{} `json:"previous"`
}

// ChangeEvent 表示任意类型的变更事件
type ChangeEvent struct {
	// 操作类型
	Operation WriteOperation `json:"operation"`
	// 文档ID，主键的值
	ID string `json:"id"`
	// 当前文档，对于删除操作可能为 nil
	Doc map[string]interface{} `json:"doc"`
	// 之前的文档，对于插入操作可能为 nil
	Previous map[string]interface{} `json:"previous"`
}

// NewInsertEvent 创建插入事件
func NewInsertEvent(id string, doc map[string]interface{}) ChangeEvent {
	return ChangeEvent{
		Operation: OperationInsert,
		ID:        id,
		Doc:       doc,
		Previous:  nil,
	}
}

// NewUpdateEvent 创建更新事件
func NewUpdateEvent(id string, doc, previous map[string]interface{}) ChangeEvent {
	return ChangeEvent{
		Operation: OperationUpdate,
		ID:        id,
		Doc:       doc,
		Previous:  previous,
	}
}

// NewDeleteEvent 创建删除事件
func NewDeleteEvent(id string, previous map[string]interface{}) ChangeEvent {
	return ChangeEvent{
		Operation: OperationDelete,
		ID:        id,
		Doc:       nil,
		Previous:  previous,
	}
}
