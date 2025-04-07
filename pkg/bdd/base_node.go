package bdd

import (
	"strconv"
)

type BaseNode struct {
	Id       string
	Deleted  bool
	RootNode *RootNode
	Level    int
}

func NewBaseNode(level int, rootNode *RootNode) *BaseNode {
	return &BaseNode{
		Id:       NextNodeId(),
		Deleted:  false,
		RootNode: rootNode,
		Level:    level,
	}
}

func (n *BaseNode) GetId() string {
	return n.Id
}

func (n *BaseNode) GetLevel() int {
	return n.Level
}

func (n *BaseNode) IsEqualToOtherNode(otherNode Node, ownString *string) bool {
	ownStringVal := ""
	if ownString != nil {
		ownStringVal = *ownString
	} else {
		ownStringVal = n.ToString()
	}
	return ownStringVal == otherNode.ToString()
}

func (n *BaseNode) Remove() {
	n.EnsureNotDeleted("remove")

	if internalNode, ok := any(n).(*InternalNode); ok {
		if internalNode.Parents.Size() > 0 {
			panic("cannot remove internal node with parents")
		}
	}

	if nonLeafNode, ok := any(n).(NonLeafNode); ok {
		branches := nonLeafNode.GetBranches()
		if branches.AreBranchesStrictEqual() {
			branches.GetBranch("0").GetParents().Remove(nonLeafNode)
		} else {
			branches.GetBranch("0").GetParents().Remove(nonLeafNode)
			branches.GetBranch("1").GetParents().Remove(nonLeafNode)
		}
	}

	n.Deleted = true

	if nonRootNode, ok := any(n).(NonRootNode); ok {
		n.RootNode.RemoveNode(nonRootNode)
	}
}

func (n *BaseNode) ToJson(withId bool) NodeJson {
	var id *string = nil
	if withId {
		id = &n.Id
	}

	ret := NodeJson{
		Id:      id,
		Deleted: n.Deleted,
		Type:    n.TypeString(),
		Level:   n.Level,
	}

	if withId {
		if nonRootNode, ok := any(n).(NonRootNode); ok {
			parents := []NodeJson{}
			for parent := range nonRootNode.GetParents().Parents {
				parents = append(parents, parent.ToJson(withId))
			}
			ret.Parents = parents
		}
	}

	if leafNode, ok := any(n).(*LeafNode); ok {
		ret.Value = &leafNode.Value
	}

	if nonLeafNode, ok := any(n).(NonLeafNode); ok {
		branches := nonLeafNode.GetBranches()
		if !branches.Deleted {
			branch0Json := branches.GetBranch("0").ToJson(withId)
			branch1Json := branches.GetBranch("1").ToJson(withId)
			ret.Branches = map[string]NodeJson{
				"0": branch0Json,
				"1": branch1Json,
			}
		}
	}

	return ret
}

func (n *BaseNode) ToString() string {
	ret := "<" + n.TypeString() + ":" + strconv.Itoa(n.Level)

	if nonLeafNode, ok := any(n).(NonLeafNode); ok {
		branches := nonLeafNode.GetBranches()
		ret += "|0" + branches.GetBranch("0").ToString()
		ret += "|1" + branches.GetBranch("1").ToString()
	}

	if leafNode, ok := any(n).(*LeafNode); ok {
		ret += "|v" + strconv.Itoa(leafNode.Value)
	}

	return ret + ">"
}

func (n *BaseNode) TypeString() string {
	switch any(n).(type) {
	case *InternalNode:
		return "InternalNode"
	case *RootNode:
		return "RootNode"
	case *LeafNode:
		return "LeafNode"
	default:
		return "Unknown"
	}
}

func (n *BaseNode) EnsureNotDeleted(op string) {
	if n.Deleted {
		panic("forbidden operation " + op + " on deleted node " + n.Id)
	}
}

func (n *BaseNode) ApplyEliminationRule(nodesOfSameLevel []Node) bool {
	n.EnsureNotDeleted("applyEliminationRule")
	if nodesOfSameLevel == nil {
		nodesOfSameLevel = make([]Node, 0)
		tmp := n.RootNode.GetNodesOfLevel(n.Level)
		for _, node := range tmp {
			nodesOfSameLevel = append(nodesOfSameLevel, node)
		}
	}

	thisNode := any(n).(Node)
	other := FindSimilarNode(thisNode, nodesOfSameLevel)

	if other != nil {
		thisNode := any(n).(NonRootNode)
		other := any(other).(NonRootNode)
		ownParents := thisNode.GetParents().GetAll()

		parentsWithStrictEqualBranches := make([]NonLeafNode, 0)
		for _, parent := range ownParents {
			branchkey := parent.GetBranches().GetKeyOfNode(thisNode)
			parent.GetBranches().SetBranch(branchkey, other)
			if parent.GetBranches().AreBranchesStrictEqual() {
				parentsWithStrictEqualBranches = append(parentsWithStrictEqualBranches, parent)
			}
			thisNode.GetParents().Remove(parent)
		}

		for _, node := range parentsWithStrictEqualBranches {
			if internalNode, ok := any(node).(*InternalNode); ok {
				internalNode.ApplyRuductionRule()
			}
		}

		return true
	} else {
		return false
	}
}

func (n *BaseNode) IsDeleted() bool {
	return n.Deleted
}
