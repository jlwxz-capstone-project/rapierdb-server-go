package bdd

// CreateBddFromTruthTable 从真值表创建二进制决策图
func CreateBddFromTruthTable(truthTable TruthTable) *RootNode {
	root := NewRootNode()

	var firstKey string
	for k := range truthTable {
		firstKey = k
		break
	}

	keyLength := len(firstKey)
	mustBeSize := 1 << keyLength // 2^keyLength

	if len(truthTable) != mustBeSize {
		panic("truth table has missing entries")
	}

	for stateSet, value := range truthTable {
		var lastNode *NonLeafNode = root.AsNode()

		for i := 0; i < (len(stateSet) - 1); i++ {
			level := i + 1
			state := string(stateSet[i])

			if lastNode.GetBranches().GetBranch(state) == nil {
				newNode := NewInternalNode(level, root, lastNode)
				lastNode.GetBranches().SetBranch(state, newNode.AsNode())
			}

			lastNode = lastNode.GetBranches().GetBranch(state)
		}

		lastState := LastChar(stateSet)
		if lastNode.GetBranches().GetBranch(lastState) != nil {
			panic("leafNode already exists, this should not happen")
		}

		leafNode := NewLeafNode(len(stateSet), root, value, lastNode)
		lastNode.GetBranches().SetBranch(lastState, leafNode.AsNode())
	}

	// fmt.Println(root.ToJson(true).ToString())

	return root
}
