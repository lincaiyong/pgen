package langparse

import (
	"github.com/lincaiyong/pgen/models"
)

func ParseGrammarRule(input *models.Snippet) (*models.GrammarRuleNode, error) {
	parser := &GrammarParser{
		BaseParser: NewBaseParser(input),
		Error:      models.NewError(),
	}
	parser.run()
	return parser.RuleNode, parser.Error.ToError()
}

type GrammarParser struct {
	*BaseParser
	RuleNode *models.GrammarRuleNode
	Error    *models.Error
}

func (p *GrammarParser) run() {
	p.RuleNode = models.NewGrammarRuleNode(models.GrammarRuleNodeTypeRule, nil)

	// name
	p.skipWhitespace()
	start := p.mark()
	name := p.expectIdentifier()
	if name == nil {
		p.Error.AddError(p.expectError("grammar rule name"))
		return
	}
	p.RuleNode.SetName(name.Text())
	// memo
	p.skipWhitespace()
	if p.expectString("(memo)") {
		p.RuleNode.SetRuleMemo(true)
	}
	// :
	p.skipWhitespace()
	if !p.expect(':') {
		p.Error.AddError(p.expectError(`":"`))
		return
	}
	// choices
	var choices []*models.GrammarRuleNode
	var err error
	choices, err = p.parseChoices(p.RuleNode)
	if err != nil {
		p.Error.AddError(err)
		return
	}
	p.RuleNode.SetChildren(choices)

	end := p.mark()
	p.RuleNode.SetSnippet(p.input.Fork(start, end))

	p.skipWhitespace()
	if !p.reachEnd() {
		p.Error.AddError(p.expectError("EOF"))
	}
}

func (p *GrammarParser) parseChoices(parent *models.GrammarRuleNode) ([]*models.GrammarRuleNode, error) {
	p.skipWhitespace()
	choices := make([]*models.GrammarRuleNode, 0)
	p.expect('|')
	var end models.Position
	for {
		choice, err := p.parseChoice(parent)
		if err != nil {
			return nil, err
		}
		choices = append(choices, choice)
		end = p.mark()

		p.skipWhitespace()
		if p.expect('|') {
			continue
		}
		inGroup := p.RuleNode != parent
		if (!inGroup && p.reachEnd()) || (inGroup && p.la == ')') {
			break
		}
	}
	p.reset(end)
	return choices, nil
}

func (p *GrammarParser) parseChoice(parent *models.GrammarRuleNode) (*models.GrammarRuleNode, error) {
	p.skipWhitespace()
	choice := models.NewGrammarRuleNode(models.GrammarRuleNodeTypeChoice, parent)
	start := p.mark()

	err := p.parseChoiceRule(choice)
	if err != nil {
		return nil, err
	}
	end := p.mark()

	p.skipWhitespace()
	if p.la == '{' {
		var choiceAction *models.GrammarRuleNode
		choiceAction, err = p.parseChoiceAction(choice)
		if err != nil {
			return nil, err
		}
		choice.SetAction(choiceAction)
		end = p.mark()
	}

	choice.SetSnippet(p.input.Fork(start, end))
	return choice, nil
}

func (p *GrammarParser) prefixOfAtom(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_' || b == '\'' || b == '('
}

func (p *GrammarParser) prefixOfItem(b byte) bool {
	return b == '[' || b == ']' || b == '~' || b == '&' || b == '!' || p.prefixOfAtom(b)
}

func (p *GrammarParser) parseChoiceRule(choice *models.GrammarRuleNode) error {
	p.skipWhitespace()
	items := make([]*models.GrammarRuleNode, 0)
	start, end := p.mark(), p.mark()
	for {
		if !p.prefixOfItem(p.la) {
			break
		}
		item, err := p.parseItem(choice)
		if err != nil {
			return err
		}
		items = append(items, item)
		end = p.mark()
		p.skipWhitespace()
	}
	p.reset(end)
	if start.SameAs(end) {
		return p.expectError(`grammar item node`)
	}
	choice.SetChildren(items)
	choice.SetSnippet(p.input.Fork(start, end))
	return nil
}

func (p *GrammarParser) parseChoiceAction(parent *models.GrammarRuleNode) (*models.GrammarRuleNode, error) {
	p.stepForward()
	action, err := p.parseActionExpr(parent)
	if err != nil {
		return nil, err
	}
	p.skipWhitespace()
	if !p.expect('}') {
		return nil, p.expectError("'}'")
	}
	return action, nil
}

func (p *GrammarParser) parseActionExpr(parent *models.GrammarRuleNode) (*models.GrammarRuleNode, error) {
	p.skipWhitespace()
	action := models.NewGrammarRuleNode("", parent)
	start := p.mark()
	if p.la == '_' {
		p.stepForward()
		if !(p.la >= 'a' && p.la <= 'z') {
			action.SetKind(models.GrammarRuleNodeTypeNullAction)
		} else {
			p.reset(start)
		}
	}
	if action.Kind() == "" {
		if p.expect('[') {
			action.SetKind(models.GrammarRuleNodeTypeListAction)
			elem, err := p.parseActionExpr(action)
			if err != nil {
				return nil, err
			}
			if !p.expect(']') {
				return nil, p.expectError("']'")
			}
			action.SetChild(elem)
		} else if p.la == '_' || (p.la >= 'a' && p.la <= 'z') {
			if tmp, err := p.parseCallActionExpr(parent); err != nil {
				return nil, err
			} else if tmp != nil {
				action = tmp
			} else {
				p.expectIdentifier()
				action.SetKind(models.GrammarRuleNodeTypeNameAction)
			}
		} else {
			return nil, p.expectError("action prefix [\\[_a-z]")
		}
	}

	end := p.mark()
	action.SetSnippet(p.input.Fork(start, end))
	return action, nil
}

func (p *GrammarParser) parseCallActionExpr(parent *models.GrammarRuleNode) (*models.GrammarRuleNode, error) {
	pos := p.mark()
	name := p.expectIdentifier()
	if !p.expect('(') {
		p.reset(pos)
		return nil, nil
	}

	callAction := models.NewGrammarRuleNode(models.GrammarRuleNodeTypeCallAction, parent)
	callAction.SetName(name.Text())
	args := make([]*models.GrammarRuleNode, 0)
	for {
		if p.la == ')' {
			break
		}
		if len(args) > 0 {
			p.skipWhitespace()
			if !p.expect(',') {
				return nil, p.expectError("','")
			}
		}
		arg, err := p.parseActionExpr(callAction)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}
	callAction.SetChildren(args)
	p.stepForward()
	return callAction, nil
}

func (p *GrammarParser) tryParseItemName() *models.Snippet {
	pos := p.mark()
	snippet := p.expectIdentifier()
	if snippet == nil {
		return nil
	}
	if !p.expect('=') {
		p.reset(pos)
		return nil
	}
	return snippet
}

func (p *GrammarParser) parseItem(parent *models.GrammarRuleNode) (*models.GrammarRuleNode, error) {
	item := models.NewGrammarRuleNode("", parent)
	start := p.mark()
	if name := p.tryParseItemName(); name != nil {
		item.SetName(name.Text())
	}
	var atom *models.GrammarRuleNode
	var err error
	if p.expect('!') {
		item.SetKind(models.GrammarRuleNodeTypeNegativeLookaheadItem)
		atom, err = p.parseAtom(item)
	} else if p.expect('&') {
		item.SetKind(models.GrammarRuleNodeTypePositiveLookaheadItem)
		atom, err = p.parseAtom(item)
	} else if p.expect('~') {
		item.SetKind(models.GrammarRuleNodeTypeForwardIfNotMatchItem)
		atom, err = p.parseAtom(item)
	} else {
		atom, err = p.parseAtom(item)
		if err != nil {
			return nil, err
		}
		if p.expect('?') {
			item.SetKind(models.GrammarRuleNodeTypeOptionalItem)
		} else if p.expect('*') {
			item.SetKind(models.GrammarRuleNodeTypeRepeat0Item)
		} else if p.expect('+') {
			item.SetKind(models.GrammarRuleNodeTypeRepeat1Item)
		} else if p.expect('.') {
			item.SetSeparator(atom)
			atom, err = p.parseAtom(item)
			if err != nil {
				return nil, err
			}
			if p.expect('*') {
				item.SetKind(models.GrammarRuleNodeTypeSeparatedRepeat0Item)
			} else if p.expect('+') {
				item.SetKind(models.GrammarRuleNodeTypeSeparatedRepeat1Item)
			} else {
				return nil, p.expectError("'*' or '+'")
			}
		} else {
			item.SetKind(models.GrammarRuleNodeTypeAtomItem)
		}
	}
	if err != nil {
		return nil, err
	}
	item.SetChild(atom)
	end := p.mark()
	item.SetSnippet(p.input.Fork(start, end))
	// suffix
	p.skipWhitespace()
	if p.la == '[' || p.la == ']' {
		item.SetSuffix(string(p.la))
		p.stepForward()
	} else {
		p.reset(end)
	}
	return item, nil
}

func (p *GrammarParser) parseAtom(parent *models.GrammarRuleNode) (*models.GrammarRuleNode, error) {
	atom := p.tryParseBracketEllipsisAtom(parent)
	if atom != nil {
		return atom, nil
	}
	if p.la == '(' {
		return p.parseGroupAtom(parent)
	} else if p.la == '\'' {
		return p.parseStringAtom(parent)
	} else if (p.la >= 'a' && p.la <= 'z') || p.la == '_' {
		return p.parseNameAtom(parent)
	} else if p.la >= 'A' && p.la <= 'Z' {
		return p.parseTokenAtom(parent)
	} else {
		return nil, p.expectError("atom prefix ['(a-zA-Z_]")
	}
}

func (p *GrammarParser) parseStringAtom(parent *models.GrammarRuleNode) (*models.GrammarRuleNode, error) {
	atom := models.NewGrammarRuleNode(models.GrammarRuleNodeTypeStringAtom, parent)
	start := p.mark()
	p.stepForward()
	var prev byte
	p.forwardUtil(func(b byte) bool {
		if prev != '\\' && b == '\'' {
			return true
		}
		prev = b
		return false
	})
	if !p.expect('\'') {
		return nil, p.expectError("'")
	}
	end := p.mark()
	atom.SetSnippet(p.input.Fork(start, end))
	return atom, nil
}

func (p *GrammarParser) parseNameAtom(parent *models.GrammarRuleNode) (*models.GrammarRuleNode, error) {
	atom := models.NewGrammarRuleNode(models.GrammarRuleNodeTypeNameAtom, parent)
	start, end := p.forwardUtil(func(b byte) bool {
		return !((b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') || b == '_')
	})
	atom.SetSnippet(p.input.Fork(start, end))
	atom.SetName(atom.Snippet().Text())
	return atom, nil
}

func (p *GrammarParser) parseTokenAtom(parent *models.GrammarRuleNode) (*models.GrammarRuleNode, error) {
	atom := models.NewGrammarRuleNode(models.GrammarRuleNodeTypeTokenAtom, parent)
	start, end := p.forwardUtil(func(b byte) bool {
		return !((b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_')
	})
	atom.SetSnippet(p.input.Fork(start, end))
	return atom, nil
}

func (p *GrammarParser) tryParseBracketEllipsisAtom(parent *models.GrammarRuleNode) *models.GrammarRuleNode {
	atom := models.NewGrammarRuleNode(models.GrammarRuleNodeTypeBracketEllipsisAtom, parent)
	start := p.mark()
	if p.expectString("'('...')'") || p.expectString("'['...']'") || p.expectString("'{'...'}'") {
		end := p.mark()
		atom.SetSnippet(p.input.Fork(start, end))
		return atom
	}
	return nil
}

func (p *GrammarParser) parseGroupAtom(parent *models.GrammarRuleNode) (*models.GrammarRuleNode, error) {
	atom := models.NewGrammarRuleNode(models.GrammarRuleNodeTypeGroupAtom, parent)
	start := p.mark()
	p.stepForward()
	choices, err := p.parseChoices(atom)
	if err != nil {
		return nil, err
	}
	if !p.expect(')') {
		return nil, p.expectError("')'")
	}
	end := p.mark()
	atom.SetSnippet(p.input.Fork(start, end))
	atom.SetChildren(choices)
	return atom, nil
}
