package stages

import (
	"os"
	"testing"
)

func TestStage33(t *testing.T) {
	b, err := os.ReadFile("../../go.txt")
	if err != nil {
		t.Fatal(err)
	}
	s1 := RunStage1(string(b))
	s2 := RunStage2(s1)
	s33 := RunStage33(s2)
	text := s33.Gen.String()
	_ = os.WriteFile("test3.txt", []byte(text), 0644)
}
