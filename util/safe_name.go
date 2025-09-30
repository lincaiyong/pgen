package util

import (
	"fmt"
)

var goReservedNames map[string]struct{}

func SafeName(name string) string {
	if goReservedNames == nil {
		goReservedNames = make(map[string]struct{})
		for _, n := range []string{"break", "case", "chan", "const", "continue", "default", "defer", "else", "false",
			"fallthrough", "for", "func", "go", "goto", "if", "import", "int", "interface", "map", "nil", "package", "range",
			"return", "select", "string", "struct", "switch", "true", "type", "var",
			"max", "min", "len",
		} {
			goReservedNames[n] = struct{}{}
		}
	}
	if _, ok := goReservedNames[name]; ok {
		return fmt.Sprintf("%s_", name)
	}
	return name
}
