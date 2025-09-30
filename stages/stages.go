package stages

func Run(content string) error {
	s1 := RunStage1(content)
	if len(s1.Error.Errors()) > 0 {
		return s1.Error.ToError()
	}
	s2 := RunStage2(s1)
	if len(s2.Error.Errors()) > 0 {
		return s2.Error.ToError()
	}
	s31 := RunStage31(s2)
	if len(s31.Error.Errors()) > 0 {
		return s31.Error.ToError()
	}
	s32 := RunStage31(s2)
	if len(s32.Error.Errors()) > 0 {
		return s32.Error.ToError()
	}
	return nil
}
