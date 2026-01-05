package gomodparser

import (
	"fmt"
	"github.com/lincaiyong/codeedge/parser/goparser"
	"testing"
)

func TestGoMod(t *testing.T) {
	ret, err := Parse("go.mod", []byte(`module github.com/lincaiyong/codeedge

go 1.25.0

require (
	    github.com/lincaiyong/goparser v1.0.1
	    github.com/lincaiyong/log v1.0.2
)

require (
	golang.org/x/mod v0.29.0
	golang.org/x/text v0.30.0 // indirect
)
`))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(goparser.SimpleDumpNode(ret))
	ret.Visit(func(node goparser.Node) (bool, bool) {
		fmt.Printf("【%s】", string(node.Code()))
		return true, false
	}, func(node goparser.Node) bool {
		return false
	})
}
