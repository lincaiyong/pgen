package stages

import (
	"fmt"
	"github.com/lincaiyong/pgen/langgen"
	"github.com/lincaiyong/pgen/models"
	"github.com/lincaiyong/pgen/snippet"
	"github.com/lincaiyong/pgen/util"
	"strings"
)

func RunStage31(s2 *Stage2) *Stage31 {
	stage3 := &Stage31{
		Description: "generate tokenizer code",
		Input:       s2,
		Gen:         langgen.NewGenerator(),
		Error:       models.NewError(),
	}
	stage3.run()
	return stage3
}

type Stage31 struct {
	Description string
	Input       *Stage2
	Gen         models.Generator
	Error       *models.Error
}

func (s *Stage31) run() {
	tokenizer := snippet.TokenizerStruct
	opCode := s.genTokenizerOpCode()
	tokenizer = strings.ReplaceAll(tokenizer, "<op_placeholder>", opCode)
	nextCode := s.genTokenizerNextCode()
	tokenizer = strings.ReplaceAll(tokenizer, "<next_placeholder>", nextCode)
	s.Gen.Put(tokenizer).PutNL()
	s.tokenizerInitKeywords().PutNL()
	for _, rule := range s.Input.Language.TokenRules() {
		err := s.genTokenRuleCode(rule)
		if err != nil {
			s.Error.AddError(err)
		} else {
			s.Gen.PutNL()
		}
	}
	s.genTokenizerOpCode()
}

func (s *Stage31) tokenizerInitKeywords() models.Generator {
	s.Gen.Put("func (tk *Tokenizer) initKeywords() {").Push()
	s.Gen.Put("tk._keywords = make(map[string]string)")
	for _, keyword := range s.Input.Language.Keywords() {
		s.Gen.Put(`tk._keywords["%s"] = TokenTypeKw%s`, keyword, util.ToPascalCase(keyword))
	}
	s.Gen.Pop().Put("}")
	return s.Gen
}

func (s *Stage31) genTokenizerOpCode() string {
	gen := langgen.NewGenerator()
	gen.PutNL().Push()
	opTree := NewOperatorNode(0, 0)
	for _, op := range s.Input.Language.Operators() {
		name := s.Input.Language.OperatorMap()[op]
		opTree.Update(name, []byte(op))
	}
	opTree.GenCode(gen)
	return gen.String()
}

func (s *Stage31) genTokenizerNextCode() string {
	gen := langgen.NewGenerator()
	gen.PutNL().Push()
	for _, rule := range s.Input.Language.TokenRules() {
		if !strings.HasPrefix(rule.Name(), "_") {
			gen.Put("} else if tk.%s() {", util.SafeName(util.ToCamelCase(rule.Name()))).Push()
			gen.Put("kind = TokenType%s", util.ToPascalCase(rule.Name())).Pop()
		}
	}
	return gen.String()
}

func (s *Stage31) genTokenRuleCode(rule *models.TokenRuleNode) error {
	s.Gen.ClearVar()
	s.Gen.Put("// %s:", rule.Name())
	s.Gen.Put("//")
	for _, choice := range rule.Children() {
		s.Gen.Put("//\t| %s", strings.ReplaceAll(choice.Snippet().Text(), "\n", " "))
	}
	rule.Visit(func(node *models.TokenRuleNode) {
		if node.Kind() == models.TokenRuleNodeTypeNameAtom && strings.HasPrefix(node.Name(), "_group_") {
			s.Gen.Put("//\t%s <-- %s", node.Name(), node.Snippet().Text())
		}
	})

	s.Gen.Put("func (tk *Tokenizer) %s() bool {", util.SafeName(util.ToCamelCase(rule.Name()))).Push()
	posVarDefined := ""
	for _, choice := range rule.Children() {
		s.Gen.Put("// %s", strings.ReplaceAll(choice.Snippet().Text(), "\n", " "))
		count := 0
		for _, item := range choice.Children() {
			if item.Kind() != models.TokenRuleNodeTypeNegativeLookaheadItem && item.Kind() != models.TokenRuleNodeTypePositiveLookaheadItem {
				count++
			}
		}

		if count <= 1 {
			depth, err := s.genEnterCode(choice, 0)
			if err != nil {
				return err
			}
			s.Gen.Put("return true")
			s.genLeaveCode(choice, depth)
		} else {
			if posVarDefined != "" {
				s.Gen.Put("%s = tk._mark()", posVarDefined)
			} else {
				posVarDefined = s.Gen.CreateVar("p")
				s.Gen.Put("%s := tk._mark()", posVarDefined)
			}
			depth, err := s.genEnterCode(choice, 0)
			if err != nil {
				return err
			}
			s.Gen.Put("return true")
			s.genLeaveCode(choice, depth)
			s.Gen.Put("tk._reset(%s)", posVarDefined)
		}
	}
	s.Gen.Put("return false")
	s.Gen.Pop().Put("}")
	return nil
}

func (s *Stage31) genEnterCode(node *models.TokenRuleNode, depth int) (int, error) {
	var err error
	switch node.Kind() {
	case models.TokenRuleNodeTypeChoice:
		for _, item := range node.Children() {
			_, err = s.genEnterCode(item, depth)
			if err != nil {
				return 0, err
			}
		}
		return depth, nil
	case models.TokenRuleNodeTypeOptionalItem:
		_, err = s.genEnterCode(node.Child(), depth)
		if err != nil {
			return 0, err
		}
		s.genLeaveCode(node.Child(), depth)
		return depth, nil
	case models.TokenRuleNodeTypeRepeat0Item, models.TokenRuleNodeTypeRepeat1Item:
		if node.Kind() == models.TokenRuleNodeTypeRepeat1Item {
			_, err = s.genEnterCode(node.Child(), depth)
			if err != nil {
				return 0, err
			}
		}
		okVar := s.Gen.CreateVar("ok")
		s.Gen.Put("for {").Push()
		s.Gen.Put("%s := false", okVar)
		_, err = s.genEnterCode(node.Child(), depth)
		if err != nil {
			return 0, err
		}
		s.Gen.Put("%s = true", okVar)
		s.genLeaveCode(node.Child(), depth)
		s.Gen.Put("if !%s {", okVar).Push()
		s.Gen.Put("break")
		s.Gen.Pop().Put("}")
		s.Gen.Pop().Put("}")
		return depth, nil
	case models.TokenRuleNodeTypeNegativeLookaheadItem:
		posVar := s.Gen.CreateVar("p")
		okVar := s.Gen.CreateVar("ok")
		s.Gen.Put("%s := tk._mark()", posVar)
		s.Gen.Put("%s := false", okVar)
		_, err = s.genEnterCode(node.Child(), depth)
		if err != nil {
			return 0, err
		}
		s.Gen.Put("%s = true", okVar)
		s.genLeaveCode(node.Child(), depth)
		s.Gen.Put("tk._reset(%s)", posVar)
		s.Gen.Put("if !%s {", okVar).Push()
		return depth + 1, nil
	case models.TokenRuleNodeTypePositiveLookaheadItem:
		posVar := s.Gen.CreateVar("p")
		okVar := s.Gen.CreateVar("ok")
		s.Gen.Put("%s := tk._mark()", posVar)
		s.Gen.Put("%s := false", okVar)
		_, err = s.genEnterCode(node.Child(), depth)
		if err != nil {
			return 0, err
		}
		s.Gen.Put("%s = true", okVar)
		s.genLeaveCode(node.Child(), depth)
		s.Gen.Put("tk._reset(%s)", posVar)
		s.Gen.Put("if %s {", okVar).Push()
		return depth + 1, nil
	case models.TokenRuleNodeTypeAtomItem:
		_, err = s.genEnterCode(node.Child(), depth)
		if err != nil {
			return 0, err
		}
		return depth, nil
	case models.TokenRuleNodeTypeNameAtom:
		name := node.Name()
		name = util.SafeName(util.ToCamelCase(name))
		s.Gen.Put("if tk.%s() {", name).Push()
		return depth + 1, nil
	case models.TokenRuleNodeTypeStringAtom:
		val := node.Snippet().Text()
		val = val[1 : len(val)-1]
		if len(val) == 1 {
			s.Gen.Put("if tk._expect(0x%X) {", val[0]).Push()
		} else {
			raw := util.SingleQuoteStringUnescape(val)
			val2 := util.DoubleQuoteStringEscape(raw)
			s.Gen.Put("if tk._expectS(\"%s\") {", val2).Push()
		}
		return depth + 1, nil
	case models.TokenRuleNodeTypeCharacterClassAtom:
		val := node.Snippet().Text()
		val = val[1 : len(val)-1]
		var ret [][]rune
		ret, err = util.ParseCharacterClass(val)
		if err != nil {
			return 0, err
		}
		conditions := make([]string, 0)
		for _, pair := range ret {
			if len(pair) == 1 {
				conditions = append(conditions, fmt.Sprintf("tk._expect(0x%X)", pair[0]))
			} else {
				conditions = append(conditions, fmt.Sprintf("tk._expectR(0x%X, 0x%X)", pair[0], pair[1]))
			}
		}
		s.Gen.Put("if %s {", strings.Join(conditions, " || ")).Push()
		return depth + 1, nil
	default:
		panic("unreachable")
	}
	return depth, nil
}

func (s *Stage31) genLeaveCode(node *models.TokenRuleNode, depth int) int {
	switch node.Kind() {
	case models.TokenRuleNodeTypeChoice:
		for i := len(node.Children()) - 1; i >= 0; i-- {
			item := node.Children()[i]
			depth = s.genLeaveCode(item, depth)
		}
		return depth
	case models.TokenRuleNodeTypeOptionalItem:
		return depth
	case models.TokenRuleNodeTypeRepeat0Item, models.TokenRuleNodeTypeRepeat1Item:
		if node.Kind() == models.TokenRuleNodeTypeRepeat1Item {
			s.genLeaveCode(node.Child(), depth)
		}
		return depth
	case models.TokenRuleNodeTypeNegativeLookaheadItem:
		s.Gen.Pop().Put("}")
		return depth - 1
	case models.TokenRuleNodeTypePositiveLookaheadItem:
		s.Gen.Pop().Put("}")
		return depth - 1
	case models.TokenRuleNodeTypeAtomItem:
		s.genLeaveCode(node.Child(), depth)
		return depth
	case models.TokenRuleNodeTypeNameAtom, models.TokenRuleNodeTypeCharacterClassAtom, models.TokenRuleNodeTypeStringAtom:
		s.Gen.Pop().Put("}")
		return depth - 1
	default:
		panic("this should never happen")
	}
	return depth
}
