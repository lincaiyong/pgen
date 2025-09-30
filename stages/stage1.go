package stages

import (
	"fmt"
	"github.com/lincaiyong/pgen/models"
	"regexp"
	"strings"
)

func RunStage1(input string) *Stage1 {
	stage1 := Stage1{
		Description: "split into snippets",
		Input:       models.NewSnippet("", []byte(input)),
		Error:       models.NewError(),
	}
	stage1.run()
	return &stage1
}

type Stage1 struct {
	Description string
	Input       *models.Snippet

	Tokens    []*models.Snippet
	Keywords  []*models.Snippet
	Operators []*models.Snippet
	Nodes     []*models.Snippet
	Grammars  []*models.Snippet
	Hack      *models.Snippet

	Error *models.Error
}

func (s *Stage1) run() {
	const SectionCount = 6
	sections := s.getSections(s.Input)
	if len(sections) != SectionCount {
		s.Error.AddError(fmt.Errorf("expected %d parts, got %d", SectionCount, len(sections)))
		return
	}
	s.Tokens = s.ruleSplit(sections[0])
	s.Keywords = s.simpleSplit(sections[1])
	s.Operators = s.simpleSplit(sections[2])
	s.Nodes = s.simpleSplit(sections[3])
	s.Grammars = s.ruleSplit(sections[4])
	s.Hack = sections[5]
}

func (s *Stage1) getSections(snippet *models.Snippet) []*models.Snippet {
	divider := strings.Repeat("-", 120) + "\n"
	parts := strings.Split(snippet.Text(), divider)
	ret := make([]*models.Snippet, len(parts))
	start := snippet.Start
	for i, part := range parts {
		end := start.MoveForward(part)
		ret[i] = snippet.Fork(start, end)
		start = end.MoveForward(divider)
	}
	return ret
}

func (s *Stage1) getSnippets(snippet *models.Snippet, pattern string) []*models.Snippet {
	re := regexp.MustCompile(pattern)
	items := re.FindAllStringSubmatch(snippet.Text(), -1)
	parts := make([]string, len(items))
	for i, item := range items {
		parts[i] = item[1]
	}
	ret := make([]*models.Snippet, len(parts))
	start := snippet.Start
	for i, part := range parts {
		end := start.MoveForward(part)
		ret[i] = snippet.Fork(start, end)
		start = end
	}
	used := strings.Join(parts, "")
	if used != snippet.Text() {
		s.Error.AddError(fmt.Errorf("invalid pattern: %s\ntarget content: %s\nmatch content: %s", pattern, snippet.Text(), used))
	}
	return ret
}

func (s *Stage1) ruleSplit(content *models.Snippet) []*models.Snippet {
	return s.getSnippets(content, `(?s)(\s*\S[^\n]*(?:\n +[^\n]*)*\n*)`)
}

func (s *Stage1) simpleSplit(content *models.Snippet) []*models.Snippet {
	return s.getSnippets(content, `(?s)(\s*[^\n]+\n+)`)
}
