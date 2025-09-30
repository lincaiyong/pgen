package langparse

import (
	"github.com/lincaiyong/pgen/models"
)

func ParseTokenRule(input *models.Snippet) (*models.TokenRuleNode, error) {
	parser := &TokenParser{
		BaseParser: NewBaseParser(input),
		Error:      models.NewError(),
	}
	parser.run()
	return parser.RuleNode, parser.Error.ToError()
}

type TokenParser struct {
	*BaseParser
	RuleNode *models.TokenRuleNode
	Error    *models.Error
}

func (p *TokenParser) run() {
	p.RuleNode = models.NewTokenRuleNode(models.TokenRuleNodeTypeRule, nil)

	p.skipWhitespace()
	start := p.mark()
	name := p.expectIdentifier()
	if name == nil {
		p.Error.AddError(p.expectError("token rule name"))
		return
	}
	p.skipWhitespace()
	if !p.expect(':') {
		p.Error.AddError(p.expectError(`":"`))
		return
	}
	var choices []*models.TokenRuleNode
	var err error
	choices, err = p.parseChoices(p.RuleNode)
	if err != nil {
		p.Error.AddError(err)
		return
	}
	end := p.mark()

	p.RuleNode.SetName(name.Text())
	p.RuleNode.SetSnippet(p.input.Fork(start, end))
	p.RuleNode.SetChildren(choices)

	p.skipWhitespace()
	if !p.reachEnd() {
		p.Error.AddError(p.expectError("EOF"))
	}
}

func (p *TokenParser) prefixOfAtom(b byte) bool {
	return b == '(' || b == '\'' || b == '[' || (b >= 'a' && b <= 'z') || b == '_'
}

func (p *TokenParser) prefixOfItem(b byte) bool {
	return p.prefixOfAtom(b) || b == '!' || b == '&'
}

func (p *TokenParser) parseChoices(parent *models.TokenRuleNode) ([]*models.TokenRuleNode, error) {
	choices := make([]*models.TokenRuleNode, 0)
	p.skipWhitespace()
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

func (p *TokenParser) parseChoice(parent *models.TokenRuleNode) (*models.TokenRuleNode, error) {
	p.skipWhitespace()
	choice := models.NewTokenRuleNode(models.TokenRuleNodeTypeChoice, parent)
	items := make([]*models.TokenRuleNode, 0)
	start, end := p.mark(), p.mark()
	for {
		if !p.prefixOfItem(p.la) {
			break
		}
		item, err := p.parseItem(choice)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
		end = p.mark()
		p.skipWhitespace()
	}
	p.reset(end)
	if start.SameAs(end) {
		return nil, p.expectError(`token item node`)
	}
	choice.SetChildren(items)
	choice.SetSnippet(p.input.Fork(start, end))
	return choice, nil
}

func (p *TokenParser) parseItem(parent *models.TokenRuleNode) (item *models.TokenRuleNode, err error) {
	p.skipWhitespace()
	item = models.NewTokenRuleNode(models.TokenRuleNodeTypeNegativeLookaheadItem, parent)
	var atom *models.TokenRuleNode
	start := p.mark()
	if p.expect('!') {
		atom, err = p.parseAtom(item)
		if err != nil {
			return nil, err
		}
		item.SetKind(models.TokenRuleNodeTypeNegativeLookaheadItem)
	} else if p.expect('&') {
		atom, err = p.parseAtom(item)
		if err != nil {
			return nil, err
		}
		item.SetKind(models.TokenRuleNodeTypePositiveLookaheadItem)
	} else {
		atom, err = p.parseAtom(item)
		if err != nil {
			return nil, err
		}
		if p.expect('?') {
			item.SetKind(models.TokenRuleNodeTypeOptionalItem)
		} else if p.expect('*') {
			item.SetKind(models.TokenRuleNodeTypeRepeat0Item)
		} else if p.expect('+') {
			item.SetKind(models.TokenRuleNodeTypeRepeat1Item)
		} else {
			item.SetKind(models.TokenRuleNodeTypeAtomItem)
		}
	}
	end := p.mark()
	item.SetSnippet(p.input.Fork(start, end))
	item.SetChild(atom)
	return item, nil
}

func (p *TokenParser) parseAtom(parent *models.TokenRuleNode) (*models.TokenRuleNode, error) {
	p.skipWhitespace()
	if p.la == '(' {
		return p.parseGroupAtom(parent)
	} else if p.la == '[' {
		return p.parseCharacterClassAtom(parent)
	} else if p.la == '\'' {
		return p.parseStringAtom(parent)
	} else if (p.la >= 'a' && p.la <= 'z') || p.la == '_' {
		return p.parseNameAtom(parent)
	} else {
		return nil, p.expectError(`atom prefix "[\[('a-z_]"`)
	}
}

func (p *TokenParser) parseGroupAtom(parent *models.TokenRuleNode) (*models.TokenRuleNode, error) {
	atom := models.NewTokenRuleNode(models.TokenRuleNodeTypeGroupAtom, parent)
	start := p.mark()
	p.stepForward()
	choices, err := p.parseChoices(atom)
	if err != nil {
		return nil, err
	}
	p.skipWhitespace()
	if !p.expect(')') {
		return nil, p.expectError(`")"`)
	}
	end := p.mark()
	atom.SetSnippet(p.input.Fork(start, end))
	atom.SetChildren(choices)
	return atom, nil
}

func (p *TokenParser) parseCharacterClassAtom(parent *models.TokenRuleNode) (*models.TokenRuleNode, error) {
	atom := models.NewTokenRuleNode(models.TokenRuleNodeTypeCharacterClassAtom, parent)
	start := p.mark()
	p.stepForward()
	p.forwardUtil(func(b byte) bool {
		return b == ']'
	})
	if !p.expect(']') {
		return nil, p.expectError(`"]"`)
	}
	end := p.mark()
	atom.SetSnippet(p.input.Fork(start, end))
	return atom, nil
}

func (p *TokenParser) parseNameAtom(parent *models.TokenRuleNode) (*models.TokenRuleNode, error) {
	atom := models.NewTokenRuleNode(models.TokenRuleNodeTypeNameAtom, parent)
	start, end := p.forwardUtil(func(b byte) bool {
		return !((b >= 'a' && b <= 'z') || b == '_')
	})
	atom.SetSnippet(p.input.Fork(start, end))
	atom.SetName(atom.Snippet().Text())
	return atom, nil
}

func (p *TokenParser) parseStringAtom(parent *models.TokenRuleNode) (*models.TokenRuleNode, error) {
	atom := models.NewTokenRuleNode(models.TokenRuleNodeTypeStringAtom, parent)
	start := p.mark()
	p.stepForward()
	afterBackslash := false
	p.forwardUtil(func(b byte) bool {
		if afterBackslash {
			afterBackslash = false
			return false
		}
		if b == '\\' {
			afterBackslash = true
			return false
		}
		return b == '\''
	})
	if !p.expect('\'') {
		return nil, p.expectError(`"'"`)
	}
	end := p.mark()
	atom.SetSnippet(p.input.Fork(start, end))
	return atom, nil
}
