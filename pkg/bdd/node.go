package bdd

type Node interface {
	isNode()
}

type NonLeafNode interface {
	isNonLeafNode()
}

func (n *InternalNode) isNonLeafNode() {}
func (n *RootNode) isNonLeafNode()     {}

type NonRootNode interface {
	isNonRootNode()
}

func (n *InternalNode) isNonRootNode() {}
func (n *LeafNode) isNonRootNode()     {}
