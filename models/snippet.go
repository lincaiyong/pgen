package models

import (
	"github.com/lincaiyong/pgen/config"
	"strings"
)

func NewSnippet(filePath string, fileContent []byte) *Snippet {
	text := string(fileContent)
	lineIdx := strings.Count(text, "\n")
	charIdx := len(text) - strings.LastIndex(text, "\n")
	ret := &Snippet{
		FileContent: fileContent,
		FilePath:    filePath,
		Start:       NewPosition(0, 0, 0),
		End:         NewPosition(len(text), lineIdx, charIdx),
	}
	if config.DebugMode() {
		ret.text = string(fileContent)
	}
	return ret
}

type Snippet struct {
	FileContent []byte
	FilePath    string
	text        string
	Start       Position
	End         Position
}

func (s *Snippet) Fork(start, end Position) *Snippet {
	ret := &Snippet{
		FileContent: s.FileContent,
		FilePath:    s.FilePath,
		Start:       start,
		End:         end,
	}
	if config.DebugMode() {
		ret.text = string(s.FileContent[start.Offset:end.Offset])
	}
	return ret
}

func (s *Snippet) Text() string {
	if s.text != "" {
		return s.text
	}
	return string(s.FileContent[s.Start.Offset:s.End.Offset])
}
