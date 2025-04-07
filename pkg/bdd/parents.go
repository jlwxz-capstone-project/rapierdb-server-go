package bdd

import "strings"

type Parents struct {
	Node    *NonRootNode
	Parents map[*NonLeafNode]struct{}
}

func NewParents(node *NonRootNode) *Parents {
	return &Parents{
		Node:    node,
		Parents: map[*NonLeafNode]struct{}{},
	}
}

func (p *Parents) Remove(node *NonLeafNode) {
	delete(p.Parents, node)

	if p.Size() == 0 {
		p.Node.Remove()
	}
}

func (p *Parents) GetAll() []*NonLeafNode {
	result := make([]*NonLeafNode, 0, len(p.Parents))
	for parent := range p.Parents {
		result = append(result, parent)
	}
	return result
}

func (p *Parents) Add(node *NonLeafNode) {
	if p.Node.Level == node.Level {
		panic("a node cannot be parent of a node with the same level")
	}
	p.Parents[node] = struct{}{}
}

func (p *Parents) Has(node *NonLeafNode) bool {
	_, ok := p.Parents[node]
	return ok
}

func (p *Parents) Size() int {
	return len(p.Parents)
}

func (p *Parents) ToString() string {
	ret := make([]string, 0, len(p.Parents))
	for parent := range p.Parents {
		ret = append(ret, parent.Id)
	}
	return strings.Join(ret, ", ")
}
