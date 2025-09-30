package langgen

import (
	"github.com/lincaiyong/pgen/config"
	"github.com/lincaiyong/pgen/models"
	"strings"
)

func NewGenerator() *Generator {
	return &Generator{
		Printer:         NewPrinter(),
		VariableManager: NewVariableManager(config.ReservedVariables()),
	}
}

type Generator struct {
	*Printer
	*VariableManager
}

func (gen *Generator) String() string {
	return strings.TrimRight(gen.sb.String(), "\n")
}

func (gen *Generator) Put(s string, a ...any) models.Generator {
	gen.put(s, a...)
	return gen
}

func (gen *Generator) PutNL() models.Generator {
	gen.putNL()
	return gen
}

func (gen *Generator) Push() models.Generator {
	gen.push()
	return gen
}

func (gen *Generator) Pop() models.Generator {
	gen.pop()
	return gen
}

func (gen *Generator) ClearVar() {
	gen.clearVar()
}

func (gen *Generator) CreateVar(s string) string {
	return gen.createVar(s)
}
