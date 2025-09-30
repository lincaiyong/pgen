package snippet

const TokenizerStruct = `func NewTokenizer(filePath string, fileContent []rune) *Tokenizer {
	tk := &Tokenizer{
		_filePath:  filePath,
		_buf:       fileContent,
		_bufSize:   len(fileContent),
		_pos:       Position{},
		_prevPos:   Position{},
		_lookahead: 0,
	}
	tk._lookahead = tk._safeRead()
	tk.initKeywords()
	return tk
}

type Tokenizer struct {
	_filePath  string
	_buf       []rune
	_bufSize   int
	_pos       Position
	_prevPos   Position
	_lookahead rune
	_keywords  map[string]string
}

func (tk *Tokenizer) Parse() (tokens []*Token, err error) {
	tokens = make([]*Token, 0)
	for {
		var tok *Token
		tok, err = tk.next()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, tok)
		if tok.Kind == TokenTypeEndOfFile {
			break
		}
	}
	return tokens, nil
}

func (tk *Tokenizer) _lineEnd(ch rune) bool {
	return ch == '\n' || (ch == '\r' && tk._pos.Offset < len(tk._buf) && tk._buf[tk._pos.Offset] != '\n')
}

func (tk *Tokenizer) _errorMsg(msg string) string {
	return fmt.Sprintf("fail to tokenize %s\n%s", msg, errorContext(tk._filePath, tk._buf, tk._prevPos.Offset, tk._prevPos.LineIdx, tk._prevPos.CharIdx))
}

func (tk *Tokenizer) _stepForward(ch rune) {
	p := &tk._pos
	p.Offset++
	p.CharIdx++
	if tk._lineEnd(ch) {
		p.LineIdx++
		p.CharIdx = 0
	}
}

func (tk *Tokenizer) _forward() {
	tk._stepForward(tk._safeRead())
	tk._lookahead = tk._safeRead()
}

func (tk *Tokenizer) _mark() Position {
	return tk._pos
}

func (tk *Tokenizer) _reset(p Position) {
	tk._pos = p
	tk._lookahead = tk._safeRead()
}

func (tk *Tokenizer) _safeRead() rune {
	if tk._pos.Offset >= tk._bufSize {
		return '\x00'
	} else {
		return tk._buf[tk._pos.Offset]
	}
}

func (tk *Tokenizer) _expect(r rune) bool {
	if equalRune(r, tk._lookahead) {
		tk._forward()
		return true
	}
	return false
}

func (tk *Tokenizer) _expectS(s string) bool {
	pos := tk._pos
	for i := 0; i < len(s); i++ {
		if equalRune(rune(s[i]), tk._lookahead) {
			tk._forward()
		} else {
			tk._reset(pos)
			return false
		}
	}
	return true
}

func (tk *Tokenizer) _expectU(s []rune) bool {
	pos := tk._pos
	for i := 0; i < len(s); i++ {
		if equalRune(s[i], tk._lookahead) {
			tk._forward()
		} else {
			tk._reset(pos)
			return false
		}
	}
	return true
}

func (tk *Tokenizer) _expectR(s, e rune) bool {
	if inRange(tk._lookahead, s, e) {
		tk._forward()
		return true
	}
	return false
}

func (tk *Tokenizer) _anyButEof() bool {
	if tk._lookahead != 0 {
		tk._forward()
		return true
	} else {
		return false
	}
}

func (tk *Tokenizer) _createToken(kind string) *Token {
	val := tk._buf[tk._prevPos.Offset:tk._pos.Offset]
	token := NewToken(kind, tk._prevPos, tk._pos, val)
	tk._prevPos = tk._pos
	return token
}

// newline:
//
//	| '\r\n'
//	| '\n'
//	| '\r'
func (tk *Tokenizer) newline() bool {
	// '\r\n'
	if tk._expectS("\r\n") {
		return true
	}
	// '\n'
	if tk._expectS("\n") {
		return true
	}
	// '\r'
	if tk._expectS("\r") {
		return true
	}
	return false
}

// _any_but_eol:
//
//	| !newline _any_but_eof
func (tk *Tokenizer) _anyButEol() bool {
	// !newline _any_but_eof
	_p := tk._mark()
	_ok := false
	if tk.newline() {
		_ok = true
	}
	tk._reset(_p)
	if !_ok {
		if tk._anyButEof() {
			return true
		}
	}
	return false
}

// _whitespace_ch:
//
//	| [ \t\f\u1680\u180E\u2000-\u200A\u202F\u205F\u3000\uFEFF\u00A0]
func (tk *Tokenizer) _whitespaceCh() bool {
	// [ \t\f\u1680\u180E\u2000-\u200A\u202F\u205F\u3000\uFEFF\u00A0]
	if tk._expect(0x20) || tk._expect(0x9) || tk._expect(0xC) || tk._expect(0x1680) || tk._expect(0x180E) || tk._expectR(0x2000, 0x200A) || tk._expect(0x202F) || tk._expect(0x205F) || tk._expect(0x3000) || tk._expect(0xFEFF) || tk._expect(0xA0) {
		return true
	}
	return false
}

// whitespace:
//
//	| _whitespace_ch+
func (tk *Tokenizer) whitespace() bool {
	// _whitespace_ch+
	if tk._whitespaceCh() {
		for {
			_ok := false
			if tk._whitespaceCh() {
				_ok = true
			}
			if !_ok {
				break
			}
		}
		return true
	}
	return false
}

func (tk *Tokenizer) op() string {
	entered := false
	kind := TokenTypeDummy
	switch tk._lookahead {<op_placeholder>
	default:
		break
	}
	if entered && kind == TokenTypeDummy {
		panic(tk._errorMsg("op"))
	} else {
		return kind
	}
}

func (tk *Tokenizer) next() (*Token, error) {
	kind := TokenTypeDummy
	if tk._lookahead == '\x00' {
		if tk._pos.Offset > tk._bufSize {
			panic(tk._errorMsg("eof"))
		}
		tk._stepForward('\x00')
		kind = TokenTypeEndOfFile
	} else if tk.whitespace() {
		kind = TokenTypeWhitespace
	} else if tk.newline() {
		kind = TokenTypeNewline<next_placeholder>
	} else {
		kind = tk.op()
		if kind == TokenTypeDummy {
			return nil, errors.New(tk._errorMsg(string(tk._buf[tk._prevPos.Offset])))
		}
	}

	var val []rune
	if kind == TokenTypeEndOfFile {
		val = []rune("END_OF_FILE")
	} else {
		val = tk._buf[tk._prevPos.Offset:tk._pos.Offset]
	}
	if kind == TokenTypeIdent {
		k, ok := tk._keywords[string(val)]
		if ok {
			kind = k
		}
	}
	ret := NewToken(kind, tk._prevPos, tk._pos, val)
	tk._prevPos = tk._pos
	return ret, nil
}`
