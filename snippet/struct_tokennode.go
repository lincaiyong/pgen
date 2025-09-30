package snippet

const TokenNodeStruct = `func NewTokenNode(filePath string, fileContent []rune, token *Token) Node {
	return &TokenNode{
		BaseNode: NewBaseNode(filePath, fileContent, NodeTypeToken, token.Start, token.End),
		token:    token,
	}
}

type TokenNode struct {
	*BaseNode
	token *Token
}

func (n *TokenNode) Token() *Token {
	return n.token
}

func (n *TokenNode) Visit(beforeChildren func(Node) (visitChildren, exit bool), afterChildren func(Node) (exit bool)) (exit bool) {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *TokenNode) Fork() Node {
	return &TokenNode{
		BaseNode: n.BaseNode.fork(),
		token:    n.token,
	}
}

func (n *TokenNode) Dump(func(Node, map[string]string) string) map[string]string {
	val := string(n.Code())
	val = strings.ReplaceAll(val, "\\", "\\\\")
	val = strings.ReplaceAll(val, "\"", "\\\"")
	val = strings.ReplaceAll(val, "\n", "\\n")
	val = strings.ReplaceAll(val, "\r", "\\r")
	val = strings.ReplaceAll(val, "\t", "\\t")
	val = fmt.Sprintf("\"%s\"", val)
	return map[string]string{
		"kind": "\"token\"",
		"code": val,
	}
}`
