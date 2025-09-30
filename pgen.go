package pgen

import (
	"errors"
	"github.com/lincaiyong/pgen/stages"
	"strings"
)

func Run(input string) (string, error) {
	s1 := stages.RunStage1(input)
	if s1.Error.ToError() != nil {
		return "", s1.Error.ToError()
	}
	s2 := stages.RunStage2(s1)
	if s2.Error.ToError() != nil {
		return "", s2.Error.ToError()
	}
	s31 := stages.RunStage31(s2)
	s32 := stages.RunStage32(s2)
	s33 := stages.RunStage33(s2)

	if s31.Error.ToError() != nil || s32.Error.ToError() != nil || s33.Error.ToError() != nil {
		var sb strings.Builder
		if s31.Error.ToError() != nil {
			sb.WriteString(s31.Error.ToError().Error())
		}
		if s32.Error.ToError() != nil {
			sb.WriteString(s32.Error.ToError().Error())
		}
		if s33.Error.ToError() != nil {
			sb.WriteString(s33.Error.ToError().Error())
		}
		return "", errors.New(sb.String())
	}
	s4 := stages.RunStage4(s31, s32, s33)
	if s4.Error.ToError() != nil {
		return "", s4.Error.ToError()
	}
	output := strings.TrimRight(s4.Gen.String(), "\n") + "\n"
	return output, nil
}
