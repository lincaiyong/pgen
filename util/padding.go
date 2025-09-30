package util

func MakePadding(count int, c byte) string {
	buf := make([]byte, 0)
	for i := 0; i < count; i++ {
		buf = append(buf, c)
	}
	return string(buf)
}
