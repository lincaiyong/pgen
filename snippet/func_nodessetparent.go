package snippet

const NodesSetParentFunc = `func nodesSetParent(targets []Node, parent Node, field string) {
	for i, target := range targets {
		target.SetParent(parent)
		target.SetSelfField(strconv.Itoa(i))
		if field != "" {
			target.SetSelfField(field)
		}
	}
}`
