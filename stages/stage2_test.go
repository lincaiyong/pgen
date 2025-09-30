package stages

import (
	"os"
	"testing"
)

func TestStage2(t *testing.T) {
	b, err := os.ReadFile("../../go.txt")
	if err != nil {
		t.Fatal(err)
	}
	s1 := RunStage1(string(b))
	s2 := RunStage2(s1)
	print(s2)
}
