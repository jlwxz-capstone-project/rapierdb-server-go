package bdd

func FindSimilarNode(own *Node, others []*Node) *Node {
	ownString := own.ToString()
	for _, other := range others {
		if own != other && !other.Deleted && own.IsEqualToOtherNode(other, &ownString) {
			return other
		}
	}
	return nil
}
