package bdd

type Node interface {
	isNode()
	GetLevel() int
	GetId() string
	IsDeleted() bool
	IsEqualToOtherNode(otherNode Node, ownString *string) bool
	Remove()
	ToJson(withId bool) NodeJson
	ToString() string
	EnsureNotDeleted(op string)
	ApplyEliminationRule(nodesOfSameLevel []Node) bool
}

type NodeJson struct {
	Id       *string             `json:"id,omitempty"`
	Deleted  bool                `json:"deleted"`
	Level    int                 `json:"level"`
	Type     string              `json:"type"`
	Parents  []NodeJson          `json:"parents,omitempty"`
	Value    *int                `json:"value,omitempty"`
	Branches map[string]NodeJson `json:"branches,omitempty"`
}

func (n *InternalNode) isNode() {}
func (n *RootNode) isNode()     {}
func (n *LeafNode) isNode()     {}

type NonLeafNode interface {
	Node
	isNonLeafNode()
	GetBranches() *Branches
}

func (n *InternalNode) isNonLeafNode() {}
func (n *RootNode) isNonLeafNode()     {}

func (n *InternalNode) GetBranches() *Branches { return n.Branches }
func (n *RootNode) GetBranches() *Branches     { return n.Branches }

type NonRootNode interface {
	Node
	isNonRootNode()
	GetParents() *Parents
}

func (n *InternalNode) isNonRootNode() {}
func (n *LeafNode) isNonRootNode()     {}

func (n *InternalNode) GetParents() *Parents { return n.Parents }
func (n *LeafNode) GetParents() *Parents     { return n.Parents }
