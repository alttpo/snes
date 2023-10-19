package xbuf

import "strconv"

const hexdigits = "0123456789abcdef"

type B []byte

func (b *B) Db(cycles byte) *B {
	*b = strconv.AppendInt(*b, int64(cycles), 10)
	return b
}

func (b *B) C(d byte) *B {
	*b = append(*b, d)
	return b
}

func (b *B) X02(d uint8) *B {
	*b = append(*b, hexdigits[d>>4&0xF], hexdigits[d&0xF])
	return b
}

func (b *B) X04(d uint16) *B {
	*b = append(*b, hexdigits[d>>12&0xF], hexdigits[d>>8&0xF], hexdigits[d>>4&0xF], hexdigits[d&0xF])
	return b
}

func (b *B) X06(d uint32) *B {
	*b = append(*b, hexdigits[d>>20&0xF], hexdigits[d>>16&0xF], hexdigits[d>>12&0xF], hexdigits[d>>8&0xF], hexdigits[d>>4&0xF], hexdigits[d&0xF])
	return b
}

func (b *B) S(s string) *B {
	sb := []byte(s)
	for i := 0; i < len(sb); i++ {
		*b = append(*b, sb[i])
	}
	return b
}

func (b *B) Sb(sb []byte) *B {
	for i := 0; i < len(sb); i++ {
		*b = append(*b, sb[i])
	}
	return b
}

func (b *B) Sn(s string, n int) *B {
	sb := []byte(s)
	for i := 0; i < len(sb); i++ {
		*b = append(*b, sb[i])
	}
	if n > len(sb) {
		for i := len(sb); i < n; i++ {
			*b = append(*b, ' ')
		}
	}
	return b
}
