package models

import "strings"

type Position struct {
	Offset  int
	LineIdx int
	CharIdx int
}

func NewPosition(offset, lineIdx, charIdx int) Position {
	return Position{offset, lineIdx, charIdx}
}

func (p *Position) MoveForward(text string) Position {
	lineCount := strings.Count(text, "\n")
	var charIdx int
	if lineCount == 0 {
		charIdx = p.CharIdx + len(text)
	} else {
		charIdx = len(text) - strings.LastIndex(text, "\n") - 1
	}
	return Position{
		Offset:  p.Offset + len(text),
		LineIdx: p.LineIdx + lineCount,
		CharIdx: charIdx,
	}
}

func (p *Position) SameAs(pp Position) bool {
	return p.Offset == pp.Offset
}
