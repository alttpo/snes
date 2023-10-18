// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cpu65c816

import (
	"fmt"
)

func appendCPUFlags(o []byte, flag byte, name string) []byte {
	if flag > 0 {
		return append(o, name...)
	} else {
		return append(o, "-"...)
	}
}

func (c *CPU) DisassembleCurrentPC(o []byte) []byte {
	//fmt.Fprintf(w, "\n%s", c.Disassemble(c.PC))
	return c.DisassembleTo(c.PC, o)
}

func (c *CPU) DisassembleTo(myPC uint16, o []byte) []byte {
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

	//if c.Cycles == 0 {
	//	_, _ = fmt.Appendf(o, "--:----│           │                 │")
	//}

	switch bytes {
	case 4:
		w0 := c.nRead(c.RK, myPC+0)
		w1 := c.nRead(c.RK, myPC+1)
		w2 := c.nRead(c.RK, myPC+2)
		w3 := c.nRead(c.RK, myPC+3)
		o = fmt.Appendf(o, "%d\t%02x:%04x│%02x %02x %02x %02x│%3s ",
			c.Cycles, c.RK, myPC, w0, w1, w2, w3, name)
		o = c.formatInstructionModeTo(o, mode, w0, w1, w2, w3)
		o = fmt.Appendf(o, "│")
	case 3:
		w0 := c.nRead(c.RK, myPC+0)
		w1 := c.nRead(c.RK, myPC+1)
		w2 := c.nRead(c.RK, myPC+2)
		o = fmt.Appendf(o, "%d\t%02x:%04x│%02x %02x %02x   │%3s ",
			c.Cycles, c.RK, myPC, w0, w1, w2, name)
		o = c.formatInstructionModeTo(o, mode, w0, w1, w2, 0)
		o = fmt.Appendf(o, "│")
	case 2:
		w0 := c.nRead(c.RK, myPC+0)
		w1 := c.nRead(c.RK, myPC+1)
		o = fmt.Appendf(o, "%d\t%02x:%04x│%02x %02x      │%3s ",
			c.Cycles, c.RK, myPC, w0, w1, name)
		o = c.formatInstructionModeTo(o, mode, w0, w1, 0, 0)
		o = fmt.Appendf(o, "│")
	case 1:
		w0 := c.nRead(c.RK, myPC+0)
		o = fmt.Appendf(o, "%d\t%02x:%04x│%02x         │%3s ",
			c.Cycles, c.RK, myPC, w0, name)
		o = c.formatInstructionModeTo(o, mode, w0, 0, 0, 0)
		o = fmt.Appendf(o, "│")
	default:
		o = fmt.Appendf(o, "%d\t%02x:%04x│err: cmd len %d│%3s ",
			c.Cycles, c.RK, myPC, bytes, name)
		o = c.formatInstructionModeTo(o, mode, 0, 0, 0, 0)
		o = fmt.Appendf(o, "│")
	}

	if c.M == 0 && c.X == 0 {
		o = fmt.Appendf(o, " A=%04x X=%04x Y=%04x", c.RA, c.RX, c.RY)
	} else if c.M == 0 && c.X != 0 {
		o = fmt.Appendf(o, " A=%04x X=--%02x Y=--%02x", c.RA, c.RXl, c.RYl)
	} else if c.M != 0 && c.X == 0 {
		o = fmt.Appendf(o, " A=--%02x X=%04x Y=%04x", c.RAl, c.RX, c.RY)
	} else if c.M != 0 && c.X != 0 {
		o = fmt.Appendf(o, " A=--%02x X=--%02x Y=--%02x", c.RAl, c.RXl, c.RYl)
	}
	o = append(o, " "...)
	o = appendCPUFlags(o, c.N, "N")
	o = appendCPUFlags(o, c.V, "V")
	o = appendCPUFlags(o, c.M, "M")
	o = appendCPUFlags(o, c.X, "X")
	o = appendCPUFlags(o, c.D, "D")
	o = appendCPUFlags(o, c.I, "I")
	o = appendCPUFlags(o, c.Z, "Z")
	o = appendCPUFlags(o, c.C, "C")
	o = append(o, "\n"...)

	return o
}

var spaces = [13]byte{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '}

func (c *CPU) formatInstructionModeTo(t []byte, mode byte, w0 byte, w1 byte, w2 byte, w3 byte) []byte {
	oa := [13]byte{}
	o := oa[:0]

	switch mode {
	case m_Absolute: // $9876       - p. 288 or 5.2
		o = fmt.Appendf(o, "$%02x%02x", w2, w1)
	case m_Absolute_X: // $9876, X    - p. 289 or 5.3
		o = fmt.Appendf(o, "$%02x%02x, X", w2, w1)
	case m_Absolute_Y: // $9876, Y    - p. 290 or 5.3
		o = fmt.Appendf(o, "$%02x%02x, Y", w2, w1)
	case m_Accumulator: // A           - p. 296 or 5.6
		o = fmt.Appendf(o, "A")
	case m_Immediate: // #$aa        - p. 306 or 5.14
		if w0 == 0xf4 {
			o = fmt.Appendf(o, "#$%02x%02x", w2, w1)
		} else {
			o = fmt.Appendf(o, "#$%02x", w1)
		}
	case m_Immediate_flagM: // #$aaaa/$#aa - p. 306 or 5.14
		if c.M == 1 {
			o = fmt.Appendf(o, "#$%02x", w1)
		} else {
			o = fmt.Appendf(o, "#$%02x%02x", w2, w1)
		}
	case m_Immediate_flagX: // #$aa        - p. 306 or 5.14 // XXX fix it
		if c.X == 1 {
			o = fmt.Appendf(o, "#$%02x", w1)
		} else {
			o = fmt.Appendf(o, "#$%02x%02x", w2, w1)
		}
	case m_Implied: // -           - p. 307 or 5.15
		//n = 0
	case m_DP: // $12         - p. 298 or 5.7
		o = fmt.Appendf(o, "$%02x", w1)
	case m_DP_X: // $12, X      - p. 299 or 5.8
		o = fmt.Appendf(o, "$%02x, X", w1)
	case m_DP_Y: // $12, Y      - p. 300 or 5.8
		o = fmt.Appendf(o, "$%02x, Y", w1)
	case m_DP_X_Indirect: // ($12, X)    - p. 301 or 5.11
		o = fmt.Appendf(o, "($%02x, X)", w1)
	case m_DP_Indirect: // ($12)       - p. 302 or 5.9
		o = fmt.Appendf(o, "($%02x)", w1)
	case m_DP_Indirect_Long: // [$12]       - p. 303 or 5.10
		o = fmt.Appendf(o, "[$%02x]", w1)
	case m_DP_Indirect_Y: // ($12), Y    - p. 304 or 5.12
		o = fmt.Appendf(o, "($%02x), Y", w1)
	case m_DP_Indirect_Long_Y: // [$12], Y    - p. 305 or 5.13
		o = fmt.Appendf(o, "[$%02x], Y", w1)
	case m_Absolute_X_Indirect: // ($1234, X)  - p. 291 or 5.5
		o = fmt.Appendf(o, "($%02x%02x, X)", w2, w1)
	case m_Absolute_Indirect: // ($1234)     - p. 292 or 5.4
		o = fmt.Appendf(o, "($%02x%02x)", w2, w1)
	case m_Absolute_Indirect_Long: // [$1234]     - p. 293 or 5.10
		o = fmt.Appendf(o, "[$%02x%02x]", w2, w1)
	case m_Absolute_Long: // $abcdef     - p. 294 or 5.16
		o = fmt.Appendf(o, "$%02x%02x%02x", w3, w2, w1)
	case m_Absolute_Long_X: // $abcdex, X  - p. 295 or 5.17
		o = fmt.Appendf(o, "$%02x%02x%02x, X", w3, w2, w1)
	case m_BlockMove: // #$12,#$34   - p. 297 or 5.19 (MVN, MVP)
		o = fmt.Appendf(o, "#$%02x,#$%02x", w2, w1) // XXX - verify it!
	case m_PC_Relative: // rel8        - p. 308 or 5.18 (BRA)
		w216 := uint16(w1)
		if w2 < 0x80 {
			dest := c.PC + 2 + w216
			o = fmt.Appendf(o, "$%02x ($%04x +)", w216, dest)
		} else {
			dest := c.PC + 2 + w216 - 0x100
			o = fmt.Appendf(o, "$%02x ($%04x -)", w216, dest)
		}
	case m_PC_Relative_Long: // rel16       - p. 309 or 5.18 (BRL)
		arg16 := uint16(w2)<<8 | uint16(w1)
		addr := c.PC + 3 + arg16
		o = fmt.Appendf(o, "$%04x", addr)
	case m_Stack_Relative: // $32, S      - p. 324 or 5.20
		o = fmt.Appendf(o, "$%02x, S", w1)
	case m_Stack_Relative_Indirect_Y: // ($32, S), Y - p. 325 or 5.21 (STACK,S),Y
		o = fmt.Appendf(o, "($%02x, S), Y", w1)
	default:
		o = fmt.Appendf(o, "! unknown !")
	}

	// emit up to 13 spaces for alignment:
	t = append(t, o...)
	t = append(t, spaces[len(o):13]...)

	return t
}
