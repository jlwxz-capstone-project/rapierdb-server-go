package bdd

var lastIdGen = 0

func NextNodeId() int {
	ret := lastIdGen
	lastIdGen++
	return ret
}
