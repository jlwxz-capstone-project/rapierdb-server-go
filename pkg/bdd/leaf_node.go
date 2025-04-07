package bdd

import "fmt"

type LeafNode struct {
	*BaseNode
	Parents *Parents
	Value   int
}

func NewLeafNode(level int, rootNode *RootNode, value int, parent *NonLeafNode) *LeafNode {
	ret := &LeafNode{
		BaseNode: NewBaseNode(level, rootNode),
		Parents:  nil, // 下面赋值
		Value:    value,
	}
	ret.outermostInstance = ret
	ret.Parents = NewParents(ret.AsNode())
	ret.Parents.Add(parent)
	fmt.Println("NewLeafNode", ret.Id)
	return ret
}

func (n *Node) IsLeafNode() bool      { _, ok := n.outermostInstance.(*LeafNode); return ok }
func (n *Node) AsLeafNode() *LeafNode { return n.outermostInstance.(*LeafNode) }

func (n *LeafNode) RemoveIfValueEquals(value int) bool {
	n.EnsureNotDeleted("removeIfValueEquals")

	if n.Value != value {
		return false
	}

	parents := n.Parents.GetAll()
	for _, parent := range parents {
		branchKey := parent.GetBranches().GetKeyOfNode(n.AsNode())
		oppositeBranchKey := string(OppositeBoolean(BooleanString(branchKey)))
		otherBranch := parent.GetBranches().GetBranch(oppositeBranchKey)
		n.Parents.Remove(parent)
		parent.GetBranches().SetBranch(branchKey, otherBranch)
		if parent.IsInternalNode() {
			internalNode := parent.AsInternalNode()
			internalNode.ApplyRuductionRule()
		}
	}

	return true
}
