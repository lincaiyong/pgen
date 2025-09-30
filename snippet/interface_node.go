package snippet

const NodeInterface = `type Node interface {
	Kind() string
	Range() (Position, Position)
	SetRange(Position, Position)
	RangeStart() Position
	RangeEnd() Position
	BuildLink()
	Parent() Node
	SetParent(Node)
	SelfField() string
	SetSelfField(string)
	Fields() []string
	ReplaceSelf(Node)
	SetReplaceSelf(func(Node))
	Child(field string) Node
	SetChild(nodes []Node)
	Fork() Node
	Visit(func(Node) (visitChildren, exit bool), func(Node) (exit bool)) (exit bool)
	FilePath() string
	FileContent() []rune
	Code() []rune
	Dump(hook func(Node, map[string]string) string) map[string]string
	IsDummy() bool
	UnpackNodes() []Node
}`
