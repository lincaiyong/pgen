package langparse

import (
	"fmt"
	"github.com/lincaiyong/pgen/config"
	"github.com/lincaiyong/pgen/models"
	"runtime"
	"strings"
)

type BaseParser struct {
	input *models.Snippet
	pos   models.Position
	max   models.Position
	la    byte

	next []byte
}

func NewBaseParser(input *models.Snippet) *BaseParser {
	ret := &BaseParser{
		input: input,
		pos:   input.Start,
		max:   input.Start,
	}
	ret.lookahead()
	return ret
}

func (p *BaseParser) expectError(expected string) error {
	var sb strings.Builder
	content := string(p.input.FileContent)
	lines := strings.Split(content, "\n")
	startLine := max(0, p.max.LineIdx-3)
	endLine := min(len(lines), p.max.LineIdx+4)
	for i := startLine; i < endLine; i++ {
		sb.WriteString(fmt.Sprintf("%d\t%s\n", i+1, lines[i]))
	}
	sb.WriteString("----------------\n")
	pc := make([]uintptr, 100)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	for {
		frame, more := frames.Next()
		if !strings.Contains(frame.File, "codeedge/") {
			break
		}
		sb.WriteString(fmt.Sprintf("%s\n - %s:%d\n", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}
	line := string(p.input.FileContent[p.max.Offset:])
	if idx := strings.Index(line, "\n"); idx != -1 {
		line = line[:idx]
	}
	return fmt.Errorf("expect %s at %d:%d, \"%s\"\n%s", expected, p.max.LineIdx+1, p.max.CharIdx+1, line, sb.String())
}

func (p *BaseParser) mark() models.Position {
	return p.pos
}

func (p *BaseParser) reset(pos models.Position) {
	p.pos = pos
	p.lookahead()
}

func (p *BaseParser) expect(v byte) bool {
	if v == p.la {
		p.stepForward()
		return true
	}
	return false
}

func (p *BaseParser) lookahead() {
	if p.pos.Offset >= p.input.End.Offset {
		p.la = 0
	} else {
		p.la = p.input.FileContent[p.pos.Offset]
	}
}

func (p *BaseParser) reachEnd() bool {
	return p.la == 0
}

func (p *BaseParser) stepForward() {
	if p.la == 0 {
		return
	}
	p.pos.Offset++
	la := p.la
	p.lookahead()
	if la == '\n' {
		p.pos.LineIdx++
		p.pos.CharIdx = 0
	} else {
		p.pos.CharIdx++
	}
	if p.pos.Offset > p.max.Offset {
		p.max = p.pos
	}

	if config.DebugMode() {
		p.next = p.input.FileContent[p.pos.Offset:]
	}
}

func (p *BaseParser) forwardUtil(breakFn func(byte) bool) (models.Position, models.Position) {
	start := p.pos
	for !p.reachEnd() {
		if breakFn(p.la) {
			break
		}
		p.stepForward()
	}
	return start, p.pos
}

func (p *BaseParser) expectIdentifier() *models.Snippet {
	start, end := p.forwardUtil(func(b byte) bool {
		return !((b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_')
	})
	if start.Offset == end.Offset {
		return nil
	}
	return p.input.Fork(start, end)
}

func (p *BaseParser) isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func (p *BaseParser) skipWhitespace() {
	p.forwardUtil(func(b byte) bool {
		return !p.isWhitespace(b)
	})
}

func (p *BaseParser) expectString(s string) bool {
	pos := p.mark()
	for _, b := range []byte(s) {
		if !p.expect(b) {
			p.reset(pos)
			return false
		}
	}
	return true
}
