package bdd

type NonRootNode = Node

func (n *Node) IsNonRootNode() bool {
	return !n.IsRootNode()
}
