package transpiler

// Scope 定义转译上下文
type Scope struct {
	// 变量和函数映射表，让 js 代码可以访问这些变量和函数
	Vars map[string]any
	// 属性访问器
	PropGetter func(chain []PropAccess, obj any) (any, error)
	// 属性赋值器
	PropMutator PropMutator
	// 父级上下文
	Parent *Scope
}

// NewScope 创建新的作用域
func NewScope(parent *Scope, propGetter PropGetter, propMutator PropMutator) *Scope {
	if propGetter == nil {
		propGetter = DefaultPropGetter
	}
	if propMutator == nil {
		propMutator = DefaultPropMutator
	}
	return &Scope{
		Vars:        make(map[string]any),
		Parent:      parent, // 保留父级引用
		PropGetter:  propGetter,
		PropMutator: propMutator,
	}
}

// GetVar 变量查找当前作用域内的变量，会向上追溯
func (ctx *Scope) GetVar(name string) (any, bool) {
	current := ctx
	for current != nil {
		if val, ok := current.Vars[name]; ok {
			return val, true
		}
		current = current.Parent
	}
	return nil, false
}
