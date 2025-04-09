package bdd

import (
	"bytes"
	"encoding/json"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/orderedmap"
)

type Branches struct {
	Deleted bool
	// Branches map[string]*NonRootNode
	Branches *orderedmap.OrderedMap[string, *NonRootNode]
	Node     *NonLeafNode
}

func NewBranches(node *NonLeafNode) *Branches {
	return &Branches{
		Deleted:  false,
		Branches: orderedmap.NewOrderedMap[string, *NonRootNode](),
		Node:     node,
	}
}

func (b *Branches) SetBranch(which string, branchNode *NonRootNode) {
	previous, ok := b.Branches.Get(which)
	if ok && previous == branchNode {
		return
	}

	b.Branches.Set(which, branchNode)
	branchNode.GetParents().Add(b.Node)
}

func (b *Branches) GetKeyOfNode(node *NonRootNode) string {
	if b.GetBranch("0") == node {
		return "0"
	} else if b.GetBranch("1") == node {
		return "1"
	} else {
		panic("node not found")
	}
}

func (b *Branches) GetBranch(which string) *NonRootNode {
	return b.Branches.MustGet(which)
}

func (b *Branches) GetBothBranches() []*NonRootNode {
	return []*NonRootNode{
		b.GetBranch("0"),
		b.GetBranch("1"),
	}
}

func (b *Branches) HasBranchAsNode(node *Node) bool {
	if b.GetBranch("0") == node || b.GetBranch("1") == node {
		return true
	}
	return false
}

func (b *Branches) HasNodeIdAsBranch(id string) bool {
	if b.GetBranch("0").Id == id || b.GetBranch("1").Id == id {
		return true
	}
	return false
}

func (b *Branches) AreBranchesStrictEqual() bool {
	return b.Branches.MustGet("0") == b.Branches.MustGet("1")
}

func (b *Branches) HasEqualBranches() bool {
	branch0Json := b.Branches.MustGet("0").ToJson(false)
	branch1Json := b.Branches.MustGet("1").ToJson(false)
	branch0JsonString, err := json.Marshal(branch0Json)
	if err != nil {
		panic("error marshalling branch0Json: " + err.Error())
	}
	branch1JsonString, err := json.Marshal(branch1Json)
	if err != nil {
		panic("error marshalling branch1Json: " + err.Error())
	}
	return bytes.Equal(branch0JsonString, branch1JsonString)
}
