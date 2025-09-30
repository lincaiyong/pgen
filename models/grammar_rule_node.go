package models

const (
	GrammarRuleNodeTypeRule                  = "rule"
	GrammarRuleNodeTypeChoice                = "choice"
	GrammarRuleNodeTypeOptionalItem          = "optional-item"
	GrammarRuleNodeTypeRepeat0Item           = "repeat-0-item"
	GrammarRuleNodeTypeRepeat1Item           = "repeat-1-item"
	GrammarRuleNodeTypeSeparatedRepeat0Item  = "separated-repeat-0-item"
	GrammarRuleNodeTypeSeparatedRepeat1Item  = "separated-repeat-1-item"
	GrammarRuleNodeTypeNegativeLookaheadItem = "negative-lookahead-item"
	GrammarRuleNodeTypePositiveLookaheadItem = "positive-lookahead-item"
	GrammarRuleNodeTypeForwardIfNotMatchItem = "forward-if-not-match-item"
	GrammarRuleNodeTypeAtomItem              = "atom-item"
	GrammarRuleNodeTypeNameAtom              = "name-atom"
	GrammarRuleNodeTypeTokenAtom             = "token-atom"
	GrammarRuleNodeTypeStringAtom            = "string-atom"
	GrammarRuleNodeTypeGroupAtom             = "group-atom"
	GrammarRuleNodeTypeBracketEllipsisAtom   = "bracket-ellipsis-atom"

	GrammarRuleNodeTypeCallAction = "call-action"
	GrammarRuleNodeTypeNameAction = "name-action"
	GrammarRuleNodeTypeListAction = "list-action"
	GrammarRuleNodeTypeNullAction = "null-action"
)

func NewGrammarRuleNode(kind string, parent *GrammarRuleNode) *GrammarRuleNode {
	return &GrammarRuleNode{
		kind:   kind,
		parent: parent,
	}
}

type GrammarRuleNode struct {
	kind     string
	parent   *GrammarRuleNode
	children []*GrammarRuleNode
	snippet  *Snippet

	name string // rule name / item name

	ruleMemo bool

	separator *GrammarRuleNode
	action    *GrammarRuleNode

	suffix string // [ or ]
}

func (g *GrammarRuleNode) Visit(fn func(*GrammarRuleNode)) {
	fn(g)
	for _, child := range g.children {
		child.Visit(fn)
	}
	if g.separator != nil {
		g.separator.Visit(fn)
	}
}

func (g *GrammarRuleNode) Kind() string {
	return g.kind
}

func (g *GrammarRuleNode) SetKind(kind string) {
	g.kind = kind
}

func (g *GrammarRuleNode) Parent() *GrammarRuleNode {
	return g.parent
}

func (g *GrammarRuleNode) SetParent(parent *GrammarRuleNode) {
	g.parent = parent
}

func (g *GrammarRuleNode) Children() []*GrammarRuleNode {
	return g.children
}

func (g *GrammarRuleNode) SetChildren(children []*GrammarRuleNode) {
	g.children = children
}

func (g *GrammarRuleNode) Child() *GrammarRuleNode {
	if len(g.children) == 0 {
		return nil
	}
	return g.children[0]
}

func (g *GrammarRuleNode) SetChild(child *GrammarRuleNode) {
	if len(g.children) == 0 {
		g.children = []*GrammarRuleNode{child}
	} else {
		g.children[0] = child
	}
}

func (g *GrammarRuleNode) Snippet() *Snippet {
	return g.snippet
}

func (g *GrammarRuleNode) SetSnippet(snippet *Snippet) {
	g.snippet = snippet
}

func (g *GrammarRuleNode) Name() string {
	return g.name
}

func (g *GrammarRuleNode) SetName(name string) {
	g.name = name
}

func (g *GrammarRuleNode) RuleMemo() bool {
	return g.ruleMemo
}

func (g *GrammarRuleNode) SetRuleMemo(memo bool) {
	g.ruleMemo = memo
}

func (g *GrammarRuleNode) Separator() *GrammarRuleNode {
	return g.separator
}

func (g *GrammarRuleNode) SetSeparator(separator *GrammarRuleNode) {
	g.separator = separator
}

func (g *GrammarRuleNode) Action() *GrammarRuleNode {
	return g.action
}

func (g *GrammarRuleNode) SetAction(action *GrammarRuleNode) {
	g.action = action
}

func (g *GrammarRuleNode) Suffix() string {
	return g.suffix
}

func (g *GrammarRuleNode) SetSuffix(suffix string) {
	g.suffix = suffix
}
