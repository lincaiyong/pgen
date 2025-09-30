package stages

import (
	"fmt"
	"github.com/lincaiyong/pgen/langgen"
	"github.com/lincaiyong/pgen/models"
	"github.com/lincaiyong/pgen/snippet"
	"github.com/lincaiyong/pgen/util"
	"regexp"
	"sort"
	"strings"
)

func RunStage32(s2 *Stage2) *Stage32 {
	stage3 := &Stage32{
		Description: "generate parser code",
		Input:       s2,
		Gen:         langgen.NewGenerator(),
		Error:       models.NewError(),
	}
	stage3.run()
	return stage3
}

type Stage32 struct {
	Description string
	Input       *Stage2
	Gen         models.Generator
	Error       *models.Error
}

func (s *Stage32) run() {
	s.genMemoIdConsts().PutNL()
	s.Gen.Put(snippet.NodeCacheStruct).PutNL()
	s.Gen.Put(snippet.ParserStruct).PutNL()
	for _, rule := range s.Input.Language.GrammarRules() {
		err := s.genGrammarRuleCode(rule)
		if err != nil {
			s.Error.AddError(err)
		}
	}
}

func (s *Stage32) genGrammarRuleCode(rule *models.GrammarRuleNode) error {
	leftRecChoices := make([]*models.GrammarRuleNode, 0)
	simpleChoices := make([]*models.GrammarRuleNode, 0)
	for _, choice := range rule.Children() {
		leftmost := make(map[string]bool)
		s.gramLeftMost(choice, leftmost)
		if _, ok := leftmost[rule.Name()]; ok {
			leftRecChoices = append(leftRecChoices, choice)
		} else {
			simpleChoices = append(simpleChoices, choice)
		}
	}
	if len(leftRecChoices) > 0 {
		s.gramLeftRecRuleCode(rule, leftRecChoices, simpleChoices)
	} else {
		s.gramSimpleRuleCode(rule)
	}
	return nil
}

func (s *Stage32) gramLeftMost(node *models.GrammarRuleNode, leftmost map[string]bool) (cont bool) {
	switch node.Kind() {
	case models.GrammarRuleNodeTypeChoice:
		for _, item := range node.Children() {
			if cont = s.gramLeftMost(item, leftmost); !cont {
				break
			}
		}
		return false
	case models.GrammarRuleNodeTypeOptionalItem, models.GrammarRuleNodeTypeRepeat0Item,
		models.GrammarRuleNodeTypeSeparatedRepeat0Item, models.GrammarRuleNodeTypeNegativeLookaheadItem:
		if node.Child() != nil {
			s.gramLeftMost(node.Child(), leftmost)
		}
		return true
	case models.GrammarRuleNodeTypeRepeat1Item, models.GrammarRuleNodeTypeAtomItem,
		models.GrammarRuleNodeTypeSeparatedRepeat1Item, models.GrammarRuleNodeTypePositiveLookaheadItem:
		if node.Child() != nil {
			s.gramLeftMost(node.Child(), leftmost)
		}
		return false
	case models.GrammarRuleNodeTypeNameAtom:
		leftmost[node.Snippet().Text()] = true
		return false
	default:
		return false
	}
}

func (s *Stage32) gramMemoCode(funName string) {
	s.Gen.Put("func (ps *Parser) %s() Node {", funName).Push()
	s.Gen.Put("pos := ps._mark()")
	s.Gen.Put("var ok bool")
	s.Gen.Put("var cache *NodeCache")
	s.Gen.Put("cacheAtPos := ps._nodeCache[pos]")
	s.Gen.Put("if cacheAtPos != nil {").Push()
	s.Gen.Put("if cache, ok = cacheAtPos[%sMemoId]; ok {", funName).Push()
	s.Gen.Put("if cache.val == nil {").Push()
	s.Gen.Put("return nil").Pop()
	s.Gen.Put("}")
	s.Gen.Put("ps._reset(cache.pos)")
	s.Gen.Put("return cache.val").Pop()
	s.Gen.Put("}").Pop()
	s.Gen.Put("} else {").Push()
	s.Gen.Put("cacheAtPos = make(map[int]*NodeCache)")
	s.Gen.Put("ps._nodeCache[pos] = cacheAtPos").Pop()
	s.Gen.Put("}")
	s.Gen.Put("t := ps.%s_()", funName)
	s.Gen.Put("cacheAtPos[%sMemoId] = &NodeCache{t, ps._mark()}", funName)
	s.Gen.Put("return t").Pop()
	s.Gen.Put("}").PutNL()
}

func (s *Stage32) gramSimpleRuleCode(rule *models.GrammarRuleNode) {
	memo := ""
	funName := util.SafeName(util.ToCamelCase(rule.Name()))
	if rule.RuleMemo() {
		s.gramMemoCode(funName)
		memo = "!"
		funName += "_"
	}

	s.Gen.Put("/*\n%s%s:", rule.Name(), memo)
	for _, choice := range rule.Children() {
		s.Gen.Put("| %s", choice.Snippet().Text())
	}
	rule.Visit(func(node *models.GrammarRuleNode) {
		if node.Kind() == models.GrammarRuleNodeTypeNameAtom && strings.HasPrefix(node.Name(), "_group_") {
			s.Gen.Put("%s <-- %s", node.Name(), node.Snippet().Text())
		}
	})
	s.Gen.Put("*/")

	s.Gen.Put("func (ps *Parser) %s() Node {", funName).Push()
	s.gramChoicesCode(rule.Children(), "")
	s.Gen.Put("return nil")
	s.Gen.Pop().Put("}").PutNL()
}

func (s *Stage32) gramLeftRecRuleCode(rule *models.GrammarRuleNode, leftRecChoices, simpleChoices []*models.GrammarRuleNode) {
	memo := ""
	funName := util.SafeName(util.ToCamelCase(rule.Name()))
	if rule.RuleMemo() {
		s.gramMemoCode(funName)
		memo = "!"
		funName += "_"
	}

	s.Gen.Put("/*\n%s%s:", rule.Name(), memo)
	for _, choice := range rule.Children() {
		s.Gen.Put("| %s", choice.Snippet().Text())
	}
	rule.Visit(func(node *models.GrammarRuleNode) {
		if node.Kind() == models.GrammarRuleNodeTypeNameAtom && strings.HasPrefix(node.Name(), "_group_") {
			s.Gen.Put("%s <-- %s", node.Name(), node.Snippet().Text())
		}
	})
	s.Gen.Put("*/")

	camelName := util.ToCamelCase(rule.Name())
	s.Gen.Put("func (ps *Parser) %s() Node {", funName).Push()
	s.Gen.Put("_left := ps.%sLeftMost()", camelName)
	s.Gen.Put("if _left == nil {").Push()
	s.Gen.Put("return nil")
	s.Gen.Pop().Put("}")
	s.Gen.Put("_ret := ps.%sRightPart(_left)", camelName)
	s.Gen.Put("for _ret != nil {").Push()
	s.Gen.Put("_left = _ret")
	s.Gen.Put("_ret = ps.%sRightPart(_left)", camelName)
	s.Gen.Pop().Put("}")
	s.Gen.Put("return _left")
	s.Gen.Pop().Put("}").PutNL()

	s.Gen.Put("func (ps *Parser) %sLeftMost() Node {", camelName).Push()
	s.gramChoicesCode(simpleChoices, "")
	s.Gen.Put("return nil")
	s.Gen.Pop().Put("}").PutNL()

	s.Gen.Put("func (ps *Parser) %sRightPart(_left Node) Node {", camelName).Push()
	s.gramChoicesCode(leftRecChoices, "_left")
	s.Gen.Put("return nil")
	s.Gen.Pop().Put("}").PutNL()
}

func (s *Stage32) gramChoicesCode(choices []*models.GrammarRuleNode, leftVar string) {
	posDefined := false
	for _, choice := range choices {
		s.Gen.Put("/* %s", regexp.MustCompile(`\s+`).ReplaceAllString(choice.Snippet().Text(), " "))
		s.Gen.Put(" */")
		needMarkReset := len(choice.Children()) > 1 || choice.Action() != nil
		if needMarkReset {
			if !posDefined {
				posDefined = true
				s.Gen.Put("pos := ps._mark()")
			}
		}

		s.gramCode(choice, "", leftVar)

		if needMarkReset {
			s.Gen.Put("ps._reset(pos)")
		}
	}
}

func (s *Stage32) gramHoistingGroupVars(atom *models.GrammarRuleNode) {
	if atom != nil && atom.Kind() == models.GrammarRuleNodeTypeGroupAtom {
		for _, item := range atom.Child().Children() {
			if item.Name() != "" {
				s.Gen.Put("var %s Node", item.Name())
			}
		}
	}
}

func (s *Stage32) gramItemNames(node *models.GrammarRuleNode, names []string) []string {
	switch node.Kind() {
	case models.GrammarRuleNodeTypeChoice:
		for _, item := range node.Children() {
			names = s.gramItemNames(item, names)
		}
	case models.GrammarRuleNodeTypeOptionalItem, models.GrammarRuleNodeTypeRepeat0Item,
		models.GrammarRuleNodeTypeSeparatedRepeat0Item, models.GrammarRuleNodeTypeNegativeLookaheadItem,
		models.GrammarRuleNodeTypeRepeat1Item, models.GrammarRuleNodeTypeAtomItem,
		models.GrammarRuleNodeTypeSeparatedRepeat1Item, models.GrammarRuleNodeTypePositiveLookaheadItem:
		if node.Name() != "" {
			names = append(names, node.Name())
		}
		if node.Child() != nil {
			names = s.gramItemNames(node.Child(), names)
		}
		return names
	case models.GrammarRuleNodeTypeGroupAtom:
		for _, item := range node.Children() {
			names = s.gramItemNames(item, names)
		}
	}
	return names
}

func (s *Stage32) gramCode(node *models.GrammarRuleNode, itemName string, leftVar string) {
	switch node.Kind() {
	case models.GrammarRuleNodeTypeChoice:
		s.Gen.ClearVar()
		s.Gen.Put("for {").Push()
		names := s.gramItemNames(node, []string{})
		sort.Strings(names)
		for _, name := range names {
			s.Gen.Put("var %s Node", name)
		}
		var breakVar string
		for i, item := range node.Children() {
			if leftVar != "" && i == 0 {
				s.Gen.Put("%s = %s", item.Name(), leftVar) // FIXME: 是不是name可能为空
			} else {
				s.gramCode(item, item.Name(), "")
				if item.Suffix() == "[" {
					breakVar = s.Gen.CreateVar("break")
					s.Gen.Put("%s := true", breakVar)
					s.Gen.Put("ps._enter()")
					s.Gen.Put("for {").Push()
				} else if item.Suffix() == "]" {
					s.Gen.Put("%s = false", breakVar)
					s.Gen.Put("break")
					s.Gen.Pop().Put("}")
					s.Gen.Put("ps._leave()")
					s.Gen.Put("if %s {", breakVar).Push()
					s.Gen.Put("break")
					s.Gen.Pop().Put("}")
				}
			}
		}
		if node.Action() == nil {
			s.Gen.Put("return _1")
		} else if node.Action().Kind() == models.GrammarRuleNodeTypeNullAction {
			s.Gen.Put("return dummyNode")
		} else {
			action := s.gramActionCode(node.Action(), leftVar)
			//if strings.Contains(action, "Node(") {
			//	s.Gen.Put(`ps.log("match %s", pos)`, n.ruleBelongTo.name.val)
			//}
			s.Gen.Put("return %s", action)
		}
		s.Gen.Pop().Put("}")
	case models.GrammarRuleNodeTypeOptionalItem:
		if itemName == "" {
			itemName = s.Gen.CreateVar("_")
			s.Gen.Put("var %s Node", itemName)
		}
		s.gramCode(node.Child(), itemName, "")
		s.Gen.Put("_ = %s", itemName)
	case models.GrammarRuleNodeTypeRepeat0Item:
		if itemName == "" {
			itemName = s.Gen.CreateVar("_")
			s.Gen.Put("var %s Node", itemName)
		}
		tmpVar := s.Gen.CreateVar("_")
		s.Gen.Put("%s := make([]Node, 0)", tmpVar)
		itemVar := s.Gen.CreateVar("_")
		s.Gen.Put("var %s Node", itemVar)
		s.Gen.Put("for {").Push()
		s.gramCode(node.Child(), itemVar, "")
		s.Gen.Put("if %s == nil {", itemVar).Push()
		s.Gen.Put("break")
		s.Gen.Pop().Put("}")
		s.Gen.Put("%s = append(%s, %s)", tmpVar, tmpVar, itemVar)
		s.Gen.Pop().Put("}")
		s.Gen.Put("%s = NewNodesNode(%s)", itemName, tmpVar)
		s.Gen.Put("_ = %s", itemName)
	case models.GrammarRuleNodeTypeRepeat1Item:
		if itemName == "" {
			itemName = s.Gen.CreateVar("_")
			s.Gen.Put("var %s Node", itemName)
		}
		tmpVar := s.Gen.CreateVar("_")
		s.Gen.Put("%s := make([]Node, 0)", tmpVar)
		itemVar := s.Gen.CreateVar("_")
		s.Gen.Put("var %s Node", itemVar)
		s.gramCode(node.Child(), itemVar, "")
		s.Gen.Put("if %s == nil {", itemVar).Push()
		s.Gen.Put("break")
		s.Gen.Pop().Put("}")
		s.Gen.Put("%s = append(%s, %s)", tmpVar, tmpVar, itemVar)
		s.Gen.Put("for {").Push()
		s.gramCode(node.Child(), itemVar, "")
		s.Gen.Put("if %s == nil {", itemVar).Push()
		s.Gen.Put("break")
		s.Gen.Pop().Put("}")
		s.Gen.Put("%s = append(%s, %s)", tmpVar, tmpVar, itemVar)
		s.Gen.Pop().Put("}")
		s.Gen.Put("%s = NewNodesNode(%s)", itemName, tmpVar)
		s.Gen.Put("_ = %s", itemName)
	case models.GrammarRuleNodeTypeSeparatedRepeat0Item:
		if itemName == "" {
			itemName = s.Gen.CreateVar("_")
			s.Gen.Put("var %s Node", itemName)
		}
		tmpVar := s.Gen.CreateVar("_")
		s.Gen.Put("%s := make([]Node, 0)", tmpVar)
		itemVar := s.Gen.CreateVar("_")
		sepVar := s.Gen.CreateVar("_")
		s.Gen.Put("var %s Node", itemVar)
		s.Gen.Put("var %s Node", sepVar)
		s.gramCode(node.Child(), itemVar, "")
		s.Gen.Put("if %s != nil {", itemVar).Push()
		s.Gen.Put("%s = append(%s, %s)", tmpVar, tmpVar, itemVar)
		s.Gen.Put("for {").Push()
		posVar := s.Gen.CreateVar("p")
		s.Gen.Put("%s := ps._mark()", posVar)
		s.gramCode(node.Separator(), sepVar, "")
		s.Gen.Put("if %s == nil {", sepVar).Push()
		s.Gen.Put("break")
		s.Gen.Pop().Put("}")
		s.gramCode(node.Child(), itemVar, "")
		s.Gen.Put("if %s == nil {", itemVar).Push()
		s.Gen.Put("ps._reset(%s)", posVar)
		s.Gen.Put("break")
		s.Gen.Pop().Put("}")
		s.Gen.Put("%s = append(%s, %s)", tmpVar, tmpVar, itemVar)
		s.Gen.Pop().Put("}")
		s.Gen.Pop().Put("}")
		s.Gen.Put("%s = NewNodesNode(%s)", itemName, tmpVar)
		s.Gen.Put("_ = %s", itemName)
	case models.GrammarRuleNodeTypeSeparatedRepeat1Item:
		if itemName == "" {
			itemName = s.Gen.CreateVar("_")
			s.Gen.Put("var %s Node", itemName)
		}
		tmpVar := s.Gen.CreateVar("_")
		s.Gen.Put("%s := make([]Node, 0)", tmpVar)
		itemVar := s.Gen.CreateVar("_")
		sepVar := s.Gen.CreateVar("_")
		s.Gen.Put("var %s, %s Node", itemVar, sepVar)
		s.gramCode(node.Child(), itemVar, "")
		s.Gen.Put("if %s == nil {", itemVar).Push()
		s.Gen.Put("break")
		s.Gen.Pop().Put("}")
		s.Gen.Put("%s = append(%s, %s)", tmpVar, tmpVar, itemVar)
		s.Gen.Put("for {").Push()
		posVar := s.Gen.CreateVar("p")
		s.Gen.Put("%s := ps._mark()", posVar)
		s.gramCode(node.Separator(), sepVar, "")
		s.Gen.Put("if %s == nil {", sepVar).Push()
		s.Gen.Put("break")
		s.Gen.Pop().Put("}")
		s.gramCode(node.Child(), itemVar, "")
		s.Gen.Put("if %s == nil {", itemVar).Push()
		s.Gen.Put("ps._reset(%s)", posVar)
		s.Gen.Put("break")
		s.Gen.Pop().Put("}")
		s.Gen.Put("%s = append(%s, %s)", tmpVar, tmpVar, itemVar)
		s.Gen.Pop().Put("}")
		s.Gen.Put("%s = NewNodesNode(%s)", itemName, tmpVar)
		s.Gen.Put("_ = %s", itemName)
	case models.GrammarRuleNodeTypePositiveLookaheadItem, models.GrammarRuleNodeTypeNegativeLookaheadItem:
		if itemName == "" {
			itemName = s.Gen.CreateVar("_")
			s.Gen.Put("var %s Node", itemName)
		}
		posVar := s.Gen.CreateVar("p")
		s.Gen.Put("%s := ps._mark()", posVar)
		s.gramCode(node.Child(), itemName, "")
		s.Gen.Put("if %s != nil {", itemName).Push()
		s.Gen.Put("ps._reset(%s)", posVar)
		s.Gen.Pop().Put("}")
		if node.Kind() == models.GrammarRuleNodeTypeNegativeLookaheadItem {
			s.Gen.Put("if %s != nil {", itemName).Push()
		} else {
			s.Gen.Put("if %s == nil {", itemName).Push()
		}
		s.Gen.Put("break")
		s.Gen.Pop().Put("}")
	case models.GrammarRuleNodeTypeForwardIfNotMatchItem:
		if itemName == "" {
			itemName = s.Gen.CreateVar("_")
			s.Gen.Put("var %s Node", itemName)
		}
		posVar := s.Gen.CreateVar("p")
		s.Gen.Put("%s := ps._mark()", posVar)
		s.gramCode(node.Child(), itemName, "")
		s.Gen.Put("if %s != nil {", itemName).Push()
		s.Gen.Put("ps._reset(%s)", posVar)
		s.Gen.Pop().Put("}")
		s.Gen.Put("if %s == nil {", itemName).Push()
		s.Gen.Put("%s = ps._anyToken()", itemName)
		s.Gen.Pop().Put("} else {").Push()
		s.Gen.Put("break")
		s.Gen.Pop().Put("}")
	case models.GrammarRuleNodeTypeAtomItem:
		if itemName == "" {
			itemName = s.Gen.CreateVar("_")
			s.Gen.Put("var %s Node", itemName)
		}
		s.gramCode(node.Child(), itemName, "")
		s.Gen.Put("if %s == nil {", itemName).Push().Put("break").Pop().Put("}")
	case models.GrammarRuleNodeTypeNameAtom:
		s.Gen.Put("%s = ps.%s()", itemName, util.SafeName(util.ToCamelCase(node.Name())))
	case models.GrammarRuleNodeTypeTokenAtom:
		val := strings.ToLower(node.Snippet().Text())
		val = util.ToPascalCase(val)
		s.Gen.Put("%s = ps._expectK(TokenType%s)", itemName, val)
	case models.GrammarRuleNodeTypeStringAtom:
		val := node.Snippet().Text()
		val = val[1 : len(val)-1]
		if name := s.Input.Language.OperatorMap()[val]; name != "" {
			s.Gen.Put("%s = ps._expectK(TokenTypeOp%s)", itemName, util.ToPascalCase(name))
		} else if _, ok := s.Input.Language.KeywordMap()[val]; ok {
			s.Gen.Put("%s = ps._expectK(TokenTypeKw%s)", itemName, util.ToPascalCase(val))
		} else {
			val = util.DoubleQuoteStringEscape(val)
			s.Gen.Put("%s = ps._expectV(\"%s\")", itemName, val)
		}
	case models.GrammarRuleNodeTypeGroupAtom:
		inputItemName := itemName
		okVar := s.Gen.CreateVar("ok")
		posVar := s.Gen.CreateVar("p")
		s.Gen.Put("for {").Push()
		s.Gen.Put("%s := false", okVar)
		s.Gen.Put("%s := ps._mark()", posVar)
		s.Gen.Put("for {").Push()
		names := make([]string, 0)
		for i, item := range node.Children() {
			if i == len(node.Children())-1 {
				if item.Name() != "" {
					s.gramCode(item, item.Name(), "")
					s.Gen.Put("%s = %s", inputItemName, item.Name())
				} else {
					s.gramCode(item, inputItemName, "")
				}
				break
			}
			itemName = item.Name()
			if itemName == "" {
				itemName = s.Gen.CreateVar("_")
				s.Gen.Put("var %s Node", itemName)
			} else {
				names = append(names, itemName)
			}
			s.gramCode(item, itemName, "")
		}
		s.Gen.Put("%s = true", okVar)
		s.Gen.Put("break")
		s.Gen.Pop().Put("}")
		s.Gen.Put("if !%s {", okVar).Push()
		s.Gen.Put("ps._reset(%s)", posVar)
		for _, name := range names {
			s.Gen.Put("%s = nil", name)
		}
		s.Gen.Pop().Put("}")
		s.Gen.Put("break")
		s.Gen.Pop().Put("}")
	case models.GrammarRuleNodeTypeBracketEllipsisAtom:
		firstVar := s.Gen.CreateVar("first")
		lastVar := s.Gen.CreateVar("last")
		text := node.Snippet().Text()
		leftBracket := text[:2]
		rightBracket := text[len(text)-2:]
		depthVar := s.Gen.CreateVar("depth")
		s.Gen.Put("for {").Push()
		s.Gen.Put("var %s, %s Node", firstVar, lastVar)
		s.Gen.Put("if %s = ps._expectV(\"%s\"); %s == nil {", firstVar, leftBracket, firstVar).Push()
		s.Gen.Put("break")
		s.Gen.Pop().Put("}")
		s.Gen.Put("%s := 1", depthVar)
		s.Gen.Put("for {").Push()
		s.Gen.Put("if ps._expectV(\"%s\") != nil {", leftBracket).Push()
		s.Gen.Put("%s++", depthVar)
		s.Gen.Pop().Put("} else if %s = ps._expectV(\"%s\"); %s != nil {", lastVar, rightBracket, lastVar).Push()
		s.Gen.Put("%s--", depthVar)
		s.Gen.Put("if %s == 0 {", depthVar).Push()
		s.Gen.Put("break")
		s.Gen.Pop().Put("}")
		s.Gen.Pop().Put("} else if ps._expectK(TokenTypeEndOfFile) != nil {").Push()
		s.Gen.Put("panic(\"bracket ellipsis reach end of file\")")
		s.Gen.Pop().Put("} else {").Push()
		s.Gen.Put("ps._anyToken()")
		s.Gen.Pop().Put("}")
		s.Gen.Pop().Put("}")
		s.Gen.Put("%s = ps._pseudoToken(%s, %s)", itemName, firstVar, lastVar)
		s.Gen.Put("break")
		s.Gen.Pop().Put("}")
	default:
		panic("this should never happen")
	}
}

func (s *Stage32) gramActionCode(action *models.GrammarRuleNode, leftVar string) string {
	position := "ps._tokens[pos].Start, ps._visibleTokenBefore(ps._mark()).End"
	if leftVar != "" {
		position = fmt.Sprintf("%s.RangeStart(), ps._visibleTokenBefore(ps._mark()).End", leftVar)
	}
	switch action.Kind() {
	case models.GrammarRuleNodeTypeCallAction:
		args := make([]string, 0)
		for _, arg := range action.Children() {
			args = append(args, s.gramActionCode(arg, leftVar))
		}
		argsText := strings.Join(args, ", ")
		calleeName := action.Name()
		if strings.HasPrefix(calleeName, "_") {
			return fmt.Sprintf("ps.%s(%s)", util.ToCamelCase(calleeName), argsText)
		}
		if argsText != "" {
			position = fmt.Sprintf(", %s", position)
		}
		return fmt.Sprintf("New%sNode(ps._filePath, ps._fileContent, %s%s)", util.ToPascalCase(calleeName), argsText, position)
	case models.GrammarRuleNodeTypeListAction:
		elem := s.gramActionCode(action.Child(), leftVar)
		return fmt.Sprintf("NewNodesNode([]Node{%s})", elem)
	case models.GrammarRuleNodeTypeNullAction:
		return "nil"
	default:
		return action.Snippet().Text()
	}
}

func (s *Stage32) genMemoIdConsts() models.Generator {
	memoIds := make(map[int]string)
	memos := make([]int, 0)
	for rule, memoId := range s.Input.Language.MemoIdMap() {
		memos = append(memos, memoId)
		memoIds[memoId] = fmt.Sprintf("const %sMemoId = %d", util.SafeName(util.ToCamelCase(rule.Name())), memoId)
	}
	sort.Ints(memos)
	for _, memoId := range memos {
		s.Gen.Put(memoIds[memoId])
	}
	return s.Gen
}
