package bdd

type BaseNode struct {
	Node
}

func (n *BaseNode) AsNode() *Node {
	return &n.Node
}

func NewBaseNode(level int, rootNode *RootNode) *BaseNode {
	node := &BaseNode{
		Node: Node{
			outermostInstance: nil, // 需要在子类构造函数中设置
			Id:                NextNodeId(),
			Deleted:           false,
			RootNode:          rootNode,
			Level:             level,
		},
	}
	if rootNode != nil {
		rootNode.AddNode(node.AsNode())
	}
	return node
}
