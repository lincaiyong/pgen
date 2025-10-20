package snippet

const NodesNodeStruct = `func NewNodesNode(nodes []Node) Node {
	if len(nodes) == 0 {
		return DummyNode
	}
	filePath := nodes[0].FilePath()
	fileContent := nodes[0].FileContent()
	start := nodes[0].RangeStart()
	end := nodes[len(nodes)-1].RangeEnd()
	ret := &NodesNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeNodes, start, end),
		nodes:    nodes,
	}
	creationHook(ret)
	return ret
}

type NodesNode struct {
	*BaseNode
	nodes []Node
}

func (n *NodesNode) Nodes() []Node {
	return n.nodes
}

func (n *NodesNode) SetNodes(v []Node) {
	n.nodes = v
}

func (n *NodesNode) Fields() []string {
	ret := make([]string, 0)
	for i := 0; i < len(n.nodes); i++ {
		ret = append(ret, strconv.Itoa(i))
	}
	return ret
}

func (n *NodesNode) BuildLink() {
	nodesSetParent(n.nodes, n, "")
	for _, target := range n.nodes {
		target.BuildLink()
		target.SetReplaceSelf(func(n Node) {
			i, _ := strconv.Atoi(n.SelfField())
			n.Parent().(INodesNode).Nodes()[i] = n
		})
	}
}

func (n *NodesNode) Child(field string) Node {
	index, err := strconv.Atoi(field)
	if err != nil {
		return DummyNode
	}
	if index >= 0 && index < len(n.nodes) {
		return n.nodes[index]
	}
	return DummyNode
}

func (n *NodesNode) SetChild(nodes []Node) {
	n.nodes = nodes
}

func (n *NodesNode) Fork() Node {
	nodes := make([]Node, 0)
	for _, n := range n.nodes {
		nodes = append(nodes, n.Fork())
	}
	_ret := &NodesNode{
		BaseNode: n.BaseNode.fork(),
		nodes:    nodes,
	}
	nodesSetParent(_ret.nodes, _ret, "")
	return _ret
}

func (n *NodesNode) Visit(beforeChildren func(Node) (visitChildren, exit bool), afterChildren func(Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if nodesVisit(n.nodes, beforeChildren, afterChildren) {
		return true
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *NodesNode) dumpNodes(hook func(Node, map[string]string) string) string {
	items := make([]string, 0)
	for _, t := range n.nodes {
		items = append(items, DumpNode(t, hook))
	}
	return fmt.Sprintf("[%s]", strings.Join(items, ", "))
}

func (n *NodesNode) Dump(hook func(Node, map[string]string) string) map[string]string {
	return map[string]string{
		"kind":  "\"nodes\"",
		"nodes": n.dumpNodes(hook),
	}
}

func (n *NodesNode) UnpackNodes() []Node {
	return n.Nodes()
}`
