package stages

import (
	"fmt"
	"github.com/lincaiyong/pgen/config"
	"github.com/lincaiyong/pgen/langparse"
	"github.com/lincaiyong/pgen/models"
	"regexp"
	"strings"
)

func RunStage2(stage1 *Stage1) *Stage2 {
	stage2 := &Stage2{
		Description: "parse into language struct",
		Input:       stage1,
		Language:    models.NewLanguage(),
		Error:       models.NewError(),
	}
	stage2.run()
	return stage2
}

type Stage2 struct {
	Description string
	Input       *Stage1
	Language    *models.Language
	Error       *models.Error
}

func (s *Stage2) run() {
	s.parseTokenRules()
	s.parseKeywords()
	s.parseOperators()
	s.parseNodes()
	s.parseGrammarRules()
	s.Language.SetHackCode(s.Input.Hack.Text())

	s.convertTokenRules()
	s.convertGrammarRules()
}

func (s *Stage2) parseTokenRules() {
	for _, snippet := range s.Input.Tokens {
		if strings.HasPrefix(snippet.Text(), "# ") {
			continue
		}
		rule, err := langparse.ParseTokenRule(snippet)
		if err != nil {
			s.Error.AddError(err)
		} else {
			s.Language.AddTokenRule(rule)
		}
	}
}

func (s *Stage2) parseKeywords() {
	for _, snippet := range s.Input.Keywords {
		text := strings.TrimSpace(snippet.Text())
		if strings.HasPrefix(text, "# ") {
			continue
		}
		if config.KeywordRegex().MatchString(text) {
			s.Language.AddKeyword(text)
		} else {
			s.Error.AddError(fmt.Errorf("invalid keyword %s at %d:%d", snippet.Text(), snippet.Start.LineIdx+1, snippet.End.LineIdx+1))
		}
	}
}

func (s *Stage2) parseOperators() {
	for _, snippet := range s.Input.Operators {
		text := strings.TrimSpace(snippet.Text())
		if strings.HasPrefix(text, "# ") {
			continue
		}
		if config.OperatorRegex().MatchString(text) {
			s.Language.AddOperator(text)
		} else {
			s.Error.AddError(fmt.Errorf("invalid operator %s at %d:%d", snippet.Text(), snippet.Start.LineIdx+1, snippet.End.LineIdx+1))
		}
	}
}

func (s *Stage2) parseNodes() {
	regex := regexp.MustCompile(" +")
	for _, snippet := range s.Input.Nodes {
		text := strings.TrimSpace(snippet.Text())
		if strings.HasPrefix(text, "# ") {
			continue
		}
		if m := config.NodeRegex().FindStringSubmatch(text); len(m) > 0 {
			args := strings.Split(regex.ReplaceAllString(m[2], " "), " ")
			node := models.NewAstNode(m[1], args, snippet)
			s.Language.AddAstNode(node)
		} else {
			s.Error.AddError(fmt.Errorf("invalid node %s at %d:%d", snippet.Text(), snippet.Start.LineIdx+1, snippet.End.LineIdx+1))
		}
	}
}

func (s *Stage2) parseGrammarRules() {
	for _, snippet := range s.Input.Grammars {
		if strings.HasPrefix(snippet.Text(), "# ") {
			continue
		}
		rule, err := langparse.ParseGrammarRule(snippet)
		if err != nil {
			s.Error.AddError(err)
		} else {
			s.Language.AddGrammarRule(rule)
		}
	}
}

func (s *Stage2) convertTokenRules() {
	atomNodes := make([]*models.TokenRuleNode, 0)
	for _, rule := range s.Language.TokenRules() {
		rule.Visit(func(node *models.TokenRuleNode) {
			if node.Kind() == models.TokenRuleNodeTypeGroupAtom {
				atomNodes = append(atomNodes, node)
			}
		})
	}
	newRules := make(map[string]*models.TokenRuleNode)
	for _, node := range atomNodes {
		node.SetKind(models.TokenRuleNodeTypeNameAtom)
		choices := node.Children()
		node.SetChildren(nil)
		key := node.Snippet().Text()
		if newRule := newRules[key]; newRule != nil {
			node.SetName(newRule.Name())
		} else {
			newRule = models.NewTokenRuleNode(models.TokenRuleNodeTypeRule, nil)
			newRule.SetChildren(choices)
			newRule.SetSnippet(node.Snippet())
			name := fmt.Sprintf("_group_%d", len(newRules)+1)
			newRule.SetName(name)
			node.SetName(name)
			newRules[key] = newRule
			s.Language.AddTokenRule(newRule)
		}
	}
}

func (s *Stage2) convertGrammarRules() {
	atomNodes := make([]*models.GrammarRuleNode, 0)
	for _, rule := range s.Language.GrammarRules() {
		rule.Visit(func(node *models.GrammarRuleNode) {
			if node.Kind() == models.GrammarRuleNodeTypeGroupAtom {
				if len(node.Children()) == 1 && node.Child().Action() == nil {
					node.SetChildren(node.Child().Children()) // atom -> items
					for _, child := range node.Children() {
						child.SetParent(node)
					}
				} else {
					atomNodes = append(atomNodes, node)
				}
			}
		})
	}
	newRules := make(map[string]*models.GrammarRuleNode)
	for _, node := range atomNodes {
		node.SetKind(models.GrammarRuleNodeTypeNameAtom)
		choices := node.Children()
		node.SetChildren(nil)
		key := node.Snippet().Text()
		if newRule := newRules[key]; newRule != nil {
			node.SetName(newRule.Name())
		} else {
			newRule = models.NewGrammarRuleNode(models.GrammarRuleNodeTypeRule, nil)
			newRule.SetChildren(choices)
			newRule.SetSnippet(node.Snippet())
			name := fmt.Sprintf("_group_%d", len(newRules)+1)
			newRule.SetName(name)
			node.SetName(name)
			newRules[key] = newRule
			s.Language.AddGrammarRule(newRule)
		}
	}
}
