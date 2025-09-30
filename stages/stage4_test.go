package stages

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestStage4(t *testing.T) {
	b, err := os.ReadFile("../../go.txt")
	if err != nil {
		t.Fatal(err)
	}
	s1 := RunStage1(string(b))
	s2 := RunStage2(s1)
	s31 := RunStage31(s2)
	s32 := RunStage32(s2)
	s33 := RunStage33(s2)
	s4 := RunStage4(s31, s32, s33)
	text := strings.TrimRight(s4.Gen.String(), "\n") + "\n"
	_ = os.WriteFile("/Users/bytedance/Code/goodfun/2509/08parser/goparser/goparser.go", []byte(text), 0644)

	err = s2.Error.ToError()
	if err != nil {
		fmt.Println(err)
	}
	err = s32.Error.ToError()
	if err != nil {
		fmt.Println(err)
	}
}
