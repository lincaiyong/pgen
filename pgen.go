package pgen

import (
	"errors"
	"github.com/lincaiyong/pgen/stages"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func PreProcess(file string) (string, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	text := string(b)
	dir := filepath.Dir(file)
	reg := regexp.MustCompile(`(?m)^#include\((.+?)\)$`)
	ret := reg.FindAllStringSubmatch(text, -1)
	if len(ret) == 0 {
		return text, nil
	}
	for _, v := range ret {
		name := v[1]
		filePath := filepath.Join(dir, name)
		b, err = os.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		text = strings.ReplaceAll(text, v[0], string(b))
	}
	return text, nil
}

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
