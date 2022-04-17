// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cpu65c816

import (
	"fmt"
	"io"
)

func printCPUFlags(flag byte, name string) string {
	if flag > 0 {
		return name
	} else {
		return "-"
	}
}

func (c *CPU) Disassemble(myPC uint16) string {
	//var myPC uint16 = c.PC
	var numeric string
	var output string

	//opcode := c.Read(myPC)
	opcode := c.nRead(c.RK, myPC)
	mode := instructions[opcode].mode

	// crude and incosistent size adjust
	var sizeAdjust byte
	if mode == m_Immediate_flagM {
		sizeAdjust = c.M
	}
	if mode == m_Immediate_flagX {
		sizeAdjust = c.X
	}

	bytes := instructions[opcode].size - sizeAdjust
	name := instructions[opcode].name

	var arg string
	switch bytes {
	case 4:
		w0 := c.nRead(c.RK, myPC+0)
		w1 := c.nRead(c.RK, myPC+1)
		w2 := c.nRead(c.RK, myPC+2)
		w3 := c.nRead(c.RK, myPC+3)
		numeric = fmt.Sprintf("%02x %02x %02x %02x", w0, w1, w2, w3)
		arg = c.formatInstructionMode(mode, w0, w1, w2, w3)
	case 3:
		w0 := c.nRead(c.RK, myPC+0)
		w1 := c.nRead(c.RK, myPC+1)
		w2 := c.nRead(c.RK, myPC+2)
		numeric = fmt.Sprintf("%02x %02x %02x", w0, w1, w2)
		arg = c.formatInstructionMode(mode, w0, w1, w2, 0)
	case 2:
		w0 := c.nRead(c.RK, myPC+0)
		w1 := c.nRead(c.RK, myPC+1)
		numeric = fmt.Sprintf("%02x %02x", w0, w1)
		arg = c.formatInstructionMode(mode, w0, w1, 0, 0)
	case 1:
		w0 := c.nRead(c.RK, myPC+0)
		numeric = fmt.Sprintf("%02x", w0)
		arg = c.formatInstructionMode(mode, w0, 0, 0, 0)
	default:
		numeric = fmt.Sprintf("err: cmd len %d", bytes)
		arg = c.formatInstructionMode(mode, 0, 0, 0, 0)
	}

	if c.Cycles == 0 {
		output = fmt.Sprintf(output, "--:----│           │                 │")
	}
	// XXX - change to different log system
	//if c.Cycles > 9 {
	//	fmt.Fprintf(logView, "warning: instruction cycles > 10\n")
	//}
	output = fmt.Sprintf("%d\t%02x:%04x│%-11v│%3s %-13v│",
		c.Cycles, c.RK, myPC, numeric, name, arg)

	//fmt.Fprintf(v, "%-38v",   "3│00:000c│02 02      │BEQ 02 ($04fa +)")

	return output
}

func (c *CPU) formatInstructionMode(mode byte, w0 byte, w1 byte, w2 byte, w3 byte) string {
	var arg string

	switch mode {
	case m_Absolute: // $9876       - p. 288 or 5.2
		arg = fmt.Sprintf("$%02x%02x", w2, w1)
	case m_Absolute_X: // $9876, X    - p. 289 or 5.3
		arg = fmt.Sprintf("$%02x%02x, X", w2, w1)
	case m_Absolute_Y: // $9876, Y    - p. 290 or 5.3
		arg = fmt.Sprintf("$%02x%02x, Y", w2, w1)
	case m_Accumulator: // A           - p. 296 or 5.6
		arg = "A"
	case m_Immediate: // #$aa        - p. 306 or 5.14
		if w0 == 0xf4 {
			arg = fmt.Sprintf("#$%02x%02x", w2, w1)
		} else {
			arg = fmt.Sprintf("#$%02x", w1)
		}
	case m_Immediate_flagM: // #$aaaa/$#aa - p. 306 or 5.14
		if c.M == 1 {
			arg = fmt.Sprintf("#$%02x", w1)
		} else {
			arg = fmt.Sprintf("#$%02x%02x", w2, w1)
		}
	case m_Immediate_flagX: // #$aa        - p. 306 or 5.14 // XXX fix it
		if c.X == 1 {
			arg = fmt.Sprintf("#$%02x", w1)
		} else {
			arg = fmt.Sprintf("#$%02x%02x", w2, w1)
		}
	case m_Implied: // -           - p. 307 or 5.15
		arg = ""
	case m_DP: // $12         - p. 298 or 5.7
		arg = fmt.Sprintf("$%02x", w1)
	case m_DP_X: // $12, X      - p. 299 or 5.8
		arg = fmt.Sprintf("$%02x, X", w1)
	case m_DP_Y: // $12, Y      - p. 300 or 5.8
		arg = fmt.Sprintf("$%02x, Y", w1)
	case m_DP_X_Indirect: // ($12, X)    - p. 301 or 5.11
		arg = fmt.Sprintf("($%02x, X)", w1)
	case m_DP_Indirect: // ($12)       - p. 302 or 5.9
		arg = fmt.Sprintf("($%02x)", w1)
	case m_DP_Indirect_Long: // [$12]       - p. 303 or 5.10
		arg = fmt.Sprintf("[$%02x]", w1)
	case m_DP_Indirect_Y: // ($12), Y    - p. 304 or 5.12
		arg = fmt.Sprintf("($%02x), Y", w1)
	case m_DP_Indirect_Long_Y: // [$12], Y    - p. 305 or 5.13
		arg = fmt.Sprintf("[$%02x], Y", w1)
	case m_Absolute_X_Indirect: // ($1234, X)  - p. 291 or 5.5
		arg = fmt.Sprintf("($%02x%02x, X)", w2, w1)
	case m_Absolute_Indirect: // ($1234)     - p. 292 or 5.4
		arg = fmt.Sprintf("($%02x%02x)", w2, w1)
	case m_Absolute_Indirect_Long: // [$1234]     - p. 293 or 5.10
		arg = fmt.Sprintf("[$%02x%02x]", w2, w1)
	case m_Absolute_Long: // $abcdef     - p. 294 or 5.16
		arg = fmt.Sprintf("$%02x%02x%02x", w3, w2, w1)
	case m_Absolute_Long_X: // $abcdex, X  - p. 295 or 5.17
		arg = fmt.Sprintf("$%02x%02x%02x, X", w3, w2, w1)
	case m_BlockMove: // #$12,#$34   - p. 297 or 5.19 (MVN, MVP)
		arg = fmt.Sprintf("#$%02x,#$%02x", w2, w1) // XXX - verify it!
	case m_PC_Relative: // rel8        - p. 308 or 5.18 (BRA)
		w216 := uint16(w1)
		if w2 < 0x80 {
			dest := c.PC + 2 + w216
			arg = fmt.Sprintf("$%02x ($%04x +)", w216, dest)
		} else {
			dest := c.PC + 2 + w216 - 0x100
			arg = fmt.Sprintf("$%02x ($%04x -)", w216, dest)
		}
	case m_PC_Relative_Long: // rel16       - p. 309 or 5.18 (BRL)
		arg16 := uint16(w2)<<8 | uint16(w1)
		addr := c.PC + 3 + arg16
		arg = fmt.Sprintf("$%04x", addr)
	case m_Stack_Relative: // $32, S      - p. 324 or 5.20
		arg = fmt.Sprintf("$%02x, S", w1)
	case m_Stack_Relative_Indirect_Y: // ($32, S), Y - p. 325 or 5.21 (STACK,S),Y
		arg = fmt.Sprintf("($%02x, S), Y", w1)
	default:
		arg = "! unknown !"
	}
	return arg
}

// XXX - create disassemble line
func (c *CPU) DisassemblePreviousPC(w io.Writer) {
	//fmt.Fprintf(w, "\n%s", c.Disassemble(c.PPC))
	c.DisassembleTo(c.PPC, w)
}

func (c *CPU) DisassembleCurrentPC(w io.Writer) {
	//fmt.Fprintf(w, "\n%s", c.Disassemble(c.PC))
	c.DisassembleTo(c.PC, w)
}

func (c *CPU) DisassembleTo(myPC uint16, w io.Writer) {
	//var myPC uint16 = c.PC

	//opcode := c.Read(myPC)
	opcode := c.nRead(c.RK, myPC)
	mode := instructions[opcode].mode

	// crude and incosistent size adjust
	var sizeAdjust byte
	if mode == m_Immediate_flagM {
		sizeAdjust = c.M
	}
	if mode == m_Immediate_flagX {
		sizeAdjust = c.X
	}

	bytes := instructions[opcode].size - sizeAdjust
	name := instructions[opcode].name

	//if c.Cycles == 0 {
	//	_, _ = fmt.Fprintf(w, "--:----│           │                 │")
	//}

	switch bytes {
	case 4:
		w0 := c.nRead(c.RK, myPC+0)
		w1 := c.nRead(c.RK, myPC+1)
		w2 := c.nRead(c.RK, myPC+2)
		w3 := c.nRead(c.RK, myPC+3)
		_, _ = fmt.Fprintf(w, "%d\t%02x:%04x│%02x %02x %02x %02x│%3s ",
			c.Cycles, c.RK, myPC, w0, w1, w2, w3, name)
		c.formatInstructionModeTo(w, mode, w0, w1, w2, w3)
		_, _ = fmt.Fprintf(w, "│")
	case 3:
		w0 := c.nRead(c.RK, myPC+0)
		w1 := c.nRead(c.RK, myPC+1)
		w2 := c.nRead(c.RK, myPC+2)
		_, _ = fmt.Fprintf(w, "%d\t%02x:%04x│%02x %02x %02x   │%3s ",
			c.Cycles, c.RK, myPC, w0, w1, w2, name)
		c.formatInstructionModeTo(w, mode, w0, w1, w2, 0)
		_, _ = fmt.Fprintf(w, "│")
	case 2:
		w0 := c.nRead(c.RK, myPC+0)
		w1 := c.nRead(c.RK, myPC+1)
		_, _ = fmt.Fprintf(w, "%d\t%02x:%04x│%02x %02x      │%3s ",
			c.Cycles, c.RK, myPC, w0, w1, name)
		c.formatInstructionModeTo(w, mode, w0, w1, 0, 0)
		_, _ = fmt.Fprintf(w, "│")
	case 1:
		w0 := c.nRead(c.RK, myPC+0)
		_, _ = fmt.Fprintf(w, "%d\t%02x:%04x│%02x         │%3s ",
			c.Cycles, c.RK, myPC, w0, name)
		c.formatInstructionModeTo(w, mode, w0, 0, 0, 0)
		_, _ = fmt.Fprintf(w, "│")
	default:
		_, _ = fmt.Fprintf(w, "%d\t%02x:%04x│err: cmd len %d│%3s ",
			c.Cycles, c.RK, myPC, bytes, name)
		c.formatInstructionModeTo(w, mode, 0, 0, 0, 0)
		_, _ = fmt.Fprintf(w, "│")
	}

	if c.M == 0 && c.X == 0 {
		_, _ = fmt.Fprintf(w, " A=%04x X=%04x Y=%04x", c.RA, c.RX, c.RY)
	} else if c.M == 0 && c.X != 0 {
		_, _ = fmt.Fprintf(w, " A=%04x X=--%02x Y=--%02x", c.RA, c.RXl, c.RYl)
	} else if c.M != 0 && c.X == 0 {
		_, _ = fmt.Fprintf(w, " A=--%02x X=%04x Y=%04x", c.RAl, c.RX, c.RY)
	} else if c.M != 0 && c.X != 0 {
		_, _ = fmt.Fprintf(w, " A=--%02x X=--%02x Y=--%02x", c.RAl, c.RXl, c.RYl)
	}
	_, _ = fmt.Fprintf(w, " %s%s%s%s%s%s%s%s",
		printCPUFlags(c.N, "N"),
		printCPUFlags(c.V, "V"),
		printCPUFlags(c.M, "M"),
		printCPUFlags(c.X, "X"),
		printCPUFlags(c.D, "D"),
		printCPUFlags(c.I, "I"),
		printCPUFlags(c.Z, "Z"),
		printCPUFlags(c.C, "C"),
	)
}

func (c *CPU) formatInstructionModeTo(w io.Writer, mode byte, w0 byte, w1 byte, w2 byte, w3 byte) {
	var n int

	switch mode {
	case m_Absolute: // $9876       - p. 288 or 5.2
		n, _ = fmt.Fprintf(w, "$%02x%02x", w2, w1)
	case m_Absolute_X: // $9876, X    - p. 289 or 5.3
		n, _ = fmt.Fprintf(w, "$%02x%02x, X", w2, w1)
	case m_Absolute_Y: // $9876, Y    - p. 290 or 5.3
		n, _ = fmt.Fprintf(w, "$%02x%02x, Y", w2, w1)
	case m_Accumulator: // A           - p. 296 or 5.6
		n, _ = fmt.Fprintf(w, "A")
	case m_Immediate: // #$aa        - p. 306 or 5.14
		if w0 == 0xf4 {
			n, _ = fmt.Fprintf(w, "#$%02x%02x", w2, w1)
		} else {
			n, _ = fmt.Fprintf(w, "#$%02x", w1)
		}
	case m_Immediate_flagM: // #$aaaa/$#aa - p. 306 or 5.14
		if c.M == 1 {
			n, _ = fmt.Fprintf(w, "#$%02x", w1)
		} else {
			n, _ = fmt.Fprintf(w, "#$%02x%02x", w2, w1)
		}
	case m_Immediate_flagX: // #$aa        - p. 306 or 5.14 // XXX fix it
		if c.X == 1 {
			n, _ = fmt.Fprintf(w, "#$%02x", w1)
		} else {
			n, _ = fmt.Fprintf(w, "#$%02x%02x", w2, w1)
		}
	case m_Implied: // -           - p. 307 or 5.15
		n = 0
	case m_DP: // $12         - p. 298 or 5.7
		n, _ = fmt.Fprintf(w, "$%02x", w1)
	case m_DP_X: // $12, X      - p. 299 or 5.8
		n, _ = fmt.Fprintf(w, "$%02x, X", w1)
	case m_DP_Y: // $12, Y      - p. 300 or 5.8
		n, _ = fmt.Fprintf(w, "$%02x, Y", w1)
	case m_DP_X_Indirect: // ($12, X)    - p. 301 or 5.11
		n, _ = fmt.Fprintf(w, "($%02x, X)", w1)
	case m_DP_Indirect: // ($12)       - p. 302 or 5.9
		n, _ = fmt.Fprintf(w, "($%02x)", w1)
	case m_DP_Indirect_Long: // [$12]       - p. 303 or 5.10
		n, _ = fmt.Fprintf(w, "[$%02x]", w1)
	case m_DP_Indirect_Y: // ($12), Y    - p. 304 or 5.12
		n, _ = fmt.Fprintf(w, "($%02x), Y", w1)
	case m_DP_Indirect_Long_Y: // [$12], Y    - p. 305 or 5.13
		n, _ = fmt.Fprintf(w, "[$%02x], Y", w1)
	case m_Absolute_X_Indirect: // ($1234, X)  - p. 291 or 5.5
		n, _ = fmt.Fprintf(w, "($%02x%02x, X)", w2, w1)
	case m_Absolute_Indirect: // ($1234)     - p. 292 or 5.4
		n, _ = fmt.Fprintf(w, "($%02x%02x)", w2, w1)
	case m_Absolute_Indirect_Long: // [$1234]     - p. 293 or 5.10
		n, _ = fmt.Fprintf(w, "[$%02x%02x]", w2, w1)
	case m_Absolute_Long: // $abcdef     - p. 294 or 5.16
		n, _ = fmt.Fprintf(w, "$%02x%02x%02x", w3, w2, w1)
	case m_Absolute_Long_X: // $abcdex, X  - p. 295 or 5.17
		n, _ = fmt.Fprintf(w, "$%02x%02x%02x, X", w3, w2, w1)
	case m_BlockMove: // #$12,#$34   - p. 297 or 5.19 (MVN, MVP)
		n, _ = fmt.Fprintf(w, "#$%02x,#$%02x", w2, w1) // XXX - verify it!
	case m_PC_Relative: // rel8        - p. 308 or 5.18 (BRA)
		w216 := uint16(w1)
		if w2 < 0x80 {
			dest := c.PC + 2 + w216
			n, _ = fmt.Fprintf(w, "$%02x ($%04x +)", w216, dest)
		} else {
			dest := c.PC + 2 + w216 - 0x100
			n, _ = fmt.Fprintf(w, "$%02x ($%04x -)", w216, dest)
		}
	case m_PC_Relative_Long: // rel16       - p. 309 or 5.18 (BRL)
		arg16 := uint16(w2)<<8 | uint16(w1)
		addr := c.PC + 3 + arg16
		n, _ = fmt.Fprintf(w, "$%04x", addr)
	case m_Stack_Relative: // $32, S      - p. 324 or 5.20
		n, _ = fmt.Fprintf(w, "$%02x, S", w1)
	case m_Stack_Relative_Indirect_Y: // ($32, S), Y - p. 325 or 5.21 (STACK,S),Y
		n, _ = fmt.Fprintf(w, "($%02x, S), Y", w1)
	default:
		n, _ = fmt.Fprintf(w, "! unknown !")
	}

	// emit up to 13 spaces for alignment:
	spaces := [13]byte{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '}
	_, _ = w.Write(spaces[n:13])

	return
}
