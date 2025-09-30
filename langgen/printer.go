package langgen

import (
	"fmt"
	"strings"
)

type Printer struct {
	sb     *strings.Builder
	indent string
}

func NewPrinter() *Printer {
	return &Printer{
		sb:     &strings.Builder{},
		indent: "",
	}
}

func (pr *Printer) push() *Printer {
	pr.indent += "\t"
	return pr
}

func (pr *Printer) pop() *Printer {
	if len(pr.indent) > 0 {
		pr.indent = pr.indent[:len(pr.indent)-1]
	}
	return pr
}

func (pr *Printer) putNL() *Printer {
	pr.sb.WriteString("\n")
	return pr
}

func (pr *Printer) put(t string, args ...interface{}) *Printer {
	if len(args) > 0 {
		t = fmt.Sprintf(t, args...)
	}

	if t != "" {
		pr.sb.WriteString(pr.indent + t)
	}
	pr.sb.WriteString("\n")
	return pr
}
