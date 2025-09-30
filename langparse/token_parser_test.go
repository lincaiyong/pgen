package langparse

import (
	"github.com/lincaiyong/pgen/models"
	"testing"
)

func TestTokenParser01(t *testing.T) {
	input := models.NewSnippet("", []byte(`newline:
| '\r\n'
| '\n'
| '\r'`))
	rules, err := ParseTokenRule(input)
	if err != nil {
		t.Fatal(err)
	}
	print(rules)
}

func TestTokenParser02(t *testing.T) {
	input := models.NewSnippet("", []byte(`_ident_ch:
    | !_whitespace_ch [a-zA-Z_\u0080-\uFFFF]`))
	rules, err := ParseTokenRule(input)
	if err != nil {
		t.Fatal(err)
	}
	print(rules)
}
