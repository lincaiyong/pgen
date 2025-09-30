package snippet

const ParseFunc = `func ParseFile(filePath string) (Node, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	r, _ := DecodeBytes(b)
	tokenizer := NewTokenizer(filePath, r)
	var tokens []*Token
	tokens, err = tokenizer.Parse()
	if err != nil {
		return nil, err
	}
	tokens = tokenizer.Clean(tokens)
	parser := NewParser(filePath, r, tokens)
	var ret Node
	ret, err = parser.Parse()
	if err != nil {
		return nil, err
	}
	if ret != nil {
		ret.BuildLink()
	}
	return ret, nil
}

func ParseBytes(filePath string, b []byte) (Node, error) {
	var err error
	r, _ := DecodeBytes(b)
	tokenizer := NewTokenizer(filePath, r)
	var tokens []*Token
	tokens, err = tokenizer.Parse()
	if err != nil {
		return nil, err
	}
	tokens = tokenizer.Clean(tokens)
	parser := NewParser(filePath, r, tokens)
	var ret Node
	ret, err = parser.Parse()
	if err != nil {
		return nil, err
	}
	if ret != nil {
		ret.BuildLink()
	}
	return ret, nil
}`
