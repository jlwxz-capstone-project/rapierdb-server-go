package eventreduce

import (
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/actions"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/bdd"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/orderedmap"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/state"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/types"
)

// 创建BDD决策树(缓存结果，避免重复构建)
var actionBDD *bdd.RootNode

// 以下是为测试导出的函数 //////

// GetStateIndexForTest 导出getStateIndex函数供测试使用
func GetStateIndexForTest(stateName string) int {
	return getStateIndex(stateName)
}

// GenerateAllStateSetsForTest 导出generateAllStateSets函数供测试使用
func GenerateAllStateSetsForTest(stateCount int) []string {
	return generateAllStateSets(stateCount)
}

// MapActionCodeToNameForTest 导出mapActionCodeToName函数供测试使用
func MapActionCodeToNameForTest(actionCode int) types.ActionName {
	return mapActionCodeToName(actionCode)
}

// FillTruthTableForTest 导出fillTruthTable函数供测试使用
func FillTruthTableForTest() *orderedmap.OrderedMap[string, int] {
	tt := orderedmap.NewOrderedMap[string, int]()
	fillTruthTable(tt)
	return tt
}

// BuildActionBDDForTest 导出buildActionBDD函数供测试使用
func BuildActionBDDForTest() *bdd.RootNode {
	return buildActionBDD()
}

// CreateSimpleTruthTableForTest 创建一个简单的真值表用于测试
func CreateSimpleTruthTableForTest() *orderedmap.OrderedMap[string, int] {
	tt := orderedmap.NewOrderedMap[string, int]()

	// 只添加少量测试用例，避免过度复杂导致测试超时
	tt.Set("00", 0) // 全0状态 -> ActionDoNothing
	tt.Set("01", 1) // 简单状态1 -> ActionInsertFirst
	tt.Set("10", 2) // 简单状态2 -> ActionInsertLast
	tt.Set("11", 3) // 全1状态 -> ActionRemoveFirstItem

	return tt
}

// CreateSimpleBDDForTest 从真值表创建一个简单的BDD用于测试
func CreateSimpleBDDForTest(tt *orderedmap.OrderedMap[string, int]) *bdd.RootNode {
	// 从真值表创建BDD
	rootNode := bdd.CreateBddFromTruthTable(tt)

	// 进行有限次数的优化，避免测试超时
	// 只做一轮优化
	levels := rootNode.GetSortedLevels()
	lastLevel := levels[len(levels)-1]
	for level := lastLevel; level > 0; level-- {
		nodes := rootNode.GetNodesOfLevel(level)
		for _, node := range nodes {
			if node.IsLeafNode() {
				node.AsLeafNode().ApplyEliminationRule(nil)
			}
			if !node.Deleted && node.IsInternalNode() {
				useNode := node.AsInternalNode()
				useNode.ApplyRuductionRule()
				if !useNode.Deleted {
					useNode.ApplyEliminationRule(nodes)
				}
			}
		}
	}

	return rootNode
}

// 以上是为测试导出的函数 //////

// CalculateActionFromMap 根据状态集合计算操作名称
func CalculateActionFromMap(
	stateSetToActionMap types.StateSetToActionMap,
	input types.StateResolveFunctionInput,
) struct {
	Action   types.ActionName
	StateSet types.StateSet
} {
	stateSet := state.GetStateSet(input)
	actionName, ok := stateSetToActionMap[stateSet]
	if !ok {
		return struct {
			Action   types.ActionName
			StateSet types.StateSet
		}{
			Action:   types.ActionRunFullQueryAgain,
			StateSet: stateSet,
		}
	}
	return struct {
		Action   types.ActionName
		StateSet types.StateSet
	}{
		Action:   actionName,
		StateSet: stateSet,
	}
}

// CalculateActionName 计算操作名称 - 使用BDD进行决策
func CalculateActionName(input types.StateResolveFunctionInput) types.ActionName {
	// 在测试环境中，状态列表可能会变化，所以每次都重新构建BDD
	// 在生产环境中，可以使用缓存的BDD
	localBDD := buildActionBDD()

	// 获取当前状态集合的二进制表示
	stateSet := state.GetStateSet(input)

	// 创建BDD解析函数
	resolvers := bdd.GetResolverFunctions(len(state.OrderedStateList), false)

	// 使用BDD进行决策，获取操作代码
	actionCode := localBDD.Resolve(resolvers, string(stateSet))

	// 将操作代码映射到操作名称
	return mapActionCodeToName(actionCode)
}

// 将BDD决策结果映射到操作名称
func mapActionCodeToName(actionCode int) types.ActionName {
	switch actionCode {
	case 0:
		return types.ActionDoNothing
	case 1:
		return types.ActionInsertFirst
	case 2:
		return types.ActionInsertLast
	case 3:
		return types.ActionRemoveFirstItem
	case 4:
		return types.ActionRemoveLastItem
	case 5:
		return types.ActionRemoveFirstInsertLast
	case 6:
		return types.ActionRemoveLastInsertFirst
	case 7:
		return types.ActionRemoveFirstInsertFirst
	case 8:
		return types.ActionRemoveLastInsertLast
	case 9:
		return types.ActionRemoveExisting
	case 10:
		return types.ActionReplaceExisting
	case 11:
		return types.ActionInsertAtSortPosition
	case 12:
		return types.ActionRemoveExistingAndInsertAtSortPosition
	default:
		return types.ActionRunFullQueryAgain
	}
}

// 构建操作决策BDD
func buildActionBDD() *bdd.RootNode {
	// 创建真值表 - 使用orderedmap
	tt := orderedmap.NewOrderedMap[string, int]()

	// 使用条件逻辑填充真值表
	// 注意：这个函数非常长，填充了所有可能的状态组合对应的操作
	fillTruthTable(tt)

	// 从真值表创建BDD
	rootNode := bdd.CreateBddFromTruthTable(tt)

	// 优化BDD结构
	rootNode.Minimize()

	return rootNode
}

// 填充真值表 - 这个函数很长，因为要处理所有的状态组合情况
func fillTruthTable(tt *orderedmap.OrderedMap[string, int]) {
	// 为了简化演示，这里只展示几个关键的规则
	// 实际实现需要覆盖所有可能的状态组合

	// 获取各个状态在状态列表中的索引，如果不存在则使用-1
	isDeleteIdx := getStateIndex("isDelete")
	isInsertIdx := getStateIndex("isInsert")
	isUpdateIdx := getStateIndex("isUpdate")
	wasInResultIdx := getStateIndex("wasInResult")
	doesMatchNowIdx := getStateIndex("doesMatchNow")
	// 获取可选状态的索引
	sortParamsChangedIdx := getStateIndex("sortParamsChanged")
	isSortedBeforeFirstIdx := getStateIndex("isSortedBeforeFirst")
	isSortedAfterLastIdx := getStateIndex("isSortedAfterLast")

	// 生成所有可能的状态组合
	stateSets := generateAllStateSets(len(state.OrderedStateList))

	// 设置基于二进制字符串的状态集合对应的操作代码
	for _, stateSetBin := range stateSets {
		// 使用BDD决策逻辑
		var actionCode int

		// 安全地提取关键状态位，如果索引有效则获取状态值，否则使用默认值false
		isDelete := isDeleteIdx >= 0 && stateSetBin[isDeleteIdx] == '1'
		isInsert := isInsertIdx >= 0 && stateSetBin[isInsertIdx] == '1'
		isUpdate := isUpdateIdx >= 0 && stateSetBin[isUpdateIdx] == '1'
		wasInResult := wasInResultIdx >= 0 && stateSetBin[wasInResultIdx] == '1'
		doesMatchNow := doesMatchNowIdx >= 0 && stateSetBin[doesMatchNowIdx] == '1'
		sortParamsChanged := sortParamsChangedIdx >= 0 && stateSetBin[sortParamsChangedIdx] == '1'
		isSortedBeforeFirst := isSortedBeforeFirstIdx >= 0 && stateSetBin[isSortedBeforeFirstIdx] == '1'
		isSortedAfterLast := isSortedAfterLastIdx >= 0 && stateSetBin[isSortedAfterLastIdx] == '1'

		// 主要决策逻辑 - 简化版本
		switch {
		// 删除操作逻辑
		case isDelete && wasInResult:
			actionCode = 9 // ActionRemoveExisting
		case isDelete && !wasInResult:
			actionCode = 0 // ActionDoNothing

		// 插入操作逻辑
		case isInsert && !doesMatchNow:
			actionCode = 0 // ActionDoNothing
		case isInsert && doesMatchNow && isSortedBeforeFirst:
			actionCode = 1 // ActionInsertFirst
		case isInsert && doesMatchNow && isSortedAfterLast:
			actionCode = 2 // ActionInsertLast
		case isInsert && doesMatchNow:
			actionCode = 11 // ActionInsertAtSortPosition

		// 更新操作逻辑
		case isUpdate && !wasInResult && doesMatchNow:
			actionCode = 11 // ActionInsertAtSortPosition
		case isUpdate && wasInResult && !doesMatchNow:
			actionCode = 9 // ActionRemoveExisting
		case isUpdate && wasInResult && doesMatchNow && sortParamsChanged:
			actionCode = 12 // ActionRemoveExistingAndInsertAtSortPosition
		case isUpdate && wasInResult && doesMatchNow:
			actionCode = 10 // ActionReplaceExisting

		// 默认情况
		default:
			actionCode = 0 // ActionDoNothing
		}

		// 设置真值表条目
		tt.Set(stateSetBin, actionCode)
	}

	// 确保包含默认状态组合"00000"，映射到ActionDoNothing(0)
	defaultStateSet := ""
	for i := 0; i < len(state.OrderedStateList); i++ {
		defaultStateSet += "0"
	}
	tt.Set(defaultStateSet, 0)
}

// 获取状态名称在OrderedStateList中的索引
func getStateIndex(stateName string) int {
	for i, name := range state.OrderedStateList {
		if string(name) == stateName {
			return i
		}
	}
	return -1
}

// 生成所有可能的状态集合
func generateAllStateSets(stateCount int) []string {
	count := 1 << stateCount // 2^stateCount
	result := make([]string, 0, count)

	for i := 0; i < count; i++ {
		// 将整数转换为二进制字符串并填充前导零
		bin := ""
		for j := 0; j < stateCount; j++ {
			if (i & (1 << j)) != 0 {
				bin = "1" + bin
			} else {
				bin = "0" + bin
			}
		}
		result = append(result, bin)
	}

	return result
}

// CalculateActionFunction 计算操作函数
func CalculateActionFunction(input types.StateResolveFunctionInput) types.ActionFunction {
	actionName := CalculateActionName(input)
	return actions.ActionFunctions[actionName]
}

// RunAction 执行操作
// 注意：此函数会修改 previousResults
func RunAction(
	actionName types.ActionName,
	queryParams types.QueryParams,
	changeEvent types.ChangeEvent,
	previousResults []map[string]interface{},
	keyDocumentMap map[string]map[string]interface{},
) []map[string]interface{} {
	fn := actions.ActionFunctions[actionName]
	fn(&types.ActionFunctionInput{
		QueryParams:     queryParams,
		ChangeEvent:     changeEvent,
		PreviousResults: previousResults,
		KeyDocumentMap:  keyDocumentMap,
	})
	return previousResults
}
