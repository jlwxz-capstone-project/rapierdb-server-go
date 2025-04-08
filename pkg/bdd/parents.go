package bdd

import (
	"strconv"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/orderedset"
)

type Parents struct {
	Node    *NonRootNode
	Parents *orderedset.OrderedSet[*NonLeafNode]
}

func NewParents(node *NonRootNode) *Parents {
	return &Parents{
		Node:    node,
		Parents: orderedset.NewOrderedSet[*NonLeafNode](),
	}
}

func (p *Parents) Remove(node *NonLeafNode) {
	p.Parents.Remove(node)

	if p.Size() == 0 {
		p.Node.Remove()
	}
}

func (p *Parents) GetAll() []*NonLeafNode {
	result := make([]*NonLeafNode, 0, p.Parents.Len())
	for parent := range p.Parents.IterValues() {
		result = append(result, parent)
	}
	return result
}

func (p *Parents) Add(node *NonLeafNode) {
	if p.Node.Level == node.Level {
		panic("a node cannot be parent of a node with the same level")
	}
	p.Parents.Add(node)
}

func (p *Parents) Has(node *NonLeafNode) bool {
	return p.Parents.Contains(node)
}

func (p *Parents) Size() int {
	return p.Parents.Len()
}

func (p *Parents) ToString() string {
	ret := make([]string, 0, p.Parents.Len())
	for parent := range p.Parents.IterValues() {
		ret = append(ret, strconv.Itoa(parent.Id))
	}
	return strings.Join(ret, ", ")
}
