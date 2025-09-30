package models

const (
	TokenRuleNodeTypeRule                  = "rule"
	TokenRuleNodeTypeChoice                = "choice"
	TokenRuleNodeTypeNegativeLookaheadItem = "negative-lookahead-item"
	TokenRuleNodeTypePositiveLookaheadItem = "positive-lookahead-item"
	TokenRuleNodeTypeRepeat0Item           = "repeat-0-item"
	TokenRuleNodeTypeRepeat1Item           = "repeat-1-item"
	TokenRuleNodeTypeOptionalItem          = "optional-item"
	TokenRuleNodeTypeAtomItem              = "atom-item"
	TokenRuleNodeTypeNameAtom              = "name-atom"
	TokenRuleNodeTypeStringAtom            = "string-atom"
	TokenRuleNodeTypeCharacterClassAtom    = "character-class-atom"
	TokenRuleNodeTypeGroupAtom             = "group-atom"
)

func NewTokenRuleNode(kind string, parent *TokenRuleNode) *TokenRuleNode {
	return &TokenRuleNode{
		kind:   kind,
		parent: parent,
	}
}

type TokenRuleNode struct {
	kind     string
	parent   *TokenRuleNode
	children []*TokenRuleNode
	snippet  *Snippet
	name     string // rule name
}

func (n *TokenRuleNode) Visit(fn func(*TokenRuleNode)) {
	fn(n)
	for _, child := range n.children {
		child.Visit(fn)
	}
}

func (n *TokenRuleNode) Kind() string {
	return n.kind
}

func (n *TokenRuleNode) SetKind(kind string) {
	n.kind = kind
}

func (n *TokenRuleNode) Parent() *TokenRuleNode {
	return n.parent
}

func (n *TokenRuleNode) SetParent(parent *TokenRuleNode) {
	n.parent = parent
}

func (n *TokenRuleNode) Children() []*TokenRuleNode {
	return n.children
}

func (n *TokenRuleNode) SetChildren(children []*TokenRuleNode) {
	n.children = children
}

func (n *TokenRuleNode) Child() *TokenRuleNode {
	if len(n.children) > 0 {
		return n.children[0]
	}
	return nil
}

func (n *TokenRuleNode) SetChild(child *TokenRuleNode) {
	if len(n.children) > 0 {
		n.children[0] = child
	} else {
		n.children = append(n.children, child)
	}
}

func (n *TokenRuleNode) Snippet() *Snippet {
	return n.snippet
}

func (n *TokenRuleNode) SetSnippet(snippet *Snippet) {
	n.snippet = snippet
}

func (n *TokenRuleNode) Name() string {
	return n.name
}

func (n *TokenRuleNode) SetName(name string) {
	n.name = name
}
