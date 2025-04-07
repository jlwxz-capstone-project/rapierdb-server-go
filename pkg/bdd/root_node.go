package bdd

import (
	"sort"
)

type RootNode struct {
	*BaseNode
	Branches     *Branches
	Levels       []int
	NodesByLevel map[int]map[Node]struct{}
	RootNode     *RootNode
}

func NewRootNode() *RootNode {
	ret := &RootNode{
		BaseNode:     NewBaseNode(0, nil),
		Branches:     nil, // 下面赋值
		Levels:       []int{},
		NodesByLevel: map[int]map[Node]struct{}{},
	}
	ret.Branches = NewBranches(ret)
	ret.Levels = append(ret.Levels, 0)
	level0Set := map[Node]struct{}{}
	level0Set[ret] = struct{}{}
	ret.NodesByLevel[0] = level0Set
	ret.RootNode = ret // 根节点的RootNode指向自己
	return ret
}

func (n *RootNode) AddNode(node NonRootNode) {
	level := node.GetLevel()

	contains := false
	for _, l := range n.Levels {
		if l == level {
			contains = true
			break
		}
	}

	if !contains {
		n.Levels = append(n.Levels, level)
	}

	n.ensureLevelSetExists(level)
	n.NodesByLevel[level][node] = struct{}{}
}

func (n *RootNode) RemoveNode(node NonRootNode) {
	level := node.GetLevel()
	set := n.NodesByLevel[level]
	if set != nil {
		delete(set, node)
	}
}

func (n *RootNode) GetSortedLevels() []int {
	ret := make([]int, len(n.Levels))
	copy(ret, n.Levels)
	sort.Ints(ret)
	return ret
}

func (n *RootNode) GetNodesOfLevel(level int) []NonRootNode {
	set := n.NodesByLevel[level]
	if set == nil {
		return []NonRootNode{}
	}

	ret := make([]NonRootNode, 0, len(set))
	for node := range set {
		if node != n { // 跳过根节点自身
			ret = append(ret, node.(NonRootNode))
		}
	}
	return ret
}

func (n *RootNode) CountNodes() int {
	ret := 0
	for _, set := range n.NodesByLevel {
		ret += len(set)
	}
	return ret
}

func (n *RootNode) Minimize() {
	done := false

	for !done {
		successCount := 0
		levels := n.GetSortedLevels()
		lastLevel := levels[len(levels)-1]
		for lastLevel > 0 {
			nodes := n.GetNodesOfLevel(lastLevel)
			nodeCount := 0
			for _, node := range nodes {
				nodeCount++
				if leafNode, ok := node.(*LeafNode); ok {
					reductionDone := leafNode.ApplyEliminationRule(nil)
					if reductionDone {
						successCount++
					}
				}
				if useNode, ok := node.(*InternalNode); ok {
					if !useNode.IsDeleted() {
						reductionDone := useNode.ApplyRuductionRule()
						eliminationDone := false
						if !useNode.IsDeleted() {
							nodes2 := make([]Node, 0)
							for _, node := range nodes {
								nodes2 = append(nodes2, node)
							}
							eliminationDone = useNode.ApplyEliminationRule(nodes2)
						}
						if reductionDone || eliminationDone {
							successCount++
						}
					}
				}
			}
			lastLevel--
		}
		if successCount == 0 {
			done = true // could do no more optimizations
		} else {
			// logging
		}
	}
}

func (n *RootNode) GetLeafNodes() []*LeafNode {
	maxLevel := -1
	for _, level := range n.Levels {
		if level > maxLevel {
			maxLevel = level
		}
	}

	if maxLevel == -1 {
		return []*LeafNode{}
	}

	set := n.NodesByLevel[maxLevel]
	if set == nil {
		return []*LeafNode{}
	}

	ret := make([]*LeafNode, 0, len(set))
	for node := range set {
		if leaf, ok := node.(*LeafNode); ok {
			ret = append(ret, leaf)
		}
	}

	return ret
}

func (n *RootNode) RemoveIrrelevantLeafNodes(leafNodeValue int) {
	done := false
	for !done {
		countRemoved := 0
		leafNodes := n.GetLeafNodes()
		for _, leafNode := range leafNodes {
			removed := leafNode.RemoveIfValueEquals(leafNodeValue)
			if removed {
				countRemoved++
			}
		}

		n.Minimize()

		if countRemoved == 0 {
			done = true
		}
	}
}

func (n *RootNode) Resolve(fns ResolverFunctions, booleanFunctionInput string) int {
	var currentNode Node = n
	for {
		booleanResult := fns[currentNode.GetLevel()](booleanFunctionInput)
		branchKey := string(BooleanToBooleanString(booleanResult))
		if nonLeafNode, ok := any(currentNode).(NonLeafNode); ok {
			currentNode = nonLeafNode.GetBranches().GetBranch(branchKey)
			if leafNode, ok := currentNode.(*LeafNode); ok {
				return leafNode.Value
			}
		} else {
			panic("IMPOSSIBLE")
		}
	}
}

func (n *RootNode) ensureLevelSetExists(level int) {
	if _, ok := n.NodesByLevel[level]; !ok {
		n.NodesByLevel[level] = map[Node]struct{}{}
	}
}
