package asm

import "fmt"

func (a *Emitter) ORA_dp_x_indirect(addr uint8) {
	var d [2]byte
	d[0] = 0x01
	d[1] = addr
	a.emit2("ora.b", "($%02[1]x,X)", d)
}

func (a *Emitter) ORA_sr(offs uint8) {
	var d [2]byte
	d[0] = 0x03
	d[1] = offs
	a.emit2("ora", "$%02[1]x,S", d)
}

func (a *Emitter) ORA_dp(addr uint8) {
	var d [2]byte
	d[0] = 0x05
	d[1] = addr
	a.emit2("ora.b", "$%02[1]x", d)
}

func (a *Emitter) ORA_dp_indirect_long(addr uint8) {
	var d [2]byte
	d[0] = 0x07
	d[1] = addr
	a.emit2("ora.b", "[$%02[1]x]", d)
}

func (a *Emitter) ORA_dp_indirect_y(addr uint8) {
	var d [2]byte
	d[0] = 0x11
	d[1] = addr
	a.emit2("ora.b", "($%02[1]x),Y", d)
}

func (a *Emitter) ORA_dp_indirect(addr uint8) {
	var d [2]byte
	d[0] = 0x12
	d[1] = addr
	a.emit2("ora.b", "($%02[1]x)", d)
}

func (a *Emitter) ORA_sr_indirect_y(offs uint8) {
	var d [2]byte
	d[0] = 0x13
	d[1] = offs
	a.emit2("ora", "($%02[1]x,S),Y", d)
}

func (a *Emitter) ORA_dp_x(addr uint8) {
	var d [2]byte
	d[0] = 0x15
	d[1] = addr
	a.emit2("ora.b", "$%02[1]x,X", d)
}

func (a *Emitter) ORA_dp_indirect_long_y(addr uint8) {
	var d [2]byte
	d[0] = 0x17
	d[1] = addr
	a.emit2("ora.b", "[$%02[1]x],Y", d)
}

func (a *Emitter) ORA_abs_x(addr uint16) {
	var d [3]byte
	d[0] = 0x1D
	d[1], d[2] = imm16(addr)
	a.emit3("ora.w", "$%02[2]x%02[1]x,X", d)
}

func (a *Emitter) ORA_long_x(addr uint32) {
	var d [4]byte
	d[0] = 0x1F
	d[1], d[2], d[3] = imm24(addr)
	a.emit4("ora.l", "$%02[3]x%02[2]x%02[1]x,X", d)
}

func (a *Emitter) AND_dp_x_indirect(addr uint8) {
	var d [2]byte
	d[0] = 0x21
	d[1] = addr
	a.emit2("and.b", "($%02[1]x,X)", d)
}

func (a *Emitter) AND_sr(offs uint8) {
	var d [2]byte
	d[0] = 0x23
	d[1] = offs
	a.emit2("and", "$%02[1]x,S", d)
}

func (a *Emitter) AND_dp(addr uint8) {
	var d [2]byte
	d[0] = 0x25
	d[1] = addr
	a.emit2("and.b", "$%02[1]x", d)
}

func (a *Emitter) AND_dp_indirect_long(addr uint8) {
	var d [2]byte
	d[0] = 0x27
	d[1] = addr
	a.emit2("and.b", "[$%02[1]x]", d)
}

func (a *Emitter) AND_dp_indirect_y(addr uint8) {
	var d [2]byte
	d[0] = 0x31
	d[1] = addr
	a.emit2("and.b", "($%02[1]x),Y", d)
}

func (a *Emitter) AND_dp_indirect(addr uint8) {
	var d [2]byte
	d[0] = 0x32
	d[1] = addr
	a.emit2("and.b", "($%02[1]x)", d)
}

func (a *Emitter) AND_sr_indirect_y(offs uint8) {
	var d [2]byte
	d[0] = 0x33
	d[1] = offs
	a.emit2("and", "($%02[1]x,S),Y", d)
}

func (a *Emitter) AND_dp_x(addr uint8) {
	var d [2]byte
	d[0] = 0x35
	d[1] = addr
	a.emit2("and.b", "$%02[1]x,X", d)
}

func (a *Emitter) AND_dp_indirect_long_y(addr uint8) {
	var d [2]byte
	d[0] = 0x37
	d[1] = addr
	a.emit2("and.b", "[$%02[1]x],Y", d)
}

func (a *Emitter) AND_abs(addr uint16) {
	var d [3]byte
	d[0] = 0x2D
	d[1], d[2] = imm16(addr)
	a.emit3("and.w", "$%02[2]x%02[1]x", d)
}

func (a *Emitter) AND_abs_y(addr uint16) {
	var d [3]byte
	d[0] = 0x39
	d[1], d[2] = imm16(addr)
	a.emit3("and.w", "$%02[2]x%02[1]x,Y", d)
}

func (a *Emitter) AND_abs_x(addr uint16) {
	var d [3]byte
	d[0] = 0x3D
	d[1], d[2] = imm16(addr)
	a.emit3("and.w", "$%02[2]x%02[1]x,X", d)
}

func (a *Emitter) AND_long_x(addr uint32) {
	var d [4]byte
	d[0] = 0x3F
	d[1], d[2], d[3] = imm24(addr)
	a.emit4("and.l", "$%02[3]x%02[2]x%02[1]x,X", d)
}

func (a *Emitter) EOR_dp_x_indirect(addr uint8) {
	var d [2]byte
	d[0] = 0x41
	d[1] = addr
	a.emit2("eor.b", "($%02[1]x,X)", d)
}

func (a *Emitter) EOR_sr(offs uint8) {
	var d [2]byte
	d[0] = 0x43
	d[1] = offs
	a.emit2("eor", "$%02[1]x,S", d)
}

func (a *Emitter) EOR_dp(addr uint8) {
	var d [2]byte
	d[0] = 0x45
	d[1] = addr
	a.emit2("eor.b", "$%02[1]x", d)
}

func (a *Emitter) EOR_dp_indirect_long(addr uint8) {
	var d [2]byte
	d[0] = 0x47
	d[1] = addr
	a.emit2("eor.b", "[$%02[1]x]", d)
}

func (a *Emitter) EOR_dp_indirect_y(addr uint8) {
	var d [2]byte
	d[0] = 0x51
	d[1] = addr
	a.emit2("eor.b", "($%02[1]x),Y", d)
}

func (a *Emitter) EOR_dp_indirect(addr uint8) {
	var d [2]byte
	d[0] = 0x52
	d[1] = addr
	a.emit2("eor.b", "($%02[1]x)", d)
}

func (a *Emitter) EOR_sr_indirect_y(offs uint8) {
	var d [2]byte
	d[0] = 0x53
	d[1] = offs
	a.emit2("eor", "($%02[1]x,S),Y", d)
}

func (a *Emitter) EOR_dp_x(addr uint8) {
	var d [2]byte
	d[0] = 0x55
	d[1] = addr
	a.emit2("eor.b", "$%02[1]x,X", d)
}

func (a *Emitter) EOR_dp_indirect_long_y(addr uint8) {
	var d [2]byte
	d[0] = 0x57
	d[1] = addr
	a.emit2("eor.b", "[$%02[1]x],Y", d)
}

func (a *Emitter) EOR_abs(addr uint16) {
	var d [3]byte
	d[0] = 0x4D
	d[1], d[2] = imm16(addr)
	a.emit3("eor.w", "$%02[2]x%02[1]x", d)
}

func (a *Emitter) EOR_abs_y(addr uint16) {
	var d [3]byte
	d[0] = 0x59
	d[1], d[2] = imm16(addr)
	a.emit3("eor.w", "$%02[2]x%02[1]x,Y", d)
}

func (a *Emitter) EOR_abs_x(addr uint16) {
	var d [3]byte
	d[0] = 0x5D
	d[1], d[2] = imm16(addr)
	a.emit3("eor.w", "$%02[2]x%02[1]x,X", d)
}

func (a *Emitter) EOR_long(addr uint32) {
	var d [4]byte
	d[0] = 0x4F
	d[1], d[2], d[3] = imm24(addr)
	a.emit4("eor.l", "$%02[3]x%02[2]x%02[1]x", d)
}

func (a *Emitter) EOR_long_x(addr uint32) {
	var d [4]byte
	d[0] = 0x5F
	d[1], d[2], d[3] = imm24(addr)
	a.emit4("eor.l", "$%02[3]x%02[2]x%02[1]x,X", d)
}

func (a *Emitter) ADC_dp_x_indirect(addr uint8) {
	var d [2]byte
	d[0] = 0x61
	d[1] = addr
	a.emit2("adc.b", "($%02[1]x,X)", d)
}

func (a *Emitter) ADC_sr(offs uint8) {
	var d [2]byte
	d[0] = 0x63
	d[1] = offs
	a.emit2("adc", "$%02[1]x,S", d)
}

func (a *Emitter) ADC_dp(addr uint8) {
	var d [2]byte
	d[0] = 0x65
	d[1] = addr
	a.emit2("adc.b", "$%02[1]x", d)
}

func (a *Emitter) ADC_dp_indirect_long(addr uint8) {
	var d [2]byte
	d[0] = 0x67
	d[1] = addr
	a.emit2("adc.b", "[$%02[1]x]", d)
}

func (a *Emitter) ADC_dp_indirect_y(addr uint8) {
	var d [2]byte
	d[0] = 0x71
	d[1] = addr
	a.emit2("adc.b", "($%02[1]x),Y", d)
}

func (a *Emitter) ADC_dp_indirect(addr uint8) {
	var d [2]byte
	d[0] = 0x72
	d[1] = addr
	a.emit2("adc.b", "($%02[1]x)", d)
}

func (a *Emitter) ADC_sr_indirect_y(offs uint8) {
	var d [2]byte
	d[0] = 0x73
	d[1] = offs
	a.emit2("adc", "($%02[1]x,S),Y", d)
}

func (a *Emitter) ADC_dp_x(addr uint8) {
	var d [2]byte
	d[0] = 0x75
	d[1] = addr
	a.emit2("adc.b", "$%02[1]x,X", d)
}

func (a *Emitter) ADC_dp_indirect_long_y(addr uint8) {
	var d [2]byte
	d[0] = 0x77
	d[1] = addr
	a.emit2("adc.b", "[$%02[1]x],Y", d)
}

func (a *Emitter) ADC_abs(addr uint16) {
	var d [3]byte
	d[0] = 0x6D
	d[1], d[2] = imm16(addr)
	a.emit3("adc.w", "$%02[2]x%02[1]x", d)
}

func (a *Emitter) ADC_abs_y(addr uint16) {
	var d [3]byte
	d[0] = 0x79
	d[1], d[2] = imm16(addr)
	a.emit3("adc.w", "$%02[2]x%02[1]x,Y", d)
}

func (a *Emitter) ADC_abs_x(addr uint16) {
	var d [3]byte
	d[0] = 0x7D
	d[1], d[2] = imm16(addr)
	a.emit3("adc.w", "$%02[2]x%02[1]x,X", d)
}

func (a *Emitter) ADC_long(addr uint32) {
	var d [4]byte
	d[0] = 0x6F
	d[1], d[2], d[3] = imm24(addr)
	a.emit4("adc.l", "$%02[3]x%02[2]x%02[1]x", d)
}

func (a *Emitter) ADC_long_x(addr uint32) {
	var d [4]byte
	d[0] = 0x7F
	d[1], d[2], d[3] = imm24(addr)
	a.emit4("adc.l", "$%02[3]x%02[2]x%02[1]x,X", d)
}

func (a *Emitter) STA_dp_x_indirect(addr uint8) {
	var d [2]byte
	d[0] = 0x81
	d[1] = addr
	a.emit2("sta.b", "($%02[1]x,X)", d)
}

func (a *Emitter) STA_sr(offs uint8) {
	var d [2]byte
	d[0] = 0x83
	d[1] = offs
	a.emit2("sta", "$%02[1]x,S", d)
}

func (a *Emitter) STA_dp_indirect_long(addr uint8) {
	var d [2]byte
	d[0] = 0x87
	d[1] = addr
	a.emit2("sta.b", "[$%02[1]x]", d)
}

func (a *Emitter) STA_dp_indirect_y(addr uint8) {
	var d [2]byte
	d[0] = 0x91
	d[1] = addr
	a.emit2("sta.b", "($%02[1]x),Y", d)
}

func (a *Emitter) STA_dp_indirect(addr uint8) {
	var d [2]byte
	d[0] = 0x92
	d[1] = addr
	a.emit2("sta.b", "($%02[1]x)", d)
}

func (a *Emitter) STA_sr_indirect_y(offs uint8) {
	var d [2]byte
	d[0] = 0x93
	d[1] = offs
	a.emit2("sta", "($%02[1]x,S),Y", d)
}

func (a *Emitter) STA_dp_x(addr uint8) {
	var d [2]byte
	d[0] = 0x95
	d[1] = addr
	a.emit2("sta.b", "$%02[1]x,X", d)
}

func (a *Emitter) STA_dp_indirect_long_y(addr uint8) {
	var d [2]byte
	d[0] = 0x97
	d[1] = addr
	a.emit2("sta.b", "[$%02[1]x],Y", d)
}

func (a *Emitter) LDA_dp_x_indirect(addr uint8) {
	var d [2]byte
	d[0] = 0xA1
	d[1] = addr
	a.emit2("lda.b", "($%02[1]x,X)", d)
}

func (a *Emitter) LDA_sr(offs uint8) {
	var d [2]byte
	d[0] = 0xA3
	d[1] = offs
	a.emit2("lda", "$%02[1]x,S", d)
}

func (a *Emitter) LDA_dp_indirect_long(addr uint8) {
	var d [2]byte
	d[0] = 0xA7
	d[1] = addr
	a.emit2("lda.b", "[$%02[1]x]", d)
}

func (a *Emitter) LDA_dp_indirect_y(addr uint8) {
	var d [2]byte
	d[0] = 0xB1
	d[1] = addr
	a.emit2("lda.b", "($%02[1]x),Y", d)
}

func (a *Emitter) LDA_dp_indirect(addr uint8) {
	var d [2]byte
	d[0] = 0xB2
	d[1] = addr
	a.emit2("lda.b", "($%02[1]x)", d)
}

func (a *Emitter) LDA_sr_indirect_y(offs uint8) {
	var d [2]byte
	d[0] = 0xB3
	d[1] = offs
	a.emit2("lda", "($%02[1]x,S),Y", d)
}

func (a *Emitter) LDA_dp_x(addr uint8) {
	var d [2]byte
	d[0] = 0xB5
	d[1] = addr
	a.emit2("lda.b", "$%02[1]x,X", d)
}

func (a *Emitter) LDA_dp_indirect_long_y(addr uint8) {
	var d [2]byte
	d[0] = 0xB7
	d[1] = addr
	a.emit2("lda.b", "[$%02[1]x],Y", d)
}

func (a *Emitter) CMP_dp_x_indirect(addr uint8) {
	var d [2]byte
	d[0] = 0xC1
	d[1] = addr
	a.emit2("cmp.b", "($%02[1]x,X)", d)
}

func (a *Emitter) CMP_sr(offs uint8) {
	var d [2]byte
	d[0] = 0xC3
	d[1] = offs
	a.emit2("cmp", "$%02[1]x,S", d)
}

func (a *Emitter) CMP_dp_indirect_long(addr uint8) {
	var d [2]byte
	d[0] = 0xC7
	d[1] = addr
	a.emit2("cmp.b", "[$%02[1]x]", d)
}

func (a *Emitter) CMP_dp_indirect_y(addr uint8) {
	var d [2]byte
	d[0] = 0xD1
	d[1] = addr
	a.emit2("cmp.b", "($%02[1]x),Y", d)
}

func (a *Emitter) CMP_dp_indirect(addr uint8) {
	var d [2]byte
	d[0] = 0xD2
	d[1] = addr
	a.emit2("cmp.b", "($%02[1]x)", d)
}

func (a *Emitter) CMP_sr_indirect_y(offs uint8) {
	var d [2]byte
	d[0] = 0xD3
	d[1] = offs
	a.emit2("cmp", "($%02[1]x,S),Y", d)
}

func (a *Emitter) CMP_dp_x(addr uint8) {
	var d [2]byte
	d[0] = 0xD5
	d[1] = addr
	a.emit2("cmp.b", "$%02[1]x,X", d)
}

func (a *Emitter) CMP_dp_indirect_long_y(addr uint8) {
	var d [2]byte
	d[0] = 0xD7
	d[1] = addr
	a.emit2("cmp.b", "[$%02[1]x],Y", d)
}

func (a *Emitter) CMP_abs_y(addr uint16) {
	var d [3]byte
	d[0] = 0xD9
	d[1], d[2] = imm16(addr)
	a.emit3("cmp.w", "$%02[2]x%02[1]x,Y", d)
}

func (a *Emitter) CMP_abs_x(addr uint16) {
	var d [3]byte
	d[0] = 0xDD
	d[1], d[2] = imm16(addr)
	a.emit3("cmp.w", "$%02[2]x%02[1]x,X", d)
}

func (a *Emitter) SBC_dp_x_indirect(addr uint8) {
	var d [2]byte
	d[0] = 0xE1
	d[1] = addr
	a.emit2("sbc.b", "($%02[1]x,X)", d)
}

func (a *Emitter) SBC_dp_indirect_long(addr uint8) {
	var d [2]byte
	d[0] = 0xE7
	d[1] = addr
	a.emit2("sbc.b", "[$%02[1]x]", d)
}

func (a *Emitter) SBC_dp_indirect_y(addr uint8) {
	var d [2]byte
	d[0] = 0xF1
	d[1] = addr
	a.emit2("sbc.b", "($%02[1]x),Y", d)
}

func (a *Emitter) SBC_dp_indirect(addr uint8) {
	var d [2]byte
	d[0] = 0xF2
	d[1] = addr
	a.emit2("sbc.b", "($%02[1]x)", d)
}

func (a *Emitter) SBC_sr_indirect_y(offs uint8) {
	var d [2]byte
	d[0] = 0xF3
	d[1] = offs
	a.emit2("sbc", "($%02[1]x,S),Y", d)
}

func (a *Emitter) SBC_dp_x(addr uint8) {
	var d [2]byte
	d[0] = 0xF5
	d[1] = addr
	a.emit2("sbc.b", "$%02[1]x,X", d)
}

func (a *Emitter) SBC_dp_indirect_long_y(addr uint8) {
	var d [2]byte
	d[0] = 0xF7
	d[1] = addr
	a.emit2("sbc.b", "[$%02[1]x],Y", d)
}

func (a *Emitter) SBC_abs_y(addr uint16) {
	var d [3]byte
	d[0] = 0xF9
	d[1], d[2] = imm16(addr)
	a.emit3("sbc.w", "$%02[2]x%02[1]x,Y", d)
}

func (a *Emitter) SBC_abs_x(addr uint16) {
	var d [3]byte
	d[0] = 0xFD
	d[1], d[2] = imm16(addr)
	a.emit3("sbc.w", "$%02[2]x%02[1]x,X", d)
}

func (a *Emitter) SBC_long_x(addr uint32) {
	var d [4]byte
	d[0] = 0xFF
	d[1], d[2], d[3] = imm24(addr)
	a.emit4("sbc.l", "$%02[3]x%02[2]x%02[1]x,X", d)
}

func (a *Emitter) SBC_imm16_w(m uint16) {
	if !a.IsM16bit() {
		panic(fmt.Errorf("asm: SBC_imm16_w called but 'm' flag is 8-bit; call REP(0x20) or AssumeREP(0x20) first"))
	}
	var d [3]byte
	d[0] = 0xE9
	d[1], d[2] = imm16(m)
	a.emit3("sbc.w", "#$%02[2]x%02[1]x", d)
}

func (a *Emitter) ASL_dp(addr uint8) {
	var d [2]byte
	d[0] = 0x06
	d[1] = addr
	a.emit2("asl.b", "$%02[1]x", d)
}

func (a *Emitter) ASL_abs(addr uint16) {
	var d [3]byte
	d[0] = 0x0E
	d[1], d[2] = imm16(addr)
	a.emit3("asl.w", "$%02[2]x%02[1]x", d)
}

func (a *Emitter) ASL_dp_x(addr uint8) {
	var d [2]byte
	d[0] = 0x16
	d[1] = addr
	a.emit2("asl.b", "$%02[1]x,X", d)
}

func (a *Emitter) ASL_abs_x(addr uint16) {
	var d [3]byte
	d[0] = 0x1E
	d[1], d[2] = imm16(addr)
	a.emit3("asl.w", "$%02[2]x%02[1]x,X", d)
}

func (a *Emitter) ROL() {
	a.emit1("rol", [1]byte{0x2A})
}

func (a *Emitter) ROL_dp(addr uint8) {
	var d [2]byte
	d[0] = 0x26
	d[1] = addr
	a.emit2("rol.b", "$%02[1]x", d)
}

func (a *Emitter) ROL_abs(addr uint16) {
	var d [3]byte
	d[0] = 0x2E
	d[1], d[2] = imm16(addr)
	a.emit3("rol.w", "$%02[2]x%02[1]x", d)
}

func (a *Emitter) ROL_dp_x(addr uint8) {
	var d [2]byte
	d[0] = 0x36
	d[1] = addr
	a.emit2("rol.b", "$%02[1]x,X", d)
}

func (a *Emitter) ROL_abs_x(addr uint16) {
	var d [3]byte
	d[0] = 0x3E
	d[1], d[2] = imm16(addr)
	a.emit3("rol.w", "$%02[2]x%02[1]x,X", d)
}

func (a *Emitter) LSR_dp(addr uint8) {
	var d [2]byte
	d[0] = 0x46
	d[1] = addr
	a.emit2("lsr.b", "$%02[1]x", d)
}

func (a *Emitter) LSR_abs(addr uint16) {
	var d [3]byte
	d[0] = 0x4E
	d[1], d[2] = imm16(addr)
	a.emit3("lsr.w", "$%02[2]x%02[1]x", d)
}

func (a *Emitter) LSR_dp_x(addr uint8) {
	var d [2]byte
	d[0] = 0x56
	d[1] = addr
	a.emit2("lsr.b", "$%02[1]x,X", d)
}

func (a *Emitter) LSR_abs_x(addr uint16) {
	var d [3]byte
	d[0] = 0x5E
	d[1], d[2] = imm16(addr)
	a.emit3("lsr.w", "$%02[2]x%02[1]x,X", d)
}

func (a *Emitter) ROR() {
	a.emit1("ror", [1]byte{0x6A})
}

func (a *Emitter) ROR_dp(addr uint8) {
	var d [2]byte
	d[0] = 0x66
	d[1] = addr
	a.emit2("ror.b", "$%02[1]x", d)
}

func (a *Emitter) ROR_abs(addr uint16) {
	var d [3]byte
	d[0] = 0x6E
	d[1], d[2] = imm16(addr)
	a.emit3("ror.w", "$%02[2]x%02[1]x", d)
}

func (a *Emitter) ROR_dp_x(addr uint8) {
	var d [2]byte
	d[0] = 0x76
	d[1] = addr
	a.emit2("ror.b", "$%02[1]x,X", d)
}

func (a *Emitter) ROR_abs_x(addr uint16) {
	var d [3]byte
	d[0] = 0x7E
	d[1], d[2] = imm16(addr)
	a.emit3("ror.w", "$%02[2]x%02[1]x,X", d)
}

func (a *Emitter) INC() {
	a.emit1("inc", [1]byte{0x1A})
}

func (a *Emitter) INC_dp_x(addr uint8) {
	var d [2]byte
	d[0] = 0xF6
	d[1] = addr
	a.emit2("inc.b", "$%02[1]x,X", d)
}

func (a *Emitter) INC_abs_x(addr uint16) {
	var d [3]byte
	d[0] = 0xFE
	d[1], d[2] = imm16(addr)
	a.emit3("inc.w", "$%02[2]x%02[1]x,X", d)
}

func (a *Emitter) DEC() {
	a.emit1("dec", [1]byte{0x3A})
}

func (a *Emitter) DEC_dp_x(addr uint8) {
	var d [2]byte
	d[0] = 0xD6
	d[1] = addr
	a.emit2("dec.b", "$%02[1]x,X", d)
}

func (a *Emitter) DEC_abs_x(addr uint16) {
	var d [3]byte
	d[0] = 0xDE
	d[1], d[2] = imm16(addr)
	a.emit3("dec.w", "$%02[2]x%02[1]x,X", d)
}

func (a *Emitter) BIT_dp(addr uint8) {
	var d [2]byte
	d[0] = 0x24
	d[1] = addr
	a.emit2("bit.b", "$%02[1]x", d)
}

func (a *Emitter) BIT_abs(addr uint16) {
	var d [3]byte
	d[0] = 0x2C
	d[1], d[2] = imm16(addr)
	a.emit3("bit.w", "$%02[2]x%02[1]x", d)
}

func (a *Emitter) BIT_dp_x(addr uint8) {
	var d [2]byte
	d[0] = 0x34
	d[1] = addr
	a.emit2("bit.b", "$%02[1]x,X", d)
}

func (a *Emitter) BIT_abs_x(addr uint16) {
	var d [3]byte
	d[0] = 0x3C
	d[1], d[2] = imm16(addr)
	a.emit3("bit.w", "$%02[2]x%02[1]x,X", d)
}

func (a *Emitter) BIT_imm8_b(m uint8) {
	if a.IsM16bit() {
		panic(fmt.Errorf("asm: BIT_imm8_b called but 'm' flag is 16-bit; call SEP(0x20) or AssumeSEP(0x20) first"))
	}
	var d [2]byte
	d[0] = 0x89
	d[1] = m
	a.emit2("bit.b", "#$%02x", d)
}

func (a *Emitter) TSB_dp(addr uint8) {
	var d [2]byte
	d[0] = 0x04
	d[1] = addr
	a.emit2("tsb.b", "$%02[1]x", d)
}

func (a *Emitter) TRB_dp(addr uint8) {
	var d [2]byte
	d[0] = 0x14
	d[1] = addr
	a.emit2("trb.b", "$%02[1]x", d)
}

func (a *Emitter) TRB_abs(addr uint16) {
	var d [3]byte
	d[0] = 0x1C
	d[1], d[2] = imm16(addr)
	a.emit3("trb.w", "$%02[2]x%02[1]x", d)
}

func (a *Emitter) STZ_dp_x(addr uint8) {
	var d [2]byte
	d[0] = 0x74
	d[1] = addr
	a.emit2("stz.b", "$%02[1]x,X", d)
}

func (a *Emitter) STX_dp(addr uint8) {
	var d [2]byte
	d[0] = 0x86
	d[1] = addr
	a.emit2("stx.b", "$%02[1]x", d)
}

func (a *Emitter) STX_dp_y(addr uint8) {
	var d [2]byte
	d[0] = 0x96
	d[1] = addr
	a.emit2("stx.b", "$%02[1]x,Y", d)
}

func (a *Emitter) LDX_dp(addr uint8) {
	var d [2]byte
	d[0] = 0xA6
	d[1] = addr
	a.emit2("ldx.b", "$%02[1]x", d)
}

func (a *Emitter) LDX_dp_y(addr uint8) {
	var d [2]byte
	d[0] = 0xB6
	d[1] = addr
	a.emit2("ldx.b", "$%02[1]x,Y", d)
}

func (a *Emitter) LDX_abs_y(addr uint16) {
	var d [3]byte
	d[0] = 0xBE
	d[1], d[2] = imm16(addr)
	a.emit3("ldx.w", "$%02[2]x%02[1]x,Y", d)
}

func (a *Emitter) LDY_dp(addr uint8) {
	var d [2]byte
	d[0] = 0xA4
	d[1] = addr
	a.emit2("ldy.b", "$%02[1]x", d)
}

func (a *Emitter) LDY_dp_x(addr uint8) {
	var d [2]byte
	d[0] = 0xB4
	d[1] = addr
	a.emit2("ldy.b", "$%02[1]x,X", d)
}

func (a *Emitter) LDY_abs_x(addr uint16) {
	var d [3]byte
	d[0] = 0xBC
	d[1], d[2] = imm16(addr)
	a.emit3("ldy.w", "$%02[2]x%02[1]x,X", d)
}

func (a *Emitter) CPX_dp(addr uint8) {
	var d [2]byte
	d[0] = 0xE4
	d[1] = addr
	a.emit2("cpx.b", "$%02[1]x", d)
}

func (a *Emitter) CPX_abs(addr uint16) {
	var d [3]byte
	d[0] = 0xEC
	d[1], d[2] = imm16(addr)
	a.emit3("cpx.w", "$%02[2]x%02[1]x", d)
}

func (a *Emitter) CPY_dp(addr uint8) {
	var d [2]byte
	d[0] = 0xC4
	d[1] = addr
	a.emit2("cpy.b", "$%02[1]x", d)
}

func (a *Emitter) CPY_abs(addr uint16) {
	var d [3]byte
	d[0] = 0xCC
	d[1], d[2] = imm16(addr)
	a.emit3("cpy.w", "$%02[2]x%02[1]x", d)
}

func (a *Emitter) CPX_imm8_b(m uint8) {
	if a.IsX16bit() {
		panic(fmt.Errorf("asm: CPX_imm8_b called but 'x' flag is 16-bit; call SEP(0x10) or AssumeSEP(0x10) first"))
	}
	var d [2]byte
	d[0] = 0xE0
	d[1] = m
	a.emit2("cpx.b", "#$%02x", d)
}

func (a *Emitter) CLV() {
	a.emit1("clv", [1]byte{0xB8})
}

func (a *Emitter) CLD() {
	a.emit1("cld", [1]byte{0xD8})
}

func (a *Emitter) SED() {
	a.emit1("sed", [1]byte{0xF8})
}

func (a *Emitter) XCE() {
	a.emit1("xce", [1]byte{0xFB})
}

func (a *Emitter) WAI() {
	a.emit1("wai", [1]byte{0xCB})
}

func (a *Emitter) TCS() {
	a.emit1("tcs", [1]byte{0x1B})
}

func (a *Emitter) TSC() {
	a.emit1("tsc", [1]byte{0x3B})
}

func (a *Emitter) TDC() {
	a.emit1("tdc", [1]byte{0x7B})
}

func (a *Emitter) TSX() {
	a.emit1("tsx", [1]byte{0xBA})
}

func (a *Emitter) TXS() {
	a.emit1("txs", [1]byte{0x9A})
}

func (a *Emitter) TXY() {
	a.emit1("txy", [1]byte{0x9B})
}

func (a *Emitter) TYX() {
	a.emit1("tyx", [1]byte{0xBB})
}

func (a *Emitter) BMI_imm8(m int8) {
	var d [2]byte
	d[0] = 0x30
	d[1] = uint8(m)
	a.emit2("bmi", "$%02x", d)
}

func (a *Emitter) BCC_imm8(m int8) {
	var d [2]byte
	d[0] = 0x90
	d[1] = uint8(m)
	a.emit2("bcc", "$%02x", d)
}

func (a *Emitter) BCS_imm8(m int8) {
	var d [2]byte
	d[0] = 0xB0
	d[1] = uint8(m)
	a.emit2("bcs", "$%02x", d)
}

func (a *Emitter) BVC_imm8(m int8) {
	var d [2]byte
	d[0] = 0x50
	d[1] = uint8(m)
	a.emit2("bvc", "$%02x", d)
}

func (a *Emitter) BVC(label string) {
	var d [2]byte
	d[0] = 0x50
	d[1] = 0xFF // will be overwritten by Finalize()
	a.emit2Label("bvc", label, d)
}

func (a *Emitter) BVS_imm8(m int8) {
	var d [2]byte
	d[0] = 0x70
	d[1] = uint8(m)
	a.emit2("bvs", "$%02x", d)
}

func (a *Emitter) BVS(label string) {
	var d [2]byte
	d[0] = 0x70
	d[1] = 0xFF // will be overwritten by Finalize()
	a.emit2Label("bvs", label, d)
}

// BRL_imm16 branches by m bytes relative to the address after the instruction, within the program bank.
func (a *Emitter) BRL_imm16(m int16) {
	var d [3]byte
	d[0] = 0x82
	d[1], d[2] = imm16(uint16(m))
	a.emit3("brl", "$%02[2]x%02[1]x", d)
}

// BRL branches to label, which must be in the same bank.
func (a *Emitter) BRL(label string) {
	var d [3]byte
	d[0] = 0x82
	d[1] = 0xFF // will be overwritten by Finalize()
	d[2] = 0xFF // will be overwritten by Finalize()
	a.emit3LabelS16("brl", label, "%s", d)
}

// PER_imm16 pushes the address after the instruction plus m.
func (a *Emitter) PER_imm16(m int16) {
	var d [3]byte
	d[0] = 0x62
	d[1], d[2] = imm16(uint16(m))
	a.emit3("per", "$%02[2]x%02[1]x", d)
}

// PER pushes the 16-bit address of label, which must be in the same bank.
func (a *Emitter) PER(label string) {
	var d [3]byte
	d[0] = 0x62
	d[1] = 0xFF // will be overwritten by Finalize()
	d[2] = 0xFF // will be overwritten by Finalize()
	a.emit3LabelS16("per", label, "%s", d)
}

// PEI pushes the 16-bit value at the direct page address.
func (a *Emitter) PEI(addr uint8) {
	var d [2]byte
	d[0] = 0xD4
	d[1] = addr
	a.emit2("pei", "($%02[1]x)", d)
}

// JMP_abs_x_indirect jumps to the address read from the program bank at addr+X.
func (a *Emitter) JMP_abs_x_indirect(addr uint16) {
	var d [3]byte
	d[0] = 0x7C
	d[1], d[2] = imm16(addr)
	a.emit3("jmp", "($%02[2]x%02[1]x,X)", d)
}

// JML_indirect jumps to the 24-bit address read from bank 0 at addr.
func (a *Emitter) JML_indirect(addr uint16) {
	var d [3]byte
	d[0] = 0xDC
	d[1], d[2] = imm16(addr)
	a.emit3("jml", "[$%02[2]x%02[1]x]", d)
}

// JSR_abs_x_indirect calls the address read from the program bank at addr+X.
func (a *Emitter) JSR_abs_x_indirect(addr uint16) {
	var d [3]byte
	d[0] = 0xFC
	d[1], d[2] = imm16(addr)
	a.emit3("jsr", "($%02[2]x%02[1]x,X)", d)
}

// BRK triggers a software interrupt through the BRK vector; m is the signature byte.
func (a *Emitter) BRK(m uint8) {
	a.emit2("brk", "#$%02x", [2]byte{0x00, m})
}

// COP triggers a software interrupt through the COP vector; m is the signature byte.
func (a *Emitter) COP(m uint8) {
	a.emit2("cop", "#$%02x", [2]byte{0x02, m})
}

func (a *Emitter) MVP(destBank uint8, srcBank uint8) {
	var d [3]byte
	d[0] = 0x44
	d[1], d[2] = destBank, srcBank
	a.emit3("mvp", "$%02[1]x,$%02[2]x", d)
}
