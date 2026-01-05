package gomodparser

import (
	"fmt"
	"github.com/lincaiyong/codeedge/parser/goparser"
	"golang.org/x/mod/modfile"
	"strings"
)

type GoModNode struct {
	*goparser.BaseNode
	module   goparser.Node
	version  goparser.Node
	requires []goparser.Node
}

func (n *GoModNode) Module() goparser.Node {
	return n.module
}

func (n *GoModNode) Version() goparser.Node {
	return n.version
}

func (n *GoModNode) Requires() []goparser.Node {
	return n.requires
}

func (n *GoModNode) Visit(beforeChildren func(goparser.Node) (bool, bool), afterChildren func(goparser.Node) bool) bool {
	vc, e := beforeChildren(n)
	if e {
		return true
	}
	if !vc {
		return false
	}
	e = n.module.Visit(beforeChildren, afterChildren)
	if e {
		return true
	}
	e = n.version.Visit(beforeChildren, afterChildren)
	if e {
		return true
	}
	for _, v := range n.requires {
		e = v.Visit(beforeChildren, afterChildren)
		if e {
			return true
		}
	}
	if afterChildren(n) {
		return true
	}
	return false
}

func (n *GoModNode) Dump(hook func(goparser.Node, map[string]string) string) map[string]string {
	items := make([]string, 0)
	for _, t := range n.requires {
		items = append(items, goparser.CustomDumpNode(t, hook))
	}
	return map[string]string{
		"kind":     "gomod",
		"module":   goparser.CustomDumpNode(n.module, hook),
		"version":  goparser.CustomDumpNode(n.version, hook),
		"requires": fmt.Sprintf("[%s]", strings.Join(items, ", ")),
	}
}

func toPosition(pos modfile.Position) goparser.Position {
	return goparser.Position{
		Offset:  pos.Byte,
		LineIdx: pos.Line - 1,
		CharIdx: pos.LineRune - 1,
	}
}

func Parse(filePath string, content []byte) (*GoModNode, error) {
	if mod, parseErr := modfile.Parse(filePath, content, nil); parseErr == nil {
		rs := []rune(string(content))
		toTokenNodeByEnd := func(val string, end modfile.Position) goparser.Node {
			start := end
			start.LineRune -= len(val)
			start.Byte -= len(val)
			tok := goparser.NewToken(goparser.TokenTypeIdent, toPosition(start), toPosition(end), []rune(val))
			return goparser.NewTokenNode(filePath, rs, tok)
		}
		toTokenNodeByStart := func(val string, start, end modfile.Position) goparser.Node {
			s := string(content[start.Byte:end.Byte])
			offset := strings.Index(s, val)
			start.LineRune += offset
			start.Byte += offset
			end.LineRune = start.LineRune + len(val)
			end.Byte = start.Byte + len(val)
			tok := goparser.NewToken(goparser.TokenTypeIdent, toPosition(start), toPosition(end), []rune(val))
			return goparser.NewTokenNode(filePath, rs, tok)
		}
		module, version := goparser.DummyNode, goparser.DummyNode
		if mod.Module != nil {
			module = toTokenNodeByEnd(mod.Module.Mod.Path, mod.Module.Syntax.End)
		}
		if mod.Go != nil {
			version = toTokenNodeByEnd(mod.Go.Version, mod.Go.Syntax.End)
		}
		requires := make([]goparser.Node, 0)
		for _, item := range mod.Require {
			if !item.Indirect {
				m := toTokenNodeByStart(item.Mod.Path, item.Syntax.Start, item.Syntax.End)
				v := toTokenNodeByEnd(item.Mod.Version, item.Syntax.End)
				require := goparser.NewNodesNode([]goparser.Node{m, v})
				requires = append(requires, require)
			}
		}
		_, lastPos_ := mod.Syntax.Span()
		lastPos := toPosition(lastPos_)
		node := &GoModNode{
			BaseNode: goparser.NewBaseNode(filePath, rs, "gomod", module.RangeStart(), lastPos),
			module:   module,
			version:  version,
			requires: requires,
		}
		return node, nil
	} else {
		return nil, parseErr
	}
}
