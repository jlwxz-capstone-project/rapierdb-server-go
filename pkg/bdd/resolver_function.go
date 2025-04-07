package bdd

// GetResolverFunctions 获取状态解析函数
type ResolverFunction func(state string) bool
type ResolverFunctions map[int]ResolverFunction

func GetResolverFunctions(size int, log bool) ResolverFunctions {
	resolvers := make(ResolverFunctions)
	for i := 0; i < size; i++ {
		index := i // 捕获变量
		fn := func(state string) bool {
			if index >= len(state) {
				return false
			}
			ret := state[index] == '1'
			if log {
				println("调用解析函数，index=", index, "返回值=", ret)
			}
			return ret
		}
		resolvers[i] = fn
	}
	return resolvers
}
