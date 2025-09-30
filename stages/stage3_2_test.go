package stages

import (
	"os"
	"testing"
)

func TestStage32(t *testing.T) {
	b, err := os.ReadFile("../../go.txt")
	if err != nil {
		t.Fatal(err)
	}
	s1 := RunStage1(string(b))
	s2 := RunStage2(s1)
	s32 := RunStage32(s2)
	text := s32.Gen.String()
	_ = os.WriteFile("test2.txt", []byte(text), 0644)
}
