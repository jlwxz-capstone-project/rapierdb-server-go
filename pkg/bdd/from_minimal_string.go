package bdd

import (
	"fmt"
	"strconv"

	pe "github.com/pkg/errors"
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func NewBddFromMinimalString(str string) (*SimpleBdd, error) {
	runes := []rune(str)
	nodesById := make(map[rune]any) // map[rune](*SimpleBdd | int)

	// Parse leaf nodes
	// Assuming the first two runes represent the count as a string
	leafNodeAmountStr := string(runes[0:2])
	leafNodeAmount := must(strconv.Atoi(leafNodeAmountStr))
	lastLeafNodeRuneIndex := 2 + leafNodeAmount*2
	if len(runes) < lastLeafNodeRuneIndex {
		return nil, pe.WithStack(fmt.Errorf("string too short for leaf nodes. expected length %d, got %d", lastLeafNodeRuneIndex, len(runes)))
	}
	leafNodeRunes := runes[2:lastLeafNodeRuneIndex]
	leafNodeChunks := splitRunesToChunks(leafNodeRunes, 2)
	for _, chunk := range leafNodeChunks {
		if len(chunk) != 2 {
			return nil, pe.WithStack(fmt.Errorf("invalid leaf node chunk size: %d, expected 2. chunk: %s", len(chunk), string(chunk)))
		}
		id := chunk[0]
		value := getNumberOfChar(chunk[1])
		nodesById[id] = value
	}

	// Parse internal nodes
	internalNodesEndIndex := len(runes) - 3
	if lastLeafNodeRuneIndex > internalNodesEndIndex {
		return nil, pe.WithStack(fmt.Errorf("string too short for internal nodes. leaf nodes end at %d, string ends at %d", lastLeafNodeRuneIndex, internalNodesEndIndex))
	}
	internalNodeRunes := runes[lastLeafNodeRuneIndex:internalNodesEndIndex]
	internalNodeChunks := splitRunesToChunks(internalNodeRunes, 4)
	for _, chunk := range internalNodeChunks {
		if len(chunk) != 4 {
			return nil, pe.WithStack(fmt.Errorf("invalid internal node chunk size: %d, expected 4. chunk: %s", len(chunk), string(chunk)))
		}
		id := chunk[0]
		idOfBranch0 := chunk[1]
		idOfBranch1 := chunk[2]
		level := getNumberOfChar(chunk[3])

		if _, ok := nodesById[idOfBranch0]; !ok {
			return nil, pe.WithStack(fmt.Errorf("missing node branch0 with id %c (%d)", idOfBranch0, idOfBranch0))
		}

		if _, ok := nodesById[idOfBranch1]; !ok {
			return nil, pe.WithStack(fmt.Errorf("missing node branch1 with id %c (%d)", idOfBranch1, idOfBranch1))
		}

		node0 := nodesById[idOfBranch0]
		node1 := nodesById[idOfBranch1]

		node := &SimpleBdd{
			Branch0: node0,
			Branch1: node1,
			Level:   level,
		}
		nodesById[id] = node
	}

	// Parse root node
	if len(runes) < 3 {
		return nil, pe.WithStack(fmt.Errorf("string too short for root node definition, length %d", len(runes)))
	}
	last3Runes := runes[len(runes)-3:]
	idOf0 := last3Runes[0]
	idOf1 := last3Runes[1]
	levelOfRoot := getNumberOfChar(last3Runes[2])

	if _, ok := nodesById[idOf0]; !ok {
		return nil, pe.WithStack(fmt.Errorf("missing root node branch0 with id %c (%d)", idOf0, idOf0))
	}
	if _, ok := nodesById[idOf1]; !ok {
		return nil, pe.WithStack(fmt.Errorf("missing root node branch1 with id %c (%d)", idOf1, idOf1))
	}

	nodeOf0 := nodesById[idOf0]
	nodeOf1 := nodesById[idOf1]

	root := &SimpleBdd{
		Branch0: nodeOf0,
		Branch1: nodeOf1,
		Level:   levelOfRoot,
	}
	return root, nil
}

// splitRunesToChunks splits a slice of runes into chunks of a specified size.
func splitRunesToChunks(runes []rune, chunkSize int) [][]rune {
	if chunkSize <= 0 {
		return nil // Or handle error appropriately
	}
	numChunks := (len(runes) + chunkSize - 1) / chunkSize
	chunks := make([][]rune, 0, numChunks)
	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, runes[i:end])
	}
	return chunks
}

// getNumberOfChar converts a rune to an integer based on an offset.
// Ensure CHAR_CODE_OFFSET is defined correctly.
func getNumberOfChar(char rune) int {
	// This assumes CHAR_CODE_OFFSET is defined elsewhere in the package
	return int(char - CHAR_CODE_OFFSET)
}

// remove splitStringToChunks as it's replaced by splitRunesToChunks
/*
func splitStringToChunks(str string, chunkSize int) []string {
	chunks := make([]string, 0, len(str)/chunkSize+1)
	for i := 0; i < len(str); i += chunkSize {
		end := i + chunkSize
		if end > len(str) {
			end = len(str)
		}
		chunks = append(chunks, str[i:end])
	}
	return chunks
}
*/
