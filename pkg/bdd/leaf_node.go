package bdd

type LeafNode struct {
	*BaseNode
	Parents *Parents
	Value   int
}

func NewLeafNode(level int, rootNode *RootNode, value int, parent NonLeafNode) *LeafNode {
	node := &LeafNode{
		BaseNode: NewBaseNode(level, rootNode),
		Parents:  nil, // 下面赋值
		Value:    value,
	}
	node.Parents = NewParents(node)
	node.Parents.Add(parent)

	return node
}

func (n *LeafNode) RemoveIfValueEquals(value int) bool {
	panic("not implemented")
}
