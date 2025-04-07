package bdd

import (
	"fmt"
	"strconv"
)

type Node struct {
	outermostInstance interface {
		AsNode() *Node
	}
	Id       string
	Deleted  bool
	RootNode *RootNode
	Level    int
}

func (n *Node) IsEqualToOtherNode(otherNode *Node, ownString *string) bool {
	ownStringVal := ""
	if ownString != nil {
		ownStringVal = *ownString
	} else {
		ownStringVal = n.ToString()
	}
	return ownStringVal == otherNode.ToString()
}

func (n *Node) GetParents() *Parents {
	switch n.outermostInstance.(type) {
	case *RootNode:
		panic("root node has no parents")
	case *LeafNode:
		return n.AsLeafNode().Parents
	case *InternalNode:
		return n.AsInternalNode().Parents
	default:
		panic("unknown node type" + fmt.Sprintf("%#v", n))
	}
}

func (n *Node) GetBranches() *Branches {
	switch n.outermostInstance.(type) {
	case *RootNode:
		return n.AsRootNode().Branches
	case *LeafNode:
		panic("leaf node has no branches")
	case *InternalNode:
		return n.AsInternalNode().Branches
	default:
		panic("unknown node type")
	}
}

func (n *Node) Remove() {
	n.EnsureNotDeleted("remove")

	if n.IsInternalNode() {
		internalNode := n.AsInternalNode()
		if internalNode.Parents.Size() > 0 {
			panic("cannot remove internal node with parents")
		}
	}

	if n.IsNonLeafNode() {
		branches := n.GetBranches()
		if branches.AreBranchesStrictEqual() {
			branches.GetBranch("0").GetParents().Remove(n)
		} else {
			branches.GetBranch("0").GetParents().Remove(n)
			branches.GetBranch("1").GetParents().Remove(n)
		}
	}

	n.Deleted = true

	if n.IsNonRootNode() {
		n.RootNode.RemoveNode(n)
	}
}

func (n *Node) ToJson(withId bool) NodeJson {
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
		if n.IsNonRootNode() {
			parents := []string{}
			for parent := range n.GetParents().Parents {
				parents = append(parents, parent.Id)
			}
			ret.Parents = parents
		}
	}

	if n.IsLeafNode() {
		leafNode := n.AsLeafNode()
		ret.Value = &leafNode.Value
	}

	if n.IsNonLeafNode() {
		branches := n.GetBranches()
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

func (n *Node) ToString() string {
	ret := "<" + n.TypeString() + ":" + strconv.Itoa(n.Level)

	if n.IsNonLeafNode() {
		branches := n.GetBranches()
		ret += "|0" + branches.GetBranch("0").ToString()
		ret += "|1" + branches.GetBranch("1").ToString()
	}

	if leafNode, ok := any(n).(*LeafNode); ok {
		ret += "|v" + strconv.Itoa(leafNode.Value)
	}

	return ret + ">"
}

func (n *Node) TypeString() string {
	switch n.outermostInstance.(type) {
	case *RootNode:
		return "RootNode"
	case *LeafNode:
		return "LeafNode"
	case *InternalNode:
		return "InternalNode"
	default:
		return "Unknown"
	}
}

func (n *Node) EnsureNotDeleted(op string) {
	if n.Deleted {
		panic("forbidden operation " + op + " on deleted node " + n.Id)
	}
}

func (n *Node) ApplyEliminationRule(nodesOfSameLevel []*Node) bool {
	n.EnsureNotDeleted("applyEliminationRule")
	if nodesOfSameLevel == nil {
		nodesOfSameLevel = make([]*Node, 0)
		tmp := n.RootNode.GetNodesOfLevel(n.Level)
		for _, node := range tmp {
			nodesOfSameLevel = append(nodesOfSameLevel, node)
		}
	}

	thisNode := n
	other := FindSimilarNode(n, nodesOfSameLevel)

	if other != nil {
		ownParents := thisNode.GetParents().GetAll()

		parentsWithStrictEqualBranches := make([]*NonLeafNode, 0)
		for _, parent := range ownParents {
			branchkey := parent.GetBranches().GetKeyOfNode(thisNode)
			parent.GetBranches().SetBranch(branchkey, other)
			if parent.GetBranches().AreBranchesStrictEqual() {
				parentsWithStrictEqualBranches = append(parentsWithStrictEqualBranches, parent)
			}
			thisNode.GetParents().Remove(parent)
		}

		for _, node := range parentsWithStrictEqualBranches {
			if node.IsInternalNode() {
				node.AsInternalNode().ApplyRuductionRule()
			}
		}

		return true
	} else {
		return false
	}
}

type NodeJson struct {
	Id       *string             `json:"id,omitempty"`
	Deleted  bool                `json:"deleted"`
	Level    int                 `json:"level"`
	Type     string              `json:"type"`
	Parents  []string            `json:"parents,omitempty"`
	Value    *int                `json:"value,omitempty"`
	Branches map[string]NodeJson `json:"branches,omitempty"`
}
