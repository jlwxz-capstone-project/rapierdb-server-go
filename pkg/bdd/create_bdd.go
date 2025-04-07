package bdd

// CreateBddFromTruthTable 从真值表创建二进制决策图
func CreateBddFromTruthTable(truthTable TruthTable) *RootNode {
	root := NewRootNode()

	// 获取第一个键并检查所有键长度是否一致
	var firstKey string
	for k := range truthTable {
		firstKey = k
		break
	}

	if firstKey == "" {
		return root
	}

	keyLength := len(firstKey)
	mustBeSize := 1 << keyLength // 2^keyLength

	if len(truthTable) != mustBeSize {
		panic("truth table has missing entries")
	}

	// 为真值表中的每个状态创建节点
	for stateSet, value := range truthTable {
		var lastNode NonLeafNode = root

		// 遍历状态的每个字符
		for i := 0; i < (len(stateSet) - 1); i++ {
			level := i + 1
			state := string(stateSet[i])

			// 如果这个状态字符的节点不存在，添加一个新的
			if lastNode.GetBranches().GetBranch(state) == nil {
				newNode := NewInternalNode(level, root, lastNode)
				lastNode.GetBranches().SetBranch(state, newNode)
			}

			lastNode = lastNode.GetBranches().GetBranch(state).(*InternalNode)
		}

		// 最后一个节点是叶节点
		lastState := LastChar(stateSet)
		if lastNode.GetBranches().GetBranch(lastState) != nil {
			panic("leafNode already exists, this should not happen")
		}

		leafNode := NewLeafNode(len(stateSet), root, value, lastNode)
		lastNode.GetBranches().SetBranch(lastState, leafNode)
	}

	return root
}
