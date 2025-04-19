package bdd

import (
	"fmt"
	"slices"
	"strings"
)

const (
	CHAR_CODE_OFFSET       = 40
	FIRST_CHAR_CODE_FOR_ID = 97
	MAX_LEAF_NODE_AMOUNT   = 99
)

func (bdd *RootNode) ToMinimalString() string {
	ret := strings.Builder{}
	currentCharCode := FIRST_CHAR_CODE_FOR_ID

	leafNodeAmount := len(bdd.GetLeafNodes())
	if leafNodeAmount > MAX_LEAF_NODE_AMOUNT {
		panic("leaf node amount is too large")
	}

	ret.WriteString(fmt.Sprintf("%02d", leafNodeAmount))

	levelsHighestFirst := slices.Clone(bdd.Levels)
	slices.Reverse(levelsHighestFirst)

	idByNode := make(map[*Node]rune)
	for _, level := range levelsHighestFirst {
		nodes := bdd.GetNodesOfLevel(level)
		for _, node := range nodes {
			id, nextCode, str := nodeToString(node, idByNode, currentCharCode)
			currentCharCode = nextCode
			idByNode[node] = id
			ret.WriteString(str)
		}
	}

	return ret.String()
}

func nodeToString(
	node *Node,
	idByNode map[*Node]rune,
	lastCode int,
) (rune, int, string) {
	nextId, nextCode := getNextCharId(lastCode)
	if node.IsLeafNode() {
		valueChar := getCharOfValue(node.AsLeafNode().Value)
		str := fmt.Sprintf("%c%c", valueChar, nextId)
		return nextId, nextCode, str
	} else if node.IsInternalNode() {
		internalNode := node.AsInternalNode()
		branch0Id := idByNode[internalNode.Branches.GetBranch("0")]
		branch1Id := idByNode[internalNode.Branches.GetBranch("1")]
		levelChar := getCharOfLevel(node.Level)
		str := fmt.Sprintf("%c%c%c%c", nextId, branch0Id, branch1Id, levelChar)
		return nextId, nextCode, str
	} else if node.IsRootNode() {
		rootNode := node.AsRootNode()
		branch0Id := idByNode[rootNode.Branches.GetBranch("0")]
		branch1Id := idByNode[rootNode.Branches.GetBranch("1")]
		levelChar := getCharOfLevel(node.Level)
		str := fmt.Sprintf("%c%c%c", branch0Id, branch1Id, levelChar)
		return nextId, nextCode, str
	}
	panic("unknown node type")
}

func getNextCharId(lastCode int) (rune, int) {
	if lastCode >= 128 && lastCode <= 160 {
		lastCode = 161
	}

	ch := rune(lastCode)
	return ch, lastCode + 1
}

func getCharOfValue(value int) rune {
	charCode := value + CHAR_CODE_OFFSET
	return rune(charCode)
}

func getCharOfLevel(level int) rune {
	charCode := level + CHAR_CODE_OFFSET
	return rune(charCode)
}
