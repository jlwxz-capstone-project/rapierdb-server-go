package bdd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func EnsureCorrectBdd(bdd *RootNode) error {
	bddJson := bdd.ToJson(true)
	bddJsonString, err := json.Marshal(bddJson)
	if err != nil {
		panic("error marshalling bddJson: " + err.Error())
	}

	allNodes := []*Node{}
	nodesById := make(map[string]*Node)
	for _, level := range bdd.GetSortedLevels() {
		levelNodes := bdd.GetNodesOfLevel(level)
		for _, node := range levelNodes {
			allNodes = append(allNodes, node)
			nodesById[node.Id] = node
		}
	}

	recursiveNodes := getNodesRecursive(bdd)

	if len(allNodes) != len(recursiveNodes) {
		allNodesIds := make([]string, 0, len(allNodes))
		for _, node := range allNodes {
			allNodesIds = append(allNodesIds, node.Id)
		}
		sort.Strings(allNodesIds)

		recursiveNodesIds := make([]string, 0, len(recursiveNodes))
		for node := range recursiveNodes {
			recursiveNodesIds = append(recursiveNodesIds, node.Id)
		}
		sort.Strings(recursiveNodesIds)

		nodesOnlyInRecursive := make([]string, 0)
		for _, nodeId := range recursiveNodesIds {
			found := false
			for _, allNodeId := range allNodesIds {
				if nodeId == allNodeId {
					found = true
					break
				}
			}
			if !found {
				nodesOnlyInRecursive = append(nodesOnlyInRecursive, nodeId)
			}
		}

		if len(recursiveNodes) > len(allNodes) {
			firstId := nodesOnlyInRecursive[0]
			var referenceToFirst *Node
			for _, n := range allNodes {
				if n.IsInternalNode() {
					if n.GetBranches().HasNodeIdAsBranch(firstId) {
						referenceToFirst = n
						break
					}
				}
			}
			if referenceToFirst != nil {
				fmt.Println("referenceToFirst:", referenceToFirst.ToString())
			}
		}

		return fmt.Errorf(
			"ensureCorrectBdd() nodes in list not equal size to recursive nodes. allNodes: %d recursiveNodes: %d",
			len(allNodes), len(recursiveNodes),
		)
	}

	for _, node := range allNodes {
		if node.IsRootNode() {
			continue
		}
		useNode := node

		if useNode.Deleted {
			return fmt.Errorf("ensureCorrectBdd() bdd includes a deleted node")
		}

		if useNode.GetParents().Size() == 0 {
			return fmt.Errorf("ensureCorrectBdd() node has no parent %s", useNode.Id)
		}

		if useNode.IsInternalNode() {
			internalNode := useNode.AsInternalNode()
			bothBranches := internalNode.GetBranches().GetBothBranches()

			if internalNode.GetBranches().AreBranchesStrictEqual() {
				branchIds := make([]string, 0, len(bothBranches))
				for _, branch := range bothBranches {
					branchIds = append(branchIds, branch.Id)
				}
				return fmt.Errorf("ensureCorrectBdd() node has two equal branches: %s", strings.Join(branchIds, ", "))
			}

			for _, branch := range bothBranches {
				if !branch.GetParents().Has(internalNode.AsNode()) {
					return fmt.Errorf("ensureCorrectBdd() branch must have the node as parent")
				}
			}
		}

		for _, parent := range useNode.GetParents().GetAll() {
			if !parent.GetBranches().HasBranchAsNode(useNode) {
				return fmt.Errorf("ensureCorrectBdd() parent node does not have child as branch")
			}
		}
	}

	if strings.Contains(string(bddJsonString), `"deleted":true`) {
		return fmt.Errorf("ensureCorrectBdd() bdd includes a deleted node")
	}

	return nil
}

func getNodesRecursive(bdd *RootNode) map[*Node]struct{} {
	return getNodesRecursiveImpl(bdd.AsNode(), make(map[*Node]struct{}))
}

func getNodesRecursiveImpl(node *Node, set map[*Node]struct{}) map[*Node]struct{} {
	set[node] = struct{}{}
	if !node.IsLeafNode() { // not leaf node
		useNode := node
		branch0 := useNode.GetBranches().GetBranch("0")
		set[branch0] = struct{}{}
		getNodesRecursiveImpl(branch0, set)

		branch1 := useNode.GetBranches().GetBranch("1")
		set[branch1] = struct{}{}
		getNodesRecursiveImpl(branch1, set)
	}
	return set
}
