package stages

import (
	"fmt"
	"github.com/lincaiyong/pgen/langgen"
	"github.com/lincaiyong/pgen/models"
	"github.com/lincaiyong/pgen/util"
	"strings"
)

func RunStage33(s2 *Stage2) *Stage33 {
	stage3 := &Stage33{
		Description: "generate other code",
		Input:       s2,
		Gen:         langgen.NewGenerator(),
		Error:       models.NewError(),
	}
	stage3.run()
	return stage3
}

type Stage33 struct {
	Description string
	Input       *Stage2
	Gen         models.Generator
	Error       *models.Error
}

func (s *Stage33) run() {
	s.nodeInterfaceAndStructs()
}

func (s *Stage33) nodeInterfaceAndStructs() {
	for _, node := range s.Input.Language.AstNodes() {
		s.Gen.Put("type I%sNode interface {", util.ToPascalCase(node.Name())).Push()
		{
			s.Gen.Put("Node")
			for _, arg := range node.Args() {
				s.Gen.Put("%s() Node", arg.Pascal())
				s.Gen.Put("Set%s(v Node)", arg.Pascal())
			}
		}
		s.Gen.Pop().Put("}").PutNL()
	}

	for _, node := range s.Input.Language.AstNodes() {
		pascalName := util.ToPascalCase(node.Name())
		params := make([]string, 0)
		for _, arg := range node.Args() {
			param := fmt.Sprintf("%s Node, ", arg.Camel())
			params = append(params, param)
		}
		s.Gen.Put("func New%sNode(filePath string, fileContent []rune, %sstart, end Position) Node {", pascalName, strings.Join(params, "")).Push()
		maxLen := 0
		for _, arg := range node.Args() {
			s.Gen.Put("if %s == nil {", arg.Camel()).Push()
			s.Gen.Put("%s = dummyNode", arg.Camel())
			s.Gen.Pop().Put("}")
			if maxLen < len(arg.Camel()) {
				maxLen = len(arg.Camel())
			}
		}
		maxLen8 := maxLen
		if maxLen8 < 8 {
			maxLen8 = 8
		}
		s.Gen.Put("return &%sNode{", pascalName).Push()
		s.Gen.Put("BaseNode:%s NewBaseNode(filePath, fileContent, NodeType%s, start, end),",
			util.MakePadding(maxLen8-8, ' '), pascalName)
		for _, arg := range node.Args() {
			name := arg.Camel()
			s.Gen.Put("%s:%s %s,", name, util.MakePadding(maxLen8-len(name), ' '), name)
		}
		s.Gen.Pop().Put("}")
		s.Gen.Pop().Put("}").PutNL()

		s.Gen.Put("type %sNode struct {", pascalName).Push()
		s.Gen.Put("*BaseNode")
		for _, arg := range node.Args() {
			name := arg.Camel()
			s.Gen.Put("%s%s Node", name, util.MakePadding(maxLen-len(name), ' '))
		}
		s.Gen.Pop().Put("}").PutNL()

		for _, arg := range node.Args() {
			s.Gen.Put("func (n *%sNode) %s() Node {", pascalName, arg.Pascal()).Push()
			s.Gen.Put("return n.%s", arg.Camel())
			s.Gen.Pop().Put("}").PutNL()
			s.Gen.Put("func (n *%sNode) Set%s(v Node) {", pascalName, arg.Pascal()).Push()
			s.Gen.Put("n.%s = v", arg.Camel())
			s.Gen.Pop().Put("}").PutNL()
		}

		s.Gen.Put("func (n *%sNode) BuildLink() {", pascalName).Push()
		for _, arg := range node.Args() {
			s.Gen.Put("if !n.%s().IsDummy() {", arg.Pascal()).Push()
			s.Gen.Put("%s := n.%s()", arg.Camel(), arg.Pascal())
			s.Gen.Put("%s.BuildLink()", arg.Camel())
			s.Gen.Put("%s.SetParent(n)", arg.Camel())
			s.Gen.Put("%s.SetSelfField(\"%s\")", arg.Camel(), arg.Normal())
			s.Gen.Put("%s.SetReplaceSelf(func(n Node) {", arg.Camel()).Push()
			s.Gen.Put("n.Parent().(I%sNode).Set%s(n)", pascalName, arg.Pascal())
			s.Gen.Pop().Put("})")
			s.Gen.Pop().Put("}")
		}
		s.Gen.Pop().Put("}").PutNL()

		if len(node.Args()) > 0 {
			s.Gen.Put("func (n *%sNode) Fields() []string {", pascalName).Push()
			s.Gen.Put("return []string{").Push()
			for _, arg := range node.Args() {
				s.Gen.Put("\"%s\",", arg.Normal())
			}
			s.Gen.Pop().Put("}")
			s.Gen.Pop().Put("}").PutNL()
		}

		s.Gen.Put("func (n *%sNode) Child(field string) Node {", pascalName).Push()
		s.Gen.Put("if field == \"\" {").Push()
		s.Gen.Put("return nil")
		s.Gen.Pop().Put("}")
		for _, arg := range node.Args() {
			s.Gen.Put("if field == \"%s\" {", arg.Normal()).Push()
			s.Gen.Put("return n.%s()", arg.Pascal())
			s.Gen.Pop().Put("}")
		}
		s.Gen.Put("return nil")
		s.Gen.Pop().Put("}").PutNL()

		s.Gen.Put("func (n *%sNode) SetChild(nodes []Node) {", pascalName).Push()
		s.Gen.Put("if len(nodes) != %d {", len(node.Args())).Push()
		s.Gen.Put("return")
		s.Gen.Pop().Put("}")
		for i, arg := range node.Args() {
			s.Gen.Put("n.Set%s(nodes[%d])", util.ToPascalCase(arg.Normal()), i)
		}
		s.Gen.Pop().Put("}").PutNL()

		s.Gen.Put("func (n *%sNode) Fork() Node {", pascalName).Push()
		s.Gen.Put("_ret := &%sNode{", pascalName).Push()
		s.Gen.Put("BaseNode:%s n.BaseNode.fork(),", util.MakePadding(maxLen8-8, ' '))
		for _, arg := range node.Args() {
			s.Gen.Put("%s:%s n.%s.Fork(),", arg.Camel(), util.MakePadding(maxLen8-len(arg.Camel()), ' '), arg.Camel())
		}
		s.Gen.Pop().Put("}")
		for _, arg := range node.Args() {
			s.Gen.Put("_ret.%s.SetParent(_ret)", arg.Camel())
		}
		s.Gen.Put("return _ret")
		s.Gen.Pop().Put("}").PutNL()

		s.Gen.Put("func (n *%sNode) Visit(beforeChildren func(node Node) (visitChildren, exit bool), afterChildren func(node Node) (exit bool)) (exit bool) {", pascalName).Push()
		s.Gen.Put("vc, e := beforeChildren(n)")
		s.Gen.Put("if e {").Push()
		s.Gen.Put("return true")
		s.Gen.Pop().Put("}")
		s.Gen.Put("if !vc {").Push()
		s.Gen.Put("return false")
		s.Gen.Pop().Put("}")
		for _, arg := range node.Args() {
			s.Gen.Put("if n.%s.Visit(beforeChildren, afterChildren) {", arg.Camel()).Push()
			s.Gen.Put("return true")
			s.Gen.Pop().Put("}")
		}
		s.Gen.Put("if afterChildren(n) {").Push()
		s.Gen.Put("return true")
		s.Gen.Pop().Put("}")
		s.Gen.Put("return false")
		s.Gen.Pop().Put("}").PutNL()

		dumpFunHead := "func (n *%sNode) Dump(hook func(Node, map[string]string) string) map[string]string {"
		if len(node.Args()) == 0 {
			dumpFunHead = strings.ReplaceAll(dumpFunHead, "hook", "_")
		}
		s.Gen.Put(dumpFunHead, pascalName).Push()
		s.Gen.Put(`ret := make(map[string]string)`)
		s.Gen.Put(`ret["kind"] = "\"%s\""`, node.Name())

		for _, arg := range node.Args() {
			s.Gen.Put(`ret["%s"] = dumpNode(n.%s(), hook)`, arg.Normal(), arg.Pascal())
		}
		s.Gen.Put("return ret")
		s.Gen.Pop().Put("}").PutNL()
	}
}
