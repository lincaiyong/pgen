package models

func NewAstNode(name string, args []string, snippet *Snippet) *AstNode {
	args2 := make([]*Name, len(args))
	for i, arg := range args {
		args2[i] = NewName(arg)
	}
	return &AstNode{
		name:    name,
		args:    args2,
		snippet: snippet,
	}
}

type AstNode struct {
	name    string
	args    []*Name
	snippet *Snippet
}

func (a *AstNode) Name() string {
	return a.name
}

func (a *AstNode) Args() []*Name {
	return a.args
}

func (a *AstNode) Snippet() *Snippet {
	return a.snippet
}
