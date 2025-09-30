package models

import "github.com/lincaiyong/pgen/util"

func NewName(val string) *Name {
	return &Name{
		normal: util.SafeName(val),
		camel:  util.SafeName(util.ToCamelCase(val)),
		pascal: util.ToPascalCase(val),
	}
}

type Name struct {
	normal string
	camel  string
	pascal string
}

func (n Name) Normal() string {
	return n.normal
}

func (n Name) Camel() string {
	return n.camel
}

func (n Name) Pascal() string {
	return n.pascal
}
