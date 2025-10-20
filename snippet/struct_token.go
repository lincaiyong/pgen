package snippet

const TokenStruct = `func NewToken(kind string, start, end Position, val []rune) *Token {
	return &Token{
		Kind:  kind,
		Start: start,
		End:   end,
		Value: val,
	}
}

type Token struct {
	Kind  string
	Start Position
	End   Position
	Value []rune
}

func (t *Token) Fork() *Token {
	return &Token{
		Kind:  t.Kind,
		Start: t.Start,
		End:   t.End,
		Value: t.Value,
	}
}`
