package types

// MongoQuery 表示MongoDB风格的查询参数
type MongoQuery struct {
	// 查询选择器
	Selector map[string]interface{} `json:"selector"`
	// 跳过的记录数
	Skip *int `json:"skip,omitempty"`
	// 返回的最大记录数
	Limit *int `json:"limit,omitempty"`
	// 排序字段，排序必须是可预测的
	Sort []string `json:"sort"`
}
