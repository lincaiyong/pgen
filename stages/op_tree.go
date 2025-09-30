package stages

import (
	"github.com/lincaiyong/pgen/models"
	"github.com/lincaiyong/pgen/util"
	"sort"
)

type OperatorNode struct {
	ch          int
	name        string
	level       int
	childrenMap map[int]*OperatorNode
	children    []int
}

func NewOperatorNode(ch int, level int) *OperatorNode {
	node := &OperatorNode{
		ch:          ch,
		name:        "",
		level:       level,
		childrenMap: make(map[int]*OperatorNode),
		children:    make([]int, 0),
	}
	return node
}

func (n *OperatorNode) Update(name string, val []byte) {
	if n.level == len(val) {
		n.name = name
		return
	}
	ch := int(val[n.level])
	if _, ok := n.childrenMap[ch]; ok {
		n.childrenMap[ch].Update(name, val)
	} else {
		sub := NewOperatorNode(ch, n.level+1)
		n.childrenMap[ch] = sub
		n.children = append(n.children, ch)
		sub.Update(name, val)
	}
}

func (n *OperatorNode) unescapeCh() string {
	if n.ch == '\\' {
		return "\\\\"
	}
	return string(byte(n.ch))
}

func (n *OperatorNode) GenCode(gen models.Generator) {
	sort.Ints(n.children)
	for _, c := range n.children {
		child := n.childrenMap[c]
		gen.Put("case '%s':", child.unescapeCh()).Push()
		gen.Put("entered = true")
		gen.Put("tk._forward()")
		child.genChildCode(gen)
		if child.name != "" {
			gen.Put("kind = TokenTypeOp%s", util.ToPascalCase(child.name))
		}
		gen.Pop()
	}
}

func (n *OperatorNode) genChildCode(gen models.Generator) {
	sort.Ints(n.children)
	for _, c := range n.children {
		child := n.childrenMap[c]
		var posVar string
		gen.Put("if tk._lookahead == '%s' {", child.unescapeCh()).Push()
		if child.name == "" {
			posVar = gen.CreateVar("p")
			gen.Put("%s := tk._mark()", posVar)
		}
		gen.Put("tk._forward()")
		child.genChildCode(gen)
		if child.name != "" {
			gen.Put("kind = TokenTypeOp%s", util.ToPascalCase(child.name))
			gen.Put("break")
		} else {
			gen.Put("tk._reset(%s)", posVar)
		}

		gen.Pop().Put("}")
	}
}
