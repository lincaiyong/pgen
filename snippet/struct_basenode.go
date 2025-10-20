package snippet

const BaseNodeStruct = `func NewBaseNode(filePath string, fileContent []rune, kind string, start, end Position) *BaseNode {
	return &BaseNode{filePath: filePath, fileContent: fileContent, kind: kind, start: start, end: end}
}

type BaseNode struct {
	filePath    string
	fileContent []rune
	kind        string
	start       Position
	end         Position
	parent      Node
	selfField   string
	replaceFun  func(Node)
	any_        any
}

func (n *BaseNode) FilePath() string {
	return n.filePath
}

func (n *BaseNode) FileContent() []rune {
	return n.fileContent
}

func (n *BaseNode) Kind() string {
	return n.kind
}

func (n *BaseNode) Range() (Position, Position) {
	return n.start, n.end
}

func (n *BaseNode) SetRange(start, end Position) {
	n.start = start
	n.end = end
}

func (n *BaseNode) RangeStart() Position {
	return n.start
}

func (n *BaseNode) RangeEnd() Position {
	return n.end
}

func (n *BaseNode) BuildLink() {
}

func (n *BaseNode) Parent() Node {
	return n.parent
}

func (n *BaseNode) SetParent(v Node) {
	n.parent = v
}

func (n *BaseNode) SelfField() string {
	return n.selfField
}

func (n *BaseNode) SetSelfField(v string) {
	n.selfField = v
}

func (n *BaseNode) ReplaceSelf(node Node) {
	node.SetReplaceSelf(n.replaceFun)
	node.SetParent(n.Parent())
	node.SetSelfField(n.SelfField())
	n.replaceFun(node)
}

func (n *BaseNode) SetReplaceSelf(fun func(Node)) {
	n.replaceFun = fun
}

func (n *BaseNode) Fields() []string {
	return nil
}

func (n *BaseNode) Child(_ string) Node {
	return DummyNode
}

func (n *BaseNode) SetChild(_ []Node) {
}

func (n *BaseNode) fork() *BaseNode {
	return &BaseNode{
		filePath:    n.filePath,
		fileContent: n.fileContent,
		kind:        n.kind,
		start:       n.start,
		end:         n.end,
		parent:      n.parent,
		selfField:   n.selfField,
		replaceFun:  n.replaceFun,
	}
}

func (n *BaseNode) Fork() Node {
	return n.fork()
}

func (n *BaseNode) Visit(func(Node) (bool, bool), func(Node) bool) bool {
	return false
}

func (n *BaseNode) Code() []rune {
	if n.fileContent == nil {
		return nil
	}
	code := n.fileContent
	start := 0
	end := len(code)
	if n.end.Offset <= len(code) && n.end.Offset >= 0 {
		end = n.end.Offset
	}
	if n.start.Offset >= 0 && n.start.Offset <= end {
		start = n.start.Offset
	}
	return code[start:end]
}

func (n *BaseNode) Dump(func(Node, map[string]string) string) map[string]string {
	return map[string]string{
		"kind": "?",
	}
}

func (n *BaseNode) IsDummy() bool {
	return n.kind == NodeTypeDummy
}

func (n *BaseNode) UnpackNodes() []Node {
	return nil
}

func (n *BaseNode) Any() any {
	return n.any_
}

func (n *BaseNode) SetAny(any_ any) {
	n.any_ = any_
}`
