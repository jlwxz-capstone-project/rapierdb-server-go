package bdd

type RootNode struct {
	Branches     Branches
	Levels       []int
	NodesByLevel map[int]map[Node]struct{}
}
