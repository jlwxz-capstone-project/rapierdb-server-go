package bdd

type Branches struct {
	Deleted  bool
	Branches map[string]NonRootNode
	Node     NonLeafNode
}

func (b *Branches) SetBranch(which string, branchNode NonRootNode) {
	prev := b.Branches[which]
	if prev == branchNode {
		return
	}

	b.Branches[which] = branchNode

}
