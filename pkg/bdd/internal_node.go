package bdd

import "fmt"

type InternalNode struct {
	*BaseNode
	Branches *Branches
	Parents  *Parents
}

func NewInternalNode(level int, rootNode *RootNode, parent *NonLeafNode) *InternalNode {
	ret := &InternalNode{
		BaseNode: NewBaseNode(level, rootNode),
	}
	ret.outermostInstance = ret
	ret.Branches = NewBranches(ret.AsNode())
	ret.Parents = NewParents(ret.AsNode())
	ret.Parents.Add(parent)
	fmt.Println("NewInternalNode", ret.Id)
	return ret
}

func (n *Node) IsInternalNode() bool          { _, ok := n.outermostInstance.(*InternalNode); return ok }
func (n *Node) AsInternalNode() *InternalNode { return n.outermostInstance.(*InternalNode) }

func (n *InternalNode) ApplyRuductionRule() bool {
	if n.Branches.HasEqualBranches() {
		n.EnsureNotDeleted("applyRuductionRule")
		keepBranch := n.Branches.GetBranch("0")

		ownParents := n.GetParents().GetAll()
		for _, parent := range ownParents {
			branchkey := parent.GetBranches().GetKeyOfNode(n.AsNode())
			parent.GetBranches().SetBranch(branchkey, keepBranch)

			n.Parents.Remove(parent)

			if parent.IsInternalNode() {
				internalNode := parent.AsInternalNode()
				if internalNode.Branches.AreBranchesStrictEqual() {
					internalNode.ApplyRuductionRule()
				}
			}
		}

		return true
	}
	return false
}
