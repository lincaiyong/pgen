package config

import "regexp"

var g struct {
	debugMode         bool
	reservedVariables map[string]struct{}
	builtinTokens     []string
	operatorCharName  map[byte]string
	operatorRegex     *regexp.Regexp
	keywordRegex      *regexp.Regexp
	nodeRegex         *regexp.Regexp
}

func DebugMode() bool {
	return g.debugMode
}

func ReservedVariables() map[string]struct{} {
	return g.reservedVariables
}

func BuiltinTokens() []string {
	return g.builtinTokens
}

func OperatorCharName() map[byte]string {
	return g.operatorCharName
}

func OperatorRegex() *regexp.Regexp {
	return g.operatorRegex
}

func KeywordRegex() *regexp.Regexp {
	return g.keywordRegex
}

func NodeRegex() *regexp.Regexp {
	return g.nodeRegex
}

func init() {
	g.debugMode = true
	g.reservedVariables = makeMap([]string{"_", "ps", "tk", "pos", "group"})
	g.builtinTokens = []string{"end_of_file", "pseudo", "whitespace", "newline"}
	g.operatorCharName = map[byte]string{
		'!':  "not", // exclamation
		'%':  "percent",
		'&':  "and", // ampersand
		'(':  "left_paren",
		')':  "right_paren",
		'*':  "star", // asterisk
		'+':  "plus",
		',':  "comma",
		'.':  "dot", // period
		'/':  "slash",
		':':  "colon",
		';':  "semi", // semicolon
		'<':  "less",
		'=':  "equal",
		'>':  "greater",
		'?':  "question",
		'@':  "at",
		'[':  "left_bracket",
		'\\': "back_slash",
		']':  "right_bracket",
		'^':  "caret",
		'{':  "left_brace",
		'|':  "bar",
		'}':  "right_brace",
		'~':  "tilde",
		'#':  "num_sign",
		'$':  "dollar",
		'-':  "minus",
	}
	g.operatorRegex = regexp.MustCompile(`^[!%&()*+,./:;<=>?@\[\\\]^{|}~#$-]+$`)
	g.keywordRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	g.nodeRegex = regexp.MustCompile(`^(\w+) +<([\w ]+)?>$`)
}

func makeMap(keys []string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, key := range keys {
		m[key] = struct{}{}
	}
	return m
}
