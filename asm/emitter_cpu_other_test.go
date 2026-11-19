package asm

// otherTests are the TestEmitter_CPU tests of the instructions outside the ALU and read-modify-write groups.
func otherTests() []cpuTest {
	// a subroutine at $00:9000 and $01:8000 that increments X; the STPs at $00:9100 and $01:9000 are jump targets:
	subs := func(r *cpuRun) {
		r.poke(0x009000, 0xE8, 0x60) // inx; rts
		r.poke(0x018000, 0xE8, 0x6B) // inx; rtl
		r.poke(0x018010, 0x6B)       // rtl
		r.poke(0x009100, 0xDB)       // stp
		r.poke(0x019000, 0xDB)       // stp
	}
	// branch emits a branch that skips an INX when taken:
	branchTests := func(name string, op byte, p uint8, label func(a *Emitter, l string), imm8 func(a *Emitter, m int8)) []cpuTest {
		return []cpuTest{
			{
				name: name, p: p,
				emit:  func(a *Emitter) { label(a, "t"); a.INX(); a.Label("t") },
				bytes: []byte{op, 0x01, 0xE8},
				check: func(r *cpuRun) { r.expectX(0) },
			},
			{
				name: name + "_imm8", p: p,
				emit:  func(a *Emitter) { imm8(a, 1); a.INX() },
				bytes: []byte{op, 0x01, 0xE8},
				check: func(r *cpuRun) { r.expectX(0) },
			},
			{
				name: name + "/backward", p: p,
				init:  func(r *cpuRun) { r.setX(0x10); r.poke(0x008000, 0xDB) },
				emit:  func(a *Emitter) { a.Label("b"); a.INX(); label(a, "b") },
				bytes: []byte{0xE8, op, 0xFD},
				stop:  0x008000, // the INX is overwritten with STP so that the backward branch stops
			},
		}
	}

	tests := []cpuTest{
		// immediates outside the ALU group:
		{
			name: "LDA_imm16_lh",
			emit: func(a *Emitter) { a.LDA_imm16_lh(0x34, 0x12) }, bytes: []byte{0xA9, 0x34, 0x12},
			check: func(r *cpuRun) { r.expectA(0x1234) },
		},
		{
			name: "BIT_imm16_w",
			init: func(r *cpuRun) { r.setA(0x00F0) },
			emit: func(a *Emitter) { a.BIT_imm16_w(0x0F00) }, bytes: []byte{0x89, 0x00, 0x0F},
			check: func(r *cpuRun) { r.expectP(flagZ|flagN|flagV, flagZ); r.expectA(0x00F0) },
		},
		{
			name: "BIT_imm8_b", p: flagM,
			init: func(r *cpuRun) { r.setA(0x00F0) },
			emit: func(a *Emitter) { a.BIT_imm8_b(0x0F) }, bytes: []byte{0x89, 0x0F},
			check: func(r *cpuRun) { r.expectP(flagZ|flagN|flagV, flagZ) },
		},
		{
			name: "LDX_imm16_w",
			emit: func(a *Emitter) { a.LDX_imm16_w(0x1234) }, bytes: []byte{0xA2, 0x34, 0x12},
			check: func(r *cpuRun) { r.expectX(0x1234) },
		},
		{
			name: "LDX_imm8_b", p: flagX,
			emit: func(a *Emitter) { a.LDX_imm8_b(0x80) }, bytes: []byte{0xA2, 0x80},
			check: func(r *cpuRun) { r.expectX(0x80); r.expectP(flagN, flagN) },
		},
		{
			name: "LDY_imm16_w",
			emit: func(a *Emitter) { a.LDY_imm16_w(0x1234) }, bytes: []byte{0xA0, 0x34, 0x12},
			check: func(r *cpuRun) { r.expectY(0x1234) },
		},
		{
			name: "LDY_imm8_b", p: flagX,
			emit: func(a *Emitter) { a.LDY_imm8_b(0x80) }, bytes: []byte{0xA0, 0x80},
			check: func(r *cpuRun) { r.expectY(0x80) },
		},
		{
			name: "CPX_imm16_w",
			init: func(r *cpuRun) { r.setX(0x1234) },
			emit: func(a *Emitter) { a.CPX_imm16_w(0x1234) }, bytes: []byte{0xE0, 0x34, 0x12},
			check: func(r *cpuRun) { r.expectP(flagZ|flagC, flagZ|flagC) },
		},
		{
			name: "CPX_imm8_b", p: flagX,
			init: func(r *cpuRun) { r.setX(0x12) },
			emit: func(a *Emitter) { a.CPX_imm8_b(0x13) }, bytes: []byte{0xE0, 0x13},
			check: func(r *cpuRun) { r.expectP(flagZ|flagC|flagN, flagN) },
		},
		{
			name: "CPY_imm16_w",
			init: func(r *cpuRun) { r.setY(0x1234) },
			emit: func(a *Emitter) { a.CPY_imm16_w(0x1234) }, bytes: []byte{0xC0, 0x34, 0x12},
			check: func(r *cpuRun) { r.expectP(flagZ|flagC, flagZ|flagC) },
		},
		{
			name: "CPY_imm8_b", p: flagX,
			init: func(r *cpuRun) { r.setY(0x12) },
			emit: func(a *Emitter) { a.CPY_imm8_b(0x13) }, bytes: []byte{0xC0, 0x13},
			check: func(r *cpuRun) { r.expectP(flagZ|flagC|flagN, flagN) },
		},

		// BIT, TSB, TRB, STZ:
		{
			name: "BIT_dp",
			init: func(r *cpuRun) { r.setA(0x00F0); r.poke16(0x000110, 0xC00F) },
			emit: func(a *Emitter) { a.BIT_dp(0x10) }, bytes: []byte{0x24, 0x10},
			check: func(r *cpuRun) { r.expectP(flagZ|flagN|flagV, flagZ|flagN|flagV) },
		},
		{
			name: "BIT_abs",
			init: func(r *cpuRun) { r.setA(0x00F0); r.poke16(0x7E1234, 0xC00F) },
			emit: func(a *Emitter) { a.BIT_abs(0x1234) }, bytes: []byte{0x2C, 0x34, 0x12},
			check: func(r *cpuRun) { r.expectP(flagZ|flagN|flagV, flagZ|flagN|flagV) },
		},
		{
			name: "BIT_dp_x",
			init: func(r *cpuRun) { r.setA(0x00F0); r.setX(2); r.poke16(0x000112, 0x400F) },
			emit: func(a *Emitter) { a.BIT_dp_x(0x10) }, bytes: []byte{0x34, 0x10},
			check: func(r *cpuRun) { r.expectP(flagZ|flagN|flagV, flagZ|flagV) },
		},
		{
			name: "BIT_abs_x",
			init: func(r *cpuRun) { r.setA(0x00F0); r.setX(2); r.poke16(0x7E1236, 0x8010) },
			emit: func(a *Emitter) { a.BIT_abs_x(0x1234) }, bytes: []byte{0x3C, 0x34, 0x12},
			check: func(r *cpuRun) { r.expectP(flagZ|flagN|flagV, flagN) },
		},
		{
			name: "TSB_dp",
			init: func(r *cpuRun) { r.setA(0x00F0); r.poke16(0x000110, 0x0F3C) },
			emit: func(a *Emitter) { a.TSB_dp(0x10) }, bytes: []byte{0x04, 0x10},
			check: func(r *cpuRun) { r.expectMem(0x000110, 0xFC, 0x0F); r.expectP(flagZ, 0) },
		},
		{
			name: "TSB_abs",
			init: func(r *cpuRun) { r.setA(0x00F0); r.poke16(0x7E1234, 0x0F0C) },
			emit: func(a *Emitter) { a.TSB_abs(0x1234) }, bytes: []byte{0x0C, 0x34, 0x12},
			check: func(r *cpuRun) { r.expectMem(0x7E1234, 0xFC, 0x0F); r.expectP(flagZ, flagZ) },
		},
		{
			name: "TRB_dp",
			init: func(r *cpuRun) { r.setA(0x00F0); r.poke16(0x000110, 0x0F3C) },
			emit: func(a *Emitter) { a.TRB_dp(0x10) }, bytes: []byte{0x14, 0x10},
			check: func(r *cpuRun) { r.expectMem(0x000110, 0x0C, 0x0F); r.expectP(flagZ, 0) },
		},
		{
			name: "TRB_abs",
			init: func(r *cpuRun) { r.setA(0x00F0); r.poke16(0x7E1234, 0x0F3C) },
			emit: func(a *Emitter) { a.TRB_abs(0x1234) }, bytes: []byte{0x1C, 0x34, 0x12},
			check: func(r *cpuRun) { r.expectMem(0x7E1234, 0x0C, 0x0F) },
		},
		{
			name: "STZ_dp",
			init: func(r *cpuRun) { r.poke16(0x000110, 0xFFFF) },
			emit: func(a *Emitter) { a.STZ_dp(0x10) }, bytes: []byte{0x64, 0x10},
			check: func(r *cpuRun) { r.expectMem(0x000110, 0, 0) },
		},
		{
			name: "STZ_dp_x",
			init: func(r *cpuRun) { r.setX(2); r.poke16(0x000112, 0xFFFF) },
			emit: func(a *Emitter) { a.STZ_dp_x(0x10) }, bytes: []byte{0x74, 0x10},
			check: func(r *cpuRun) { r.expectMem(0x000112, 0, 0) },
		},
		{
			name: "STZ_abs",
			init: func(r *cpuRun) { r.poke16(0x7E1234, 0xFFFF) },
			emit: func(a *Emitter) { a.STZ_abs(0x1234) }, bytes: []byte{0x9C, 0x34, 0x12},
			check: func(r *cpuRun) { r.expectMem(0x7E1234, 0, 0) },
		},
		{
			name: "STZ_abs_x",
			init: func(r *cpuRun) { r.setX(2); r.poke16(0x7E1236, 0xFFFF) },
			emit: func(a *Emitter) { a.STZ_abs_x(0x1234) }, bytes: []byte{0x9E, 0x34, 0x12},
			check: func(r *cpuRun) { r.expectMem(0x7E1236, 0, 0) },
		},

		// index register loads, stores and compares:
		{
			name: "LDX_dp",
			init: func(r *cpuRun) { r.poke16(0x000110, 0x1234) },
			emit: func(a *Emitter) { a.LDX_dp(0x10) }, bytes: []byte{0xA6, 0x10},
			check: func(r *cpuRun) { r.expectX(0x1234) },
		},
		{
			name: "LDX_dp_y",
			init: func(r *cpuRun) { r.setY(2); r.poke16(0x000112, 0x1234) },
			emit: func(a *Emitter) { a.LDX_dp_y(0x10) }, bytes: []byte{0xB6, 0x10},
			check: func(r *cpuRun) { r.expectX(0x1234) },
		},
		{
			name: "LDX_abs",
			init: func(r *cpuRun) { r.poke16(0x7E1234, 0x5678) },
			emit: func(a *Emitter) { a.LDX_abs(0x1234) }, bytes: []byte{0xAE, 0x34, 0x12},
			check: func(r *cpuRun) { r.expectX(0x5678) },
		},
		{
			name: "LDX_abs_y",
			init: func(r *cpuRun) { r.setY(4); r.poke16(0x7E1238, 0x5678) },
			emit: func(a *Emitter) { a.LDX_abs_y(0x1234) }, bytes: []byte{0xBE, 0x34, 0x12},
			check: func(r *cpuRun) { r.expectX(0x5678) },
		},
		{
			name: "LDY_dp",
			init: func(r *cpuRun) { r.poke16(0x000110, 0x1234) },
			emit: func(a *Emitter) { a.LDY_dp(0x10) }, bytes: []byte{0xA4, 0x10},
			check: func(r *cpuRun) { r.expectY(0x1234) },
		},
		{
			name: "LDY_dp_x",
			init: func(r *cpuRun) { r.setX(2); r.poke16(0x000112, 0x1234) },
			emit: func(a *Emitter) { a.LDY_dp_x(0x10) }, bytes: []byte{0xB4, 0x10},
			check: func(r *cpuRun) { r.expectY(0x1234) },
		},
		{
			name: "LDY_abs",
			init: func(r *cpuRun) { r.poke16(0x7E1234, 0x5678) },
			emit: func(a *Emitter) { a.LDY_abs(0x1234) }, bytes: []byte{0xAC, 0x34, 0x12},
			check: func(r *cpuRun) { r.expectY(0x5678) },
		},
		{
			name: "LDY_abs_x",
			init: func(r *cpuRun) { r.setX(2); r.poke16(0x7E1236, 0x5678) },
			emit: func(a *Emitter) { a.LDY_abs_x(0x1234) }, bytes: []byte{0xBC, 0x34, 0x12},
			check: func(r *cpuRun) { r.expectY(0x5678) },
		},
		{
			name: "STX_dp",
			init: func(r *cpuRun) { r.setX(0x1234) },
			emit: func(a *Emitter) { a.STX_dp(0x10) }, bytes: []byte{0x86, 0x10},
			check: func(r *cpuRun) { r.expectMem(0x000110, 0x34, 0x12) },
		},
		{
			name: "STX_dp_y",
			init: func(r *cpuRun) { r.setX(0x1234); r.setY(2) },
			emit: func(a *Emitter) { a.STX_dp_y(0x10) }, bytes: []byte{0x96, 0x10},
			check: func(r *cpuRun) { r.expectMem(0x000112, 0x34, 0x12) },
		},
		{
			name: "STX_abs",
			init: func(r *cpuRun) { r.setX(0x1234) },
			emit: func(a *Emitter) { a.STX_abs(0x1234) }, bytes: []byte{0x8E, 0x34, 0x12},
			check: func(r *cpuRun) { r.expectMem(0x7E1234, 0x34, 0x12) },
		},
		{
			name: "STY_dp",
			init: func(r *cpuRun) { r.setY(0x1234) },
			emit: func(a *Emitter) { a.STY_dp(0x10) }, bytes: []byte{0x84, 0x10},
			check: func(r *cpuRun) { r.expectMem(0x000110, 0x34, 0x12) },
		},
		{
			name: "STY_dp_x",
			init: func(r *cpuRun) { r.setY(0x1234); r.setX(2) },
			emit: func(a *Emitter) { a.STY_dp_x(0x10) }, bytes: []byte{0x94, 0x10},
			check: func(r *cpuRun) { r.expectMem(0x000112, 0x34, 0x12) },
		},
		{
			name: "STY_abs",
			init: func(r *cpuRun) { r.setY(0x1234) },
			emit: func(a *Emitter) { a.STY_abs(0x1234) }, bytes: []byte{0x8C, 0x34, 0x12},
			check: func(r *cpuRun) { r.expectMem(0x7E1234, 0x34, 0x12) },
		},
		{
			name: "CPX_dp",
			init: func(r *cpuRun) { r.setX(0x1234); r.poke16(0x000110, 0x1235) },
			emit: func(a *Emitter) { a.CPX_dp(0x10) }, bytes: []byte{0xE4, 0x10},
			check: func(r *cpuRun) { r.expectP(flagZ|flagC|flagN, flagN) },
		},
		{
			name: "CPX_abs",
			init: func(r *cpuRun) { r.setX(0x1234); r.poke16(0x7E1234, 0x1234) },
			emit: func(a *Emitter) { a.CPX_abs(0x1234) }, bytes: []byte{0xEC, 0x34, 0x12},
			check: func(r *cpuRun) { r.expectP(flagZ|flagC|flagN, flagZ|flagC) },
		},
		{
			name: "CPY_dp",
			init: func(r *cpuRun) { r.setY(0x1234); r.poke16(0x000110, 0x1235) },
			emit: func(a *Emitter) { a.CPY_dp(0x10) }, bytes: []byte{0xC4, 0x10},
			check: func(r *cpuRun) { r.expectP(flagZ|flagC|flagN, flagN) },
		},
		{
			name: "CPY_abs",
			init: func(r *cpuRun) { r.setY(0x1234); r.poke16(0x7E1234, 0x1234) },
			emit: func(a *Emitter) { a.CPY_abs(0x1234) }, bytes: []byte{0xCC, 0x34, 0x12},
			check: func(r *cpuRun) { r.expectP(flagZ|flagC|flagN, flagZ|flagC) },
		},

		// flags:
		{name: "CLC", p: flagC, emit: func(a *Emitter) { a.CLC() }, bytes: []byte{0x18}, check: func(r *cpuRun) { r.expectP(flagC, 0) }},
		{name: "SEC", emit: func(a *Emitter) { a.SEC() }, bytes: []byte{0x38}, check: func(r *cpuRun) { r.expectP(flagC, flagC) }},
		{name: "CLI", p: flagI, emit: func(a *Emitter) { a.CLI() }, bytes: []byte{0x58}, check: func(r *cpuRun) { r.expectP(flagI, 0) }},
		{name: "SEI", emit: func(a *Emitter) { a.SEI() }, bytes: []byte{0x78}, check: func(r *cpuRun) { r.expectP(flagI, flagI) }},
		{name: "CLV", p: flagV, emit: func(a *Emitter) { a.CLV() }, bytes: []byte{0xB8}, check: func(r *cpuRun) { r.expectP(flagV, 0) }},
		{name: "CLD", p: flagD, emit: func(a *Emitter) { a.CLD() }, bytes: []byte{0xD8}, check: func(r *cpuRun) { r.expectP(flagD, 0) }},
		{name: "SED", emit: func(a *Emitter) { a.SED() }, bytes: []byte{0xF8}, check: func(r *cpuRun) { r.expectP(flagD, flagD) }},
		{
			name: "REP", p: flagM | flagX | flagC,
			emit: func(a *Emitter) {
				a.REP(0x31)
				if !a.IsM16bit() || !a.IsX16bit() {
					panic("REP did not track m and x")
				}
			},
			bytes: []byte{0xC2, 0x31},
			check: func(r *cpuRun) { r.expectP(flagM|flagX|flagC, 0) },
		},
		{
			name: "SEP",
			emit: func(a *Emitter) {
				a.SEP(0x31)
				if a.IsM16bit() || a.IsX16bit() {
					panic("SEP did not track m and x")
				}
			},
			bytes: []byte{0xE2, 0x31},
			check: func(r *cpuRun) { r.expectP(flagM|flagX|flagC, flagM|flagX|flagC) },
		},
		{
			name: "XCE/native", emit: func(a *Emitter) { a.XCE() }, bytes: []byte{0xFB},
			check: func(r *cpuRun) {
				if r.s.CPU.E != 0 {
					r.t.Error("E = 1; want 0")
				}
			},
		},
		{
			name: "XCE", p: flagC, emit: func(a *Emitter) { a.XCE() }, bytes: []byte{0xFB},
			check: func(r *cpuRun) {
				if r.s.CPU.E != 1 {
					r.t.Error("E = 0; want 1")
				}
				r.expectP(flagC|flagM|flagX, flagM|flagX)
			},
		},

		// transfers:
		{name: "TAX", init: func(r *cpuRun) { r.setA(0x1234) }, emit: func(a *Emitter) { a.TAX() }, bytes: []byte{0xAA}, check: func(r *cpuRun) { r.expectX(0x1234) }},
		{name: "TXA", init: func(r *cpuRun) { r.setX(0x1234) }, emit: func(a *Emitter) { a.TXA() }, bytes: []byte{0x8A}, check: func(r *cpuRun) { r.expectA(0x1234) }},
		{name: "TAY", init: func(r *cpuRun) { r.setA(0x1234) }, emit: func(a *Emitter) { a.TAY() }, bytes: []byte{0xA8}, check: func(r *cpuRun) { r.expectY(0x1234) }},
		{name: "TYA", init: func(r *cpuRun) { r.setY(0x1234) }, emit: func(a *Emitter) { a.TYA() }, bytes: []byte{0x98}, check: func(r *cpuRun) { r.expectA(0x1234) }},
		{name: "TXY", init: func(r *cpuRun) { r.setX(0x1234) }, emit: func(a *Emitter) { a.TXY() }, bytes: []byte{0x9B}, check: func(r *cpuRun) { r.expectY(0x1234) }},
		{name: "TYX", init: func(r *cpuRun) { r.setY(0x1234) }, emit: func(a *Emitter) { a.TYX() }, bytes: []byte{0xBB}, check: func(r *cpuRun) { r.expectX(0x1234) }},
		{name: "TCS", init: func(r *cpuRun) { r.setA(0x01E0) }, emit: func(a *Emitter) { a.TCS() }, bytes: []byte{0x1B}, check: func(r *cpuRun) { r.expectS(0x01E0) }},
		{name: "TSC", emit: func(a *Emitter) { a.TSC() }, bytes: []byte{0x3B}, check: func(r *cpuRun) { r.expectA(testS) }},
		{name: "TXS", init: func(r *cpuRun) { r.setX(0x01E0) }, emit: func(a *Emitter) { a.TXS() }, bytes: []byte{0x9A}, check: func(r *cpuRun) { r.expectS(0x01E0) }},
		{name: "TSX", emit: func(a *Emitter) { a.TSX() }, bytes: []byte{0xBA}, check: func(r *cpuRun) { r.expectX(testS) }},
		{name: "TCD", init: func(r *cpuRun) { r.setA(0x4300) }, emit: func(a *Emitter) { a.TCD() }, bytes: []byte{0x5B}, check: func(r *cpuRun) { r.expectD(0x4300) }},
		{name: "TDC", emit: func(a *Emitter) { a.TDC() }, bytes: []byte{0x7B}, check: func(r *cpuRun) { r.expectA(testD) }},
		{name: "XBA", init: func(r *cpuRun) { r.setA(0x1234) }, emit: func(a *Emitter) { a.XBA() }, bytes: []byte{0xEB}, check: func(r *cpuRun) { r.expectA(0x3412) }},
		{name: "INX", init: func(r *cpuRun) { r.setX(0x00FF) }, emit: func(a *Emitter) { a.INX() }, bytes: []byte{0xE8}, check: func(r *cpuRun) { r.expectX(0x0100) }},
		{name: "INY", init: func(r *cpuRun) { r.setY(0x00FF) }, emit: func(a *Emitter) { a.INY() }, bytes: []byte{0xC8}, check: func(r *cpuRun) { r.expectY(0x0100) }},
		{name: "DEX", init: func(r *cpuRun) { r.setX(0x0100) }, emit: func(a *Emitter) { a.DEX() }, bytes: []byte{0xCA}, check: func(r *cpuRun) { r.expectX(0x00FF) }},
		{name: "DEY", init: func(r *cpuRun) { r.setY(0x0100) }, emit: func(a *Emitter) { a.DEY() }, bytes: []byte{0x88}, check: func(r *cpuRun) { r.expectY(0x00FF) }},

		// stack:
		{
			name: "PHA", init: func(r *cpuRun) { r.setA(0x1234) }, emit: func(a *Emitter) { a.PHA() }, bytes: []byte{0x48},
			check: func(r *cpuRun) { r.expectS(testS - 2); r.expectMem(testS-1, 0x34, 0x12) },
		},
		{
			name: "PHX", init: func(r *cpuRun) { r.setX(0x1234) }, emit: func(a *Emitter) { a.PHX() }, bytes: []byte{0xDA},
			check: func(r *cpuRun) { r.expectS(testS - 2); r.expectMem(testS-1, 0x34, 0x12) },
		},
		{
			name: "PHY", init: func(r *cpuRun) { r.setY(0x1234) }, emit: func(a *Emitter) { a.PHY() }, bytes: []byte{0x5A},
			check: func(r *cpuRun) { r.expectS(testS - 2); r.expectMem(testS-1, 0x34, 0x12) },
		},
		{
			name: "PHD", emit: func(a *Emitter) { a.PHD() }, bytes: []byte{0x0B},
			check: func(r *cpuRun) { r.expectS(testS - 2); r.expectMem(testS-1, 0x00, 0x01) },
		},
		{
			name: "PHB", emit: func(a *Emitter) { a.PHB() }, bytes: []byte{0x8B},
			check: func(r *cpuRun) { r.expectS(testS - 1); r.expectMem(testS, testDB) },
		},
		{
			name: "PHK", init: func(r *cpuRun) { r.poke(testS, 0xFF) }, emit: func(a *Emitter) { a.PHK() }, bytes: []byte{0x4B},
			check: func(r *cpuRun) { r.expectS(testS - 1); r.expectMem(testS, 0x00) },
		},
		{
			name: "PHP", p: flagC | flagN, emit: func(a *Emitter) { a.PHP() }, bytes: []byte{0x08},
			check: func(r *cpuRun) { r.expectS(testS - 1); r.expectMem(testS, flagC|flagN) },
		},
		{
			name: "PEA", emit: func(a *Emitter) { a.PEA(0xABCD) }, bytes: []byte{0xF4, 0xCD, 0xAB},
			check: func(r *cpuRun) { r.expectS(testS - 2); r.expectMem(testS-1, 0xCD, 0xAB) },
		},
		{
			name: "PEI", init: func(r *cpuRun) { r.poke16(0x000110, 0xABCD) }, emit: func(a *Emitter) { a.PEI(0x10) }, bytes: []byte{0xD4, 0x10},
			check: func(r *cpuRun) { r.expectS(testS - 2); r.expectMem(testS-1, 0xCD, 0xAB) },
		},
		{
			name: "PER_imm16", emit: func(a *Emitter) { a.PER_imm16(0x10) }, bytes: []byte{0x62, 0x10, 0x00},
			check: func(r *cpuRun) { r.expectS(testS - 2); r.expectMem(testS-1, 0x13, 0x80) },
		},
		{
			name: "PER", emit: func(a *Emitter) { a.PER("t"); a.NOP(); a.Label("t") }, bytes: []byte{0x62, 0x01, 0x00, 0xEA},
			check: func(r *cpuRun) { r.expectS(testS - 2); r.expectMem(testS-1, 0x04, 0x80) },
		},
		{
			name: "PER/backward", emit: func(a *Emitter) { a.Label("t"); a.NOP(); a.PER("t") }, bytes: []byte{0xEA, 0x62, 0xFC, 0xFF},
			check: func(r *cpuRun) { r.expectMem(testS-1, 0x00, 0x80) },
		},
		{
			name: "PLA", init: func(r *cpuRun) { r.s.CPU.SP = testS - 2; r.poke16(testS-1, 0x1234) }, emit: func(a *Emitter) { a.PLA() }, bytes: []byte{0x68},
			check: func(r *cpuRun) { r.expectS(testS); r.expectA(0x1234) },
		},
		{
			name: "PLX", init: func(r *cpuRun) { r.s.CPU.SP = testS - 2; r.poke16(testS-1, 0x1234) }, emit: func(a *Emitter) { a.PLX() }, bytes: []byte{0xFA},
			check: func(r *cpuRun) { r.expectS(testS); r.expectX(0x1234) },
		},
		{
			name: "PLY", init: func(r *cpuRun) { r.s.CPU.SP = testS - 2; r.poke16(testS-1, 0x1234) }, emit: func(a *Emitter) { a.PLY() }, bytes: []byte{0x7A},
			check: func(r *cpuRun) { r.expectS(testS); r.expectY(0x1234) },
		},
		{
			name: "PLD", init: func(r *cpuRun) { r.s.CPU.SP = testS - 2; r.poke16(testS-1, 0x4300) }, emit: func(a *Emitter) { a.PLD() }, bytes: []byte{0x2B},
			check: func(r *cpuRun) { r.expectS(testS); r.expectD(0x4300) },
		},
		{
			name: "PLB", init: func(r *cpuRun) { r.s.CPU.SP = testS - 1; r.poke(testS, 0x7F) }, emit: func(a *Emitter) { a.PLB() }, bytes: []byte{0xAB},
			check: func(r *cpuRun) { r.expectS(testS); r.expectDB(0x7F) },
		},
		{
			name: "PLP", init: func(r *cpuRun) { r.s.CPU.SP = testS - 1; r.poke(testS, flagC|flagV) }, emit: func(a *Emitter) { a.PLP() }, bytes: []byte{0x28},
			check: func(r *cpuRun) { r.expectS(testS); r.expectP(0xFF, flagC|flagV) },
		},

		// block moves:
		{
			name: "MVN",
			init: func(r *cpuRun) {
				r.setA(2)
				r.setX(0x4000)
				r.setY(0x2000)
				r.poke(0x7F4000, 1, 2, 3, 4)
			},
			emit: func(a *Emitter) { a.MVN(0x7E, 0x7F) }, bytes: []byte{0x54, 0x7E, 0x7F},
			check: func(r *cpuRun) {
				r.expectMem(0x7E2000, 1, 2, 3, 0)
				r.expectA(0xFFFF)
				r.expectX(0x4003)
				r.expectY(0x2003)
				r.expectDB(0x7E)
			},
		},
		{
			name: "MVP",
			init: func(r *cpuRun) {
				r.setA(2)
				r.setX(0x4002)
				r.setY(0x2002)
				r.poke(0x7F4000, 1, 2, 3, 4)
			},
			emit: func(a *Emitter) { a.MVP(0x7E, 0x7F) }, bytes: []byte{0x44, 0x7E, 0x7F},
			check: func(r *cpuRun) {
				r.expectMem(0x7E1FFF, 0, 1, 2, 3)
				r.expectA(0xFFFF)
				r.expectX(0x3FFF)
				r.expectY(0x1FFF)
			},
		},

		// jumps, calls and returns:
		{
			name: "JMP_abs", init: subs, emit: func(a *Emitter) { a.JMP_abs("t"); a.INX(); a.Label("t") }, bytes: []byte{0x4C, 0x04, 0x80, 0xE8},
			check: func(r *cpuRun) { r.expectX(0) },
		},
		{
			name: "JMP_abs_imm16_w", init: subs, emit: func(a *Emitter) { a.JMP_abs_imm16_w(0x9100) }, bytes: []byte{0x4C, 0x00, 0x91},
			stop: 0x009100,
		},
		{
			name: "JMP_indirect", init: func(r *cpuRun) { subs(r); r.poke16(0x000200, 0x9100) },
			emit: func(a *Emitter) { a.JMP_indirect(0x0200) }, bytes: []byte{0x6C, 0x00, 0x02},
			stop: 0x009100,
		},
		{
			name: "JMP_abs_x_indirect", init: func(r *cpuRun) { subs(r); r.setX(2); r.poke16(0x00A002, 0x9100) },
			emit: func(a *Emitter) { a.JMP_abs_x_indirect(0xA000) }, bytes: []byte{0x7C, 0x00, 0xA0},
			stop: 0x009100,
		},
		{
			name: "JML", init: subs, emit: func(a *Emitter) { a.JML(0x019000) }, bytes: []byte{0x5C, 0x00, 0x90, 0x01},
			stop: 0x019000,
		},
		{
			name: "JML_indirect", init: func(r *cpuRun) { subs(r); r.poke(0x000200, 0x00, 0x90, 0x01) },
			emit: func(a *Emitter) { a.JML_indirect(0x0200) }, bytes: []byte{0xDC, 0x00, 0x02},
			stop: 0x019000,
		},
		{
			name: "JSR_abs", init: subs, emit: func(a *Emitter) { a.JSR_abs(0x9000) }, bytes: []byte{0x20, 0x00, 0x90},
			check: func(r *cpuRun) { r.expectX(1); r.expectS(testS) },
		},
		{
			name: "JSR_abs_x_indirect", init: func(r *cpuRun) { subs(r); r.setX(2); r.poke16(0x00A002, 0x9000) },
			emit: func(a *Emitter) { a.JSR_abs_x_indirect(0xA000) }, bytes: []byte{0xFC, 0x00, 0xA0},
			check: func(r *cpuRun) { r.expectX(3); r.expectS(testS) },
		},
		{
			name: "JSL", init: subs, emit: func(a *Emitter) { a.JSL(0x018000) }, bytes: []byte{0x22, 0x00, 0x80, 0x01},
			check: func(r *cpuRun) { r.expectX(1); r.expectS(testS) },
		},
		{
			name: "JSL_lhb", init: subs, emit: func(a *Emitter) { a.JSL_lhb(0x00, 0x80, 0x01) }, bytes: []byte{0x22, 0x00, 0x80, 0x01},
			check: func(r *cpuRun) { r.expectX(1); r.expectS(testS) },
		},
		{
			name: "JSRBanked", init: func(r *cpuRun) { subs(r); r.poke(0x018000, 0xE8, 0x60) /* inx; rts */ },
			emit:  func(a *Emitter) { a.JSRBanked(0x018000, 0x018010) },
			bytes: []byte{0x4B, 0xF4, 0x0A, 0x80, 0xF4, 0x0F, 0x80, 0x5C, 0x00, 0x80, 0x01},
			check: func(r *cpuRun) { r.expectX(1); r.expectS(testS) },
		},
		{
			name: "RTS", init: func(r *cpuRun) { subs(r); r.s.CPU.SP = testS - 2; r.poke16(testS-1, 0x90FF) },
			emit: func(a *Emitter) { a.RTS() }, bytes: []byte{0x60},
			stop: 0x009100,
		},
		{
			name: "RTL", init: func(r *cpuRun) { subs(r); r.s.CPU.SP = testS - 3; r.poke(testS-2, 0xFF, 0x8F, 0x01) },
			emit: func(a *Emitter) { a.RTL() }, bytes: []byte{0x6B},
			stop: 0x019000,
		},
		{
			name: "RTI", init: func(r *cpuRun) { subs(r); r.s.CPU.SP = testS - 4; r.poke(testS-3, flagC, 0x00, 0x90, 0x01) },
			emit: func(a *Emitter) { a.RTI() }, bytes: []byte{0x40},
			stop:  0x019000,
			check: func(r *cpuRun) { r.expectP(0xFF, flagC) },
		},
		{
			name: "BRK", init: func(r *cpuRun) { subs(r); r.poke16(0x00FFE6, 0x9100) },
			emit: func(a *Emitter) { a.BRK(0x12) }, bytes: []byte{0x00, 0x12},
			stop: 0x009100,
			check: func(r *cpuRun) {
				r.expectS(testS - 4)
				r.expectMem(testS-3, 0x00, 0x02, 0x80, 0x00)
				r.expectP(flagI, flagI)
			},
		},
		{
			name: "COP", init: func(r *cpuRun) { subs(r); r.poke16(0x00FFE4, 0x9100) },
			emit: func(a *Emitter) { a.COP(0x12) }, bytes: []byte{0x02, 0x12},
			stop: 0x009100,
			check: func(r *cpuRun) {
				r.expectS(testS - 4)
				r.expectMem(testS-3, 0x00, 0x02, 0x80, 0x00)
			},
		},

		// branches outside the conditional ones:
		{
			name: "BRL", emit: func(a *Emitter) { a.BRL("t"); a.INX(); a.Label("t") }, bytes: []byte{0x82, 0x01, 0x00, 0xE8},
			check: func(r *cpuRun) { r.expectX(0) },
		},
		{
			name: "BRL_imm16", emit: func(a *Emitter) { a.BRL_imm16(1); a.INX() }, bytes: []byte{0x82, 0x01, 0x00, 0xE8},
			check: func(r *cpuRun) { r.expectX(0) },
		},
		{
			name: "BRL/backward", init: func(r *cpuRun) { r.poke(0x008000, 0xDB) },
			emit: func(a *Emitter) { a.Label("b"); a.INX(); a.BRL("b") }, bytes: []byte{0xE8, 0x82, 0xFC, 0xFF},
			stop: 0x008000,
		},
		{
			name: "BNE_far", emit: func(a *Emitter) { a.BNE_far("t"); a.INX(); a.Label("t") },
			bytes: []byte{0xF0, 0x03, 0x4C, 0x06, 0x80, 0xE8},
			check: func(r *cpuRun) { r.expectX(0) },
		},
		{
			name: "BNE_far/not_taken", p: flagZ, emit: func(a *Emitter) { a.BNE_far("t"); a.INX(); a.Label("t") },
			bytes: []byte{0xF0, 0x03, 0x4C, 0x06, 0x80, 0xE8},
			check: func(r *cpuRun) { r.expectX(1) },
		},

		// the rest:
		{name: "NOP", emit: func(a *Emitter) { a.NOP() }, bytes: []byte{0xEA}},
		{
			// WAI waits for an interrupt; an IRQ while interrupts are disabled ends the wait without being taken:
			name: "WAI", p: flagI, emit: func(a *Emitter) { a.WAI() }, bytes: []byte{0xCB},
			init: func(r *cpuRun) {
				waited := 0
				r.s.CPU.OnPC = map[uint32]func(){r.end: func() {
					if r.s.CPU.Waiting {
						if waited++; waited == 3 {
							r.s.CPU.TriggerIRQ()
						}
					}
				}}
			},
			check: func(r *cpuRun) {
				if r.s.CPU.AllCycles < 3+3 {
					r.t.Errorf("ran %d cycles; expected the WAI and at least 3 cycles waiting", r.s.CPU.AllCycles)
				}
			},
		},
		{
			name: "WDM", emit: func(a *Emitter) { a.WDM(0x5A) }, bytes: []byte{0x42, 0x5A},
			check: func(r *cpuRun) {
				if r.s.CPU.WDM != 0x5A {
					r.t.Errorf("WDM = $%02x; want $5a", r.s.CPU.WDM)
				}
			},
		},
		{
			name: "STP", emit: func(a *Emitter) { a.STP(); a.INX() }, bytes: []byte{0xDB, 0xE8},
			stop:  testCodeBase,
			check: func(r *cpuRun) { r.expectX(0) },
		},
	}

	tests = append(tests, branchTests("BRA", 0x80, 0, (*Emitter).BRA, (*Emitter).BRA_imm8)...)
	tests = append(tests, branchTests("BEQ", 0xF0, flagZ, (*Emitter).BEQ, (*Emitter).BEQ_imm8)...)
	tests = append(tests, branchTests("BNE", 0xD0, 0, (*Emitter).BNE, (*Emitter).BNE_imm8)...)
	tests = append(tests, branchTests("BCC", 0x90, 0, (*Emitter).BCC, (*Emitter).BCC_imm8)...)
	tests = append(tests, branchTests("BCS", 0xB0, flagC, (*Emitter).BCS, (*Emitter).BCS_imm8)...)
	tests = append(tests, branchTests("BPL", 0x10, 0, (*Emitter).BPL, (*Emitter).BPL_imm8)...)
	tests = append(tests, branchTests("BMI", 0x30, flagN, (*Emitter).BMI, (*Emitter).BMI_imm8)...)
	tests = append(tests, branchTests("BVC", 0x50, 0, (*Emitter).BVC, (*Emitter).BVC_imm8)...)
	tests = append(tests, branchTests("BVS", 0x70, flagV, (*Emitter).BVS, (*Emitter).BVS_imm8)...)
	// the conditional branches are not taken when their condition is false:
	for _, nt := range []struct {
		name string
		op   byte
		p    uint8
		emit func(a *Emitter, l string)
	}{
		{"BEQ", 0xF0, 0, (*Emitter).BEQ}, {"BNE", 0xD0, flagZ, (*Emitter).BNE},
		{"BCC", 0x90, flagC, (*Emitter).BCC}, {"BCS", 0xB0, 0, (*Emitter).BCS},
		{"BPL", 0x10, flagN, (*Emitter).BPL}, {"BMI", 0x30, 0, (*Emitter).BMI},
		{"BVC", 0x50, flagV, (*Emitter).BVC}, {"BVS", 0x70, 0, (*Emitter).BVS},
	} {
		nt := nt
		tests = append(tests, cpuTest{
			name: nt.name + "/not_taken", p: nt.p,
			emit:  func(a *Emitter) { nt.emit(a, "t"); a.INX(); a.Label("t") },
			bytes: []byte{nt.op, 0x01, 0xE8},
			check: func(r *cpuRun) { r.expectX(1) },
		})
	}
	return tests
}
