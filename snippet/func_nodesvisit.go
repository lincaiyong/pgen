package snippet

const NodesVisitFunc = `func nodesVisit(nodes []Node, before func(Node) (visitChild, exit bool), after func(Node) (exit bool)) (exit bool) {
	for _, node := range nodes {
		if node.Visit(before, after) {
			return true
		}
	}
	return false
}`
