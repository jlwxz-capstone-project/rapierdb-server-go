package bdd

type NonLeafNode = Node

func (n *Node) IsNonLeafNode() bool {
	return !n.IsLeafNode()
}
