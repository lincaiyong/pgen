package models

import (
	"github.com/lincaiyong/pgen/config"
	"strings"
)

type Language struct {
	name         string
	tokenRules   []*TokenRuleNode
	keywords     []string
	operators    []string
	astNodes     []*AstNode
	grammarRules []*GrammarRuleNode
	hackCode     string

	operatorMap map[string]string
	keywordMap  map[string]struct{}
	memoIdMap   map[*GrammarRuleNode]int
}

func (lang *Language) MemoIdMap() map[*GrammarRuleNode]int {
	return lang.memoIdMap
}

func (lang *Language) OperatorMap() map[string]string {
	return lang.operatorMap
}

func (lang *Language) KeywordMap() map[string]struct{} {
	return lang.keywordMap
}

func NewLanguage() *Language {
	return &Language{
		operatorMap: make(map[string]string),
		keywordMap:  make(map[string]struct{}),
		memoIdMap:   make(map[*GrammarRuleNode]int),
	}
}

func (lang *Language) Name() string {
	return lang.name
}

func (lang *Language) SetName(name string) {
	lang.name = name
}

func (lang *Language) Keywords() []string {
	return lang.keywords
}

func (lang *Language) AddKeyword(keyword string) {
	lang.keywords = append(lang.keywords, keyword)
	lang.keywordMap[keyword] = struct{}{}
}

func (lang *Language) Operators() []string {
	return lang.operators
}

func (lang *Language) AddOperator(operator string) {
	lang.operators = append(lang.operators, operator)
	opCharNames := config.OperatorCharName()
	names := make([]string, len(operator))
	for i, b := range []byte(operator) {
		names[i] = opCharNames[b]
	}
	lang.operatorMap[operator] = strings.Join(names, "_")
}

func (lang *Language) TokenRules() []*TokenRuleNode {
	return lang.tokenRules
}

func (lang *Language) AddTokenRule(token *TokenRuleNode) {
	lang.tokenRules = append(lang.tokenRules, token)
}

func (lang *Language) AstNodes() []*AstNode {
	return lang.astNodes
}

func (lang *Language) AddAstNode(node *AstNode) {
	lang.astNodes = append(lang.astNodes, node)
}

func (lang *Language) GrammarRules() []*GrammarRuleNode {
	return lang.grammarRules
}

func (lang *Language) AddGrammarRule(rule *GrammarRuleNode) {
	lang.grammarRules = append(lang.grammarRules, rule)
	if rule.RuleMemo() {
		lang.memoIdMap[rule] = len(lang.memoIdMap)
	}
}

func (lang *Language) HackCode() string {
	return lang.hackCode
}

func (lang *Language) SetHackCode(hack string) {
	lang.hackCode = hack
}
