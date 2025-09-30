package stages

import (
	"os"
	"testing"
)

func TestStage31(t *testing.T) {
	b, err := os.ReadFile("../../go.txt")
	if err != nil {
		t.Fatal(err)
	}
	s1 := RunStage1(string(b))
	s2 := RunStage2(s1)
	s31 := RunStage31(s2)
	text := s31.Gen.String()
	_ = os.WriteFile("test.txt", []byte(text), 0644)
}
