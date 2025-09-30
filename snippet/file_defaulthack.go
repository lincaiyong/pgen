package snippet

const DefaultHackFile = `import (
	"errors"
	"fmt"
)

func (tk *Tokenizer) Parse() (tokens []*Token, err error) {
	defer func() {
		err2 := recover()
		if err2 != nil {
			err = errors.New(fmt.Sprintf("%v", err2))
		}
	}()

	tokens = make([]*Token, 0)
	for {
		var tok *Token
		tok, err = tk.next()
		if err != nil {
			return nil, err
		}

		// filter out
		if tok.Kind == TokenTypeNewline || tok.Kind == TokenTypeWhitespace {
			continue
		}

		// accept
		tokens = append(tokens, tok)
		if tok.Kind == TokenTypeEndOfFile {
			break
		}
	}
	return tokens, nil
}`
