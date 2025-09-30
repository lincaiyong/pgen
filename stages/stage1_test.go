package stages

import (
	"os"
	"testing"
)

func TestStage1(t *testing.T) {
	b, err := os.ReadFile("../../go.txt")
	if err != nil {
		t.Fatal(err)
	}
	s1 := RunStage1(string(b))
	print(s1)
}
