package bdd

type SimpleBdd struct {
	Branch0 any // *SimpleBdd | int
	Branch1 any // *SimpleBdd | int
	Level   int
}

func (b *SimpleBdd) Resolve(fns ResolverFunctions, input string) int {
	currentNode := any(b)
	currentLevel := b.Level
	for {
		res := fns[currentLevel](input)
		branchKey := BooleanToBooleanString(res)
		switch node := currentNode.(type) {
		case int:
			return node
		case *SimpleBdd:
			if branchKey == "0" {
				currentNode = node.Branch0
			} else {
				currentNode = node.Branch1
			}
			currentLevel = node.Level
		}
	}
}
