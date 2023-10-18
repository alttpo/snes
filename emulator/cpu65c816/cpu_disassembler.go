// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cpu65c816

import (
	"strconv"
)

const hexdigits = "0123456789abcdef"

type xbuf []byte

func (b *xbuf) Db(cycles byte) *xbuf {
	*b = strconv.AppendInt(*b, int64(cycles), 10)
	return b
}

func (b *xbuf) C(d byte) *xbuf {
	*b = append(*b, d)
	return b
}

func (b *xbuf) X02(d uint8) *xbuf {
	*b = append(*b, hexdigits[d>>4&0xF], hexdigits[d&0xF])
	return b
}

func (b *xbuf) X04(d uint16) *xbuf {
	*b = append(*b, hexdigits[d>>12&0xF], hexdigits[d>>8&0xF], hexdigits[d>>4&0xF], hexdigits[d&0xF])
	return b
}

func (b *xbuf) S(s string) *xbuf {
	sb := []byte(s)
	for i := 0; i < len(sb); i++ {
		*b = append(*b, sb[i])
	}
	return b
}

func (b *xbuf) Sb(sb []byte) *xbuf {
	for i := 0; i < len(sb); i++ {
		*b = append(*b, sb[i])
	}
	return b
}

func (b *xbuf) Sn(s string, n int) *xbuf {
	sb := []byte(s)
	if n <= len(sb) {
		for i := 0; i < n; i++ {
			*b = append(*b, sb[i])
		}
	}
	if n > len(sb) {
		for i := len(sb); i < n; i++ {
			*b = append(*b, ' ')
		}
	}
	return b
}

func appendCPUFlags(o xbuf, flag byte, name byte) xbuf {
	if flag > 0 {
		o.C(name)
	} else {
		o.C('-')
	}
	return o
}

func (c *CPU) DisassembleCurrentPC(o []byte) []byte {
	//fmt.Fprintf(w, "\n%s", c.Disassemble(c.PC))
	return c.DisassembleTo(c.PC, o)
}

func (c *CPU) DisassembleTo(myPC uint16, o []byte) []byte {
	xb := xbuf(o)

	//var myPC uint16 = c.PC

	//opcode := c.Read(myPC)
	opcode := c.nRead(c.RK, myPC)
	mode := c.instructions[opcode].mode

	// crude and incosistent size adjust
	var sizeAdjust byte
	if mode == m_Immediate_flagM {
		sizeAdjust = c.M
	}
	if mode == m_Immediate_flagX {
		sizeAdjust = c.X
	}

	bytes := c.instructions[opcode].size - sizeAdjust
	name := c.instructions[opcode].name

	xb.Db(c.Cycles).C('\t').X02(c.RK).C(':').X04(myPC).C('|')

	switch bytes {
	case 4:
		w0 := c.nRead(c.RK, myPC+0)
		w1 := c.nRead(c.RK, myPC+1)
		w2 := c.nRead(c.RK, myPC+2)
		w3 := c.nRead(c.RK, myPC+3)
		//o = fmt.Appendf(o, "%d\t%02x:%04x│%02x %02x %02x %02x│%3s ",
		//	c.Cycles, c.RK, myPC, w0, w1, w2, w3, name)
		xb.X02(w0).C(' ').X02(w1).C(' ').X02(w2).C(' ').X02(w3)
		xb.C('|').Sn(name, 3).C(' ')
		xb = c.formatInstructionModeTo(xb, mode, w0, w1, w2, w3)
	case 3:
		w0 := c.nRead(c.RK, myPC+0)
		w1 := c.nRead(c.RK, myPC+1)
		w2 := c.nRead(c.RK, myPC+2)
		//o = fmt.Appendf(o, "%d\t%02x:%04x│%02x %02x %02x   │%3s ",
		//	c.Cycles, c.RK, myPC, w0, w1, w2, name)
		xb.X02(w0).C(' ').X02(w1).C(' ').X02(w2).S("   ")
		xb.C('|').Sn(name, 3).C(' ')
		xb = c.formatInstructionModeTo(xb, mode, w0, w1, w2, 0)
	case 2:
		w0 := c.nRead(c.RK, myPC+0)
		w1 := c.nRead(c.RK, myPC+1)
		//o = fmt.Appendf(o, "%d\t%02x:%04x│%02x %02x      │%3s ",
		//	c.Cycles, c.RK, myPC, w0, w1, name)
		xb.X02(w0).C(' ').X02(w1).S("      ")
		xb.C('|').Sn(name, 3).C(' ')
		xb = c.formatInstructionModeTo(xb, mode, w0, w1, 0, 0)
	case 1:
		w0 := c.nRead(c.RK, myPC+0)
		//o = fmt.Appendf(o, "%d\t%02x:%04x│%02x         │%3s ",
		//	c.Cycles, c.RK, myPC, w0, name)
		xb.X02(w0).S("         ")
		xb.C('|').Sn(name, 3).C(' ')
		xb = c.formatInstructionModeTo(xb, mode, w0, 0, 0, 0)
	default:
		//o = fmt.Appendf(o, "%d\t%02x:%04x│err: cmd len %d│%3s ",
		//	c.Cycles, c.RK, myPC, bytes, name)
		xb.S("???        ")
		xb.C('|').Sn(name, 3).C(' ')
		xb = c.formatInstructionModeTo(xb, mode, 0, 0, 0, 0)
	}
	xb.C('|')

	if c.M == 0 && c.X == 0 {
		xb.S(" A=").X04(c.RA)
		xb.S(" X=").X04(c.RX)
		xb.S(" Y=").X04(c.RY)
	} else if c.M == 0 && c.X != 0 {
		xb.S(" A=").X04(c.RA)
		xb.S(" X=--").X02(c.RXl)
		xb.S(" Y=--").X02(c.RYl)
	} else if c.M != 0 && c.X == 0 {
		xb.S(" A=--").X02(c.RAl)
		xb.S(" X=").X04(c.RX)
		xb.S(" Y=").X04(c.RY)
	} else if c.M != 0 && c.X != 0 {
		xb.S(" A=--").X02(c.RAl)
		xb.S(" X=--").X02(c.RXl)
		xb.S(" Y=--").X02(c.RYl)
	}

	xb.C(' ')
	xb = appendCPUFlags(xb, c.N, 'N')
	xb = appendCPUFlags(xb, c.V, 'V')
	xb = appendCPUFlags(xb, c.M, 'M')
	xb = appendCPUFlags(xb, c.X, 'X')
	xb = appendCPUFlags(xb, c.D, 'D')
	xb = appendCPUFlags(xb, c.I, 'I')
	xb = appendCPUFlags(xb, c.Z, 'Z')
	xb = appendCPUFlags(xb, c.C, 'C')
	xb.C('\n')

	return xb
}

var spaces = [13]byte{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '}

func (c *CPU) formatInstructionModeTo(xb xbuf, mode byte, w0 byte, w1 byte, w2 byte, w3 byte) xbuf {
	switch mode {
	case m_Absolute: // $9876       - p. 288 or 5.2
		//o = fmt.Appendf(o, "$%02x%02x", w2, w1)
		xb.C('$').X02(w2).X02(w1).Sb(spaces[5:13])
	case m_Absolute_X: // $9876, X    - p. 289 or 5.3
		//o = fmt.Appendf(o, "$%02x%02x, X", w2, w1)
		xb.C('$').X02(w2).X02(w1).S(", X").Sb(spaces[8:13])
	case m_Absolute_Y: // $9876, Y    - p. 290 or 5.3
		//o = fmt.Appendf(o, "$%02x%02x, Y", w2, w1)
		xb.C('$').X02(w2).X02(w1).S(", Y").Sb(spaces[8:13])
	case m_Accumulator: // A           - p. 296 or 5.6
		//o = fmt.Appendf(o, "A")
		xb.C('A').Sb(spaces[1:13])
	case m_Immediate: // #$aa        - p. 306 or 5.14
		if w0 == 0xf4 {
			//o = fmt.Appendf(o, "#$%02x%02x", w2, w1)
			xb.C('#').C('$').X02(w2).X02(w1).Sb(spaces[6:13])
		} else {
			//o = fmt.Appendf(o, "#$%02x", w1)
			xb.C('#').C('$').X02(w1).Sb(spaces[4:13])
		}
	case m_Immediate_flagM: // #$aaaa/$#aa - p. 306 or 5.14
		if c.M == 1 {
			//o = fmt.Appendf(o, "#$%02x", w1)
			xb.C('#').C('$').X02(w1).Sb(spaces[4:13])
		} else {
			//o = fmt.Appendf(o, "#$%02x%02x", w2, w1)
			xb.C('#').C('$').X02(w2).X02(w1).Sb(spaces[6:13])
		}
	case m_Immediate_flagX: // #$aa        - p. 306 or 5.14 // XXX fix it
		if c.X == 1 {
			//o = fmt.Appendf(o, "#$%02x", w1)
			xb.C('#').C('$').X02(w1).Sb(spaces[4:13])
		} else {
			//o = fmt.Appendf(o, "#$%02x%02x", w2, w1)
			xb.C('#').C('$').X02(w2).X02(w1).Sb(spaces[6:13])
		}
	case m_Implied: // -           - p. 307 or 5.15
		xb.Sb(spaces[:13])
	case m_DP: // $12         - p. 298 or 5.7
		//o = fmt.Appendf(o, "$%02x", w1)
		xb.C('$').X02(w1).Sb(spaces[3:13])
	case m_DP_X: // $12, X      - p. 299 or 5.8
		//o = fmt.Appendf(o, "$%02x, X", w1)
		xb.C('$').X02(w1).S(", X").Sb(spaces[6:13])
	case m_DP_Y: // $12, Y      - p. 300 or 5.8
		//o = fmt.Appendf(o, "$%02x, Y", w1)
		xb.C('$').X02(w1).S(", Y").Sb(spaces[6:13])
	case m_DP_X_Indirect: // ($12, X)    - p. 301 or 5.11
		//o = fmt.Appendf(o, "($%02x, X)", w1)
		xb.C('(').C('$').X02(w1).S(", X)").Sb(spaces[8:13])
	case m_DP_Indirect: // ($12)       - p. 302 or 5.9
		//o = fmt.Appendf(o, "($%02x)", w1)
		xb.C('(').C('$').X02(w1).C(')').Sb(spaces[5:13])
	case m_DP_Indirect_Long: // [$12]       - p. 303 or 5.10
		//o = fmt.Appendf(o, "[$%02x]", w1)
		xb.C('[').C('$').X02(w1).C(']').Sb(spaces[5:13])
	case m_DP_Indirect_Y: // ($12), Y    - p. 304 or 5.12
		//o = fmt.Appendf(o, "($%02x), Y", w1)
		xb.C('(').C('$').X02(w1).S("), Y").Sb(spaces[8:13])
	case m_DP_Indirect_Long_Y: // [$12], Y    - p. 305 or 5.13
		//o = fmt.Appendf(o, "[$%02x], Y", w1)
		xb.C('[').C('$').X02(w1).S("], Y").Sb(spaces[8:13])
	case m_Absolute_X_Indirect: // ($1234, X)  - p. 291 or 5.5
		//o = fmt.Appendf(o, "($%02x%02x, X)", w2, w1)
		xb.C('(').C('$').X02(w2).X02(w1).S(", X)").Sb(spaces[10:13])
	case m_Absolute_Indirect: // ($1234)     - p. 292 or 5.4
		//o = fmt.Appendf(o, "($%02x%02x)", w2, w1)
		xb.C('(').C('$').X02(w2).X02(w1).C(')').Sb(spaces[7:13])
	case m_Absolute_Indirect_Long: // [$1234]     - p. 293 or 5.10
		//o = fmt.Appendf(o, "[$%02x%02x]", w2, w1)
		xb.C('[').C('$').X02(w2).X02(w1).C(']').Sb(spaces[7:13])
	case m_Absolute_Long: // $abcdef     - p. 294 or 5.16
		//o = fmt.Appendf(o, "$%02x%02x%02x", w3, w2, w1)
		xb.C('$').X02(w3).X02(w2).X02(w1).Sb(spaces[7:13])
	case m_Absolute_Long_X: // $abcdex, X  - p. 295 or 5.17
		//o = fmt.Appendf(o, "$%02x%02x%02x, X", w3, w2, w1)
		xb.C('$').X02(w3).X02(w2).X02(w1).S(", X").Sb(spaces[10:13])
	case m_BlockMove: // #$12,#$34   - p. 297 or 5.19 (MVN, MVP)
		//o = fmt.Appendf(o, "#$%02x,#$%02x", w2, w1) // XXX - verify it!
		xb.C('#').C('$').X02(w2).C(',').C('#').C('$').X02(w1).Sb(spaces[9:13])
	case m_PC_Relative: // rel8        - p. 308 or 5.18 (BRA)
		w216 := uint16(w1)
		if w2 < 0x80 {
			dest := c.PC + 2 + w216
			//o = fmt.Appendf(o, "$%02x ($%04x +)", w216, dest)
			xb.C('$').X02(w1).S(" ($").X04(dest).S(" +)").Sb(spaces[13:13])
		} else {
			dest := c.PC + 2 + w216 - 0x100
			//o = fmt.Appendf(o, "$%02x ($%04x -)", w216, dest)
			xb.C('$').X02(w1).S(" ($").X04(dest).S(" -)").Sb(spaces[13:13])
		}
	case m_PC_Relative_Long: // rel16       - p. 309 or 5.18 (BRL)
		arg16 := uint16(w2)<<8 | uint16(w1)
		addr := c.PC + 3 + arg16
		//o = fmt.Appendf(o, "$%04x", addr)
		xb.C('$').X04(addr).Sb(spaces[5:13])
	case m_Stack_Relative: // $32, Sn      - p. 324 or 5.20
		//o = fmt.Appendf(o, "$%02x, Sn", w1)
		xb.C('$').X02(w1).S(", Sn").Sb(spaces[9:13])
	case m_Stack_Relative_Indirect_Y: // ($32, Sn), Y - p. 325 or 5.21 (STACK,Sn),Y
		//o = fmt.Appendf(o, "($%02x, Sn), Y", w1)
		xb.C('$').C('(').X02(w1).S(", Sn), Y").Sb(spaces[12:13])
	default:
		//o = fmt.Appendf(o, "! unknown !")
		xb.S("! unknown !").Sb(spaces[11:13])
	}

	return xb
}
