package stages

import (
	"fmt"
	"github.com/lincaiyong/pgen/config"
	"github.com/lincaiyong/pgen/langgen"
	"github.com/lincaiyong/pgen/models"
	"github.com/lincaiyong/pgen/snippet"
	"github.com/lincaiyong/pgen/util"
	"sort"
	"strings"
)

func RunStage4(stage31 *Stage31, stage32 *Stage32, stage33 *Stage33) *Stage4 {
	stage4 := &Stage4{
		Description: "generate final code",
		Input1:      stage31,
		Input2:      stage32,
		Input3:      stage33,
		Gen:         langgen.NewGenerator(),
		Error:       models.NewError(),
	}
	stage4.run()
	return stage4
}

type Stage4 struct {
	Description string
	Input1      *Stage31
	Input2      *Stage32
	Input3      *Stage33
	Gen         models.Generator
	Error       *models.Error
}

func (s *Stage4) run() {
	s.Gen.Put("package goparser").PutNL()
	s.Gen.Put(snippet.ImportCode).PutNL()
	s.Gen.Put(snippet.PositionStruct).PutNL()
	s.Gen.Put(snippet.TokenStruct).PutNL()
	s.Gen.Put(snippet.NodeInterface).PutNL()
	s.constTokenTypes().PutNL()
	s.constNodeTypes().PutNL()
	//
	s.Gen.Put(snippet.ErrorContextFunc).PutNL()
	s.Gen.Put(snippet.ToSnakeCaseFunc).PutNL()
	s.Gen.Put(snippet.ToCamelCaseFunc).PutNL()
	s.Gen.Put(snippet.DecodeBytesFunc).PutNL()
	s.Gen.Put(snippet.TypeNameOfFunc).PutNL()
	s.Gen.Put(snippet.EqualRuneFunc).PutNL()
	s.Gen.Put(snippet.InRangeFunc).PutNL()
	s.Gen.Put(snippet.NodesSetParentFunc).PutNL()
	s.Gen.Put(snippet.NodesVisitFunc).PutNL()
	s.Gen.Put(snippet.CreationHookVar).PutNL()
	s.Gen.Put(snippet.DummyNodeVar).PutNL()
	s.Gen.Put(snippet.BaseNodeStruct).PutNL()
	s.Gen.Put(snippet.NodesNodeStruct).PutNL()
	s.Gen.Put(snippet.TokenNodeStruct).PutNL()
	s.Gen.Put(s.Input3.Gen.String()).PutNL()
	s.Gen.Put(s.Input1.Gen.String()).PutNL()
	s.Gen.Put(s.Input2.Gen.String()).PutNL()
	s.Gen.Put(s.Input1.Input.Language.HackCode())
	s.Gen.Put(snippet.DumpNodeFunc).PutNL()
	s.Gen.Put(snippet.QueryNodeFunc).PutNL()
	s.Gen.Put(snippet.ParseFunc).PutNL()
}

func (s *Stage4) constNodeTypes() models.Generator {
	nodeTypes := []string{"dummy", "token", "nodes"}
	for _, node := range s.Input1.Input.Language.AstNodes() {
		nodeTypes = append(nodeTypes, node.Name())
	}
	for _, t := range nodeTypes {
		s.Gen.Put("const NodeType%s = \"%s\"", util.ToPascalCase(t), t)
	}
	return s.Gen
}

func (s *Stage4) constTokenTypes() models.Generator {
	tokens := make([]string, 0)
	for _, rule := range s.Input1.Input.Language.TokenRules() {
		if !strings.HasPrefix(rule.Name(), "_") {
			tokens = append(tokens, rule.Name())
		}
	}
	sort.Strings(tokens)
	tokens = append(config.BuiltinTokens(), tokens...)

	operators := make([]string, 0)
	m := make(map[string]string)
	for op, name := range s.Input1.Input.Language.OperatorMap() {
		opName := fmt.Sprintf("op_%s", name)
		operators = append(operators, opName)
		m[opName] = op
	}
	sort.Strings(operators)

	keywords := make([]string, 0)
	for _, name := range s.Input1.Input.Language.Keywords() {
		keywords = append(keywords, fmt.Sprintf("kw_%s", name))
	}
	sort.Strings(keywords)

	tokenTypes := []string{"dummy"}
	tokenTypes = append(tokenTypes, tokens...)
	tokenTypes = append(tokenTypes, operators...)
	tokenTypes = append(tokenTypes, keywords...)
	for _, t := range tokenTypes {
		var v string
		if v = m[t]; v == "" {
			v = t
		}
		s.Gen.Put("const TokenType%s = \"%s\"", util.ToPascalCase(t), v)
	}
	return s.Gen
}
