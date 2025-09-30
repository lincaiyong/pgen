package models

type Generator interface {
	String() string
	Put(string, ...any) Generator
	PutNL() Generator
	Push() Generator
	Pop() Generator
	ClearVar()
	CreateVar(string) string
}
