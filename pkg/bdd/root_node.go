package bdd

import (
	"sort"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/orderedmap"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/orderedset"
)

type RootNode struct {
	*BaseNode
	Branches *Branches
	Levels   []int
	// NodesByLevel map[int]map[*Node]struct{}
	NodesByLevel *orderedmap.OrderedMap[int, *orderedset.OrderedSet[*Node]]
	RootNode     *RootNode
}

func NewRootNode() *RootNode {
	ret := &RootNode{
		BaseNode:     NewBaseNode(0, nil),
		Branches:     nil, // 下面赋值
		Levels:       []int{},
		NodesByLevel: orderedmap.NewOrderedMap[int, *orderedset.OrderedSet[*Node]](),
	}
	ret.outermostInstance = ret
	ret.Branches = NewBranches(ret.AsNode())
	ret.Levels = append(ret.Levels, 0)
	level0Set := orderedset.NewOrderedSet[*Node]()
	level0Set.Add(ret.AsNode())
	ret.NodesByLevel.Set(0, level0Set)
	ret.RootNode = ret // 根节点的RootNode指向自己
	return ret
}

func (n *Node) IsRootNode() bool      { _, ok := n.outermostInstance.(*RootNode); return ok }
func (n *Node) AsRootNode() *RootNode { return n.outermostInstance.(*RootNode) }

func (n *RootNode) AddNode(node *NonRootNode) {
	level := node.Level

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
	n.NodesByLevel.MustGet(level).Add(node)
}

func (n *RootNode) RemoveNode(node *NonRootNode) {
	level := node.Level
	set, ok := n.NodesByLevel.Get(level)
	if ok {
		set.Remove(node)
	}
}

func (n *RootNode) GetSortedLevels() []int {
	ret := make([]int, len(n.Levels))
	copy(ret, n.Levels)
	sort.Ints(ret)
	return ret
}

func (n *RootNode) GetNodesOfLevel(level int) []*NonRootNode {
	n.ensureLevelSetExists(level)
	set := n.NodesByLevel.MustGet(level)
	ret := make([]*NonRootNode, 0, set.Len())
	for node := range set.IterValues() {
		ret = append(ret, node)
	}
	// 按照 id 从小到大排序，保证和 Js 实现一致
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Id < ret[j].Id
	})
	return ret
}

func (n *RootNode) CountNodes() int {
	ret := 0
	for set := range n.NodesByLevel.IterValues() {
		ret += set.Len()
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
				if node.IsLeafNode() {
					reductionDone := node.AsLeafNode().ApplyEliminationRule(nil)
					if reductionDone {
						successCount++
					}
				}
				if !node.Deleted && node.IsInternalNode() {
					useNode := node.AsInternalNode()
					reductionDone := useNode.ApplyRuductionRule()
					eliminationDone := false
					if !useNode.Deleted {
						eliminationDone = useNode.ApplyEliminationRule(nodes)
					}
					if reductionDone || eliminationDone {
						successCount++
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

	set, ok := n.NodesByLevel.Get(maxLevel)
	if !ok {
		return []*LeafNode{}
	}

	ret := make([]*LeafNode, 0, set.Len())
	for node := range set.IterValues() {
		if node.IsLeafNode() {
			leaf := node.AsLeafNode()
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
	var currentNode *Node = n.AsNode()
	for {
		booleanResult := fns[currentNode.Level](booleanFunctionInput)
		branchKey := string(BooleanToBooleanString(booleanResult))
		if currentNode.IsNonLeafNode() {
			currentNode = currentNode.GetBranches().GetBranch(branchKey)
			if currentNode.IsLeafNode() {
				leafNode := currentNode.AsLeafNode()
				return leafNode.Value
			}
		} else {
			panic("IMPOSSIBLE")
		}
	}
}

func (n *RootNode) ensureLevelSetExists(level int) {
	set, ok := n.NodesByLevel.Get(level)
	if !ok {
		set = orderedset.NewOrderedSet[*Node]()
		n.NodesByLevel.Set(level, set)
	}
}
