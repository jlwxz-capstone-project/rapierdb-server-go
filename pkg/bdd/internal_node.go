package bdd

type InternalNode struct {
	*BaseNode
	Branches *Branches
	Parents  *Parents
}

func NewInternalNode(level int, rootNode *RootNode, parent NonLeafNode) *InternalNode {
	node := &InternalNode{
		BaseNode: NewBaseNode(level, rootNode),
	}
	node.Branches = NewBranches(node)
	node.Parents = NewParents(node)
	node.Parents.Add(parent)

	return node
}

func (n *InternalNode) ApplyRuductionRule() bool {
	if n.Branches.HasEqualBranches() {
		n.EnsureNotDeleted("applyRuductionRule")
		keepBranch := n.Branches.GetBranch("0")

		ownParents := n.GetParents().GetAll()
		for _, parent := range ownParents {
			branchkey := parent.GetBranches().GetKeyOfNode(n)
			parent.GetBranches().SetBranch(branchkey, keepBranch)

			n.Parents.Remove(parent)

			if parent, ok := parent.(*InternalNode); ok {
				if parent.Branches.AreBranchesStrictEqual() {
					parent.ApplyRuductionRule()
				}
			}
		}

		return true
	}
	return false
}
