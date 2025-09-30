package langgen

import (
	"fmt"
	"strings"
)

type VariableManager struct {
	table    map[string]struct{}
	reserved map[string]struct{}
}

func NewVariableManager(reserved map[string]struct{}) *VariableManager {
	v := &VariableManager{
		reserved: make(map[string]struct{}),
	}
	for r := range reserved {
		v.reserved[r] = struct{}{}
	}
	v.clearVar()
	return v
}

func (v *VariableManager) clearVar() {
	v.table = make(map[string]struct{})
	for n := range v.reserved {
		v.table[n] = struct{}{}
	}
}

func (v *VariableManager) createVar(name string) string {
	if !strings.HasPrefix(name, "_") {
		name = fmt.Sprintf("_%s", name)
	}
	newName := name
	index := 1
	for {
		if _, ok := v.table[newName]; ok {
			newName = fmt.Sprintf("%s%d", name, index)
			index++
		} else {
			break
		}
	}
	v.table[newName] = struct{}{}
	return newName
}
