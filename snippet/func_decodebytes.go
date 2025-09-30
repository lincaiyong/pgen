package snippet

const DecodeBytesFunc = `func DecodeBytes(bs []byte) ([]rune, [][3]int) {
	var encoding string
	var r *bufio.Reader

	file := bytes.NewBuffer(bs)

	skipBytes := 0
	// check BOM
	if len(bs) > 2 && bs[0] == 0xef && bs[1] == 0xbb && bs[2] == 0xbf {
		encoding = "utf-8-bom"
		r = bufio.NewReader(file)
		skipBytes = 3
	} else if len(bs) > 1 && bs[0] == 0xff && bs[1] == 0xfe {
		encoding = "utf-16le-bom"
		r = bufio.NewReader(transform.NewReader(file, unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()))
		skipBytes = 2
	} else if len(bs) > 1 && bs[0] == 0xfe && bs[1] == 0xff {
		encoding = "utf-16be-bom"
		r = bufio.NewReader(transform.NewReader(file, unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewDecoder()))
		skipBytes = 2
	} else if utf8.Valid(bs) {
		encoding = "utf-8"
		r = bufio.NewReader(file)
	} else {
		encoding = "gbk"
		r = bufio.NewReader(transform.NewReader(file, simplifiedchinese.GBK.NewDecoder()))
	}

	// r: rune-offset, b: byte-offset, s: size
	offsets := make([][3]int, 0)
	offsets = append(offsets, [3]int{0, 0, skipBytes})
	byteOffset := skipBytes
	result := make([]rune, 0)
	// read and decode text
	for {
		c, s, err := r.ReadRune()
		if err != nil {
			break
		}
		if c == 0xfeff {
			continue
		}
		if s > 1 {
			offsets = append(offsets, [3]int{len(result), byteOffset, s})
		}
		byteOffset += s
		result = append(result, c)
	}

	_ = encoding

	return result, offsets
}`
