package transpiler

// PropAccess 表示一次属性访问
//
// 例如：obj.method(arg1, arg2) 对应的 PropAccess 为：
//
//	PropAccess{
//		Prop: "method",
//		Args: []any{"arg1", "arg2"},
//		IsCall: true,
//	}
//
// obj.name 对应的 PropAccess 为：
//
//	PropAccess{
//		Prop: "name",
//	}
type PropAccess struct {
	// 属性名
	Prop string
	// 如果是函数调用，这里是参数
	Args []any
	// 是否是函数调用
	IsCall bool
}

type PropAccessHandler func(access PropAccess, obj any) (any, error)
