package snippet

const ParserStruct = `type Parser struct {
	_filePath    string
	_fileContent []rune

	_tokens []*Token
	_max    int
	_pos    int
	_x      int

	_bracketDepth  int
	_bracketDepths []int

	_nodeCache []map[int]*NodeCache

	_any any
}

func NewParser(filePath string, fileContent []rune, tokens []*Token) *Parser {
	ps := Parser{_filePath: filePath, _fileContent: fileContent, _tokens: tokens}
	ps._max = len(ps._tokens)
	ps._pos = 0
	ps._x = 0

	ps._bracketDepths = make([]int, ps._max+1)
	ps._nodeCache = make([]map[int]*NodeCache, ps._max)

	return &ps
}

func (ps *Parser) _mark() int {
	ps._bracketDepths[ps._pos] = ps._bracketDepth
	return ps._pos
}

func (ps *Parser) _reset(pos int) {
	ps._pos = pos
	ps._bracketDepth = ps._bracketDepths[ps._pos]
}

func (ps *Parser) _stepForward(tok *Token) {
	if len(tok.Value) == 1 {
		val := tok.Value[0]
		if val == '(' || val == '[' || val == '{' {
			ps._bracketDepth++
		} else if val == ')' || val == ']' || val == '}' {
			ps._bracketDepth--
		}
	}
	ps._pos++
	if ps._pos >= ps._max {
		ps._pos = ps._max - 1
	}
	if ps._pos > ps._x {
		ps._x = ps._pos
	}
}

func (ps *Parser) _expectK(kind string) Node {
	tok := ps._tokens[ps._pos]
	if tok.Kind == kind {
		ps._stepForward(tok)
		return NewTokenNode(ps._filePath, ps._fileContent, tok)
	}
	return nil
}

func (ps *Parser) _expectV(val string) Node {
	tok := ps._tokens[ps._pos]
	if len(tok.Value) == len(val) && string(tok.Value) == val {
		ps._stepForward(tok)
		return NewTokenNode(ps._filePath, ps._fileContent, tok)
	}
	return nil
}

func (ps *Parser) _anyToken() Node {
	tok := ps._tokens[ps._pos]
	ps._stepForward(tok)
	return NewTokenNode(ps._filePath, ps._fileContent, tok)
}

func (ps *Parser) _pseudoToken(v ...any) Node {
	var start, end *Position
	for _, t := range v {
		switch vv := t.(type) {
		case *Token:
			if vv == nil {
				continue
			}
			if start == nil {
				start = &vv.Start
			}
			end = &vv.End
		case []*Token:
			if vv == nil {
				continue
			}
			if len(vv) > 0 {
				if start == nil {
					start = &vv[0].Start
				}
				end = &vv[len(vv)-1].End
			}
		case Node:
			if vv == nil {
				continue
			}
			if start == nil {
				p := vv.RangeStart()
				start = &p
			}
			p := vv.RangeEnd()
			end = &p
		case []Node:
			if vv == nil {
				continue
			}
			if len(vv) > 0 {
				if start == nil {
					p := vv[0].RangeStart()
					start = &p
				}
				p := vv[len(vv)-1].RangeEnd()
				end = &p
			}
		default:
			return nil
		}
	}
	if start == nil || end == nil {
		return nil
	}
	val := ps._fileContent[start.Offset:end.Offset]
	return NewTokenNode(ps._filePath, ps._fileContent, NewToken(TokenTypePseudo, *start, *end, val))
}

func (ps *Parser) _expectPseudoNewline() Node {
	if ps._pos < 1 || ps._pos >= len(ps._tokens) {
		return nil
	}
	current := ps._tokens[ps._pos-1]
	lookahead := ps._tokens[ps._pos]
	if current.End.LineIdx == lookahead.Start.LineIdx {
		return nil
	}
	return NewTokenNode(ps._filePath, ps._fileContent, lookahead)
}

func (ps *Parser) _visibleTokenBefore(pos int) *Token {
	for i := pos - 1; i >= 0; i-- {
		kind := ps._tokens[i].Kind
		if kind != TokenTypeWhitespace && kind != TokenTypeNewline {
			return ps._tokens[i]
		}
	}
	return nil
}

func (ps *Parser) _mergeNodes(items ...any) Node {
	ret := make([]Node, 0)
	for _, item := range items {
		if item == nil {
			continue
		}
		if n, ok := item.(Node); ok {
			if n != nil && !n.IsDummy() {
				ret = append(ret, n)
			}
		} else if s, ok := item.([]Node); ok {
			if s != nil {
				for _, ss := range s {
					if ss != nil && !ss.IsDummy() {
						ret = append(ret, ss)
					}
				}
			}
		} else {
			panic("misused merge_nodes api")
		}
	}
	return NewNodesNode(ret)
}

func (ps *Parser) Parse() (ret Node, err error) {
	ret = ps.file()
	if ps._expectK(TokenTypeEndOfFile) != nil {
		return ret, nil
	}
	tok := ps._tokens[ps._x]
	return nil, fmt.Errorf("fail to parse: %s\n%s", ps._filePath, errorContext(ps._filePath, ps._fileContent, tok.Start.Offset, tok.Start.LineIdx, tok.Start.CharIdx))
}`
