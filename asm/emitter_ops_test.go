package asm

import (
	"bytes"
	"testing"
)

func TestEmitter_Encodings(t *testing.T) {
	tests := []struct {
		name     string
		emit     func(a *Emitter)
		expected []byte
	}{
		{"LSR", func(a *Emitter) { a.LSR() }, []byte{0x4A}},
		{"SEC", func(a *Emitter) { a.SEC() }, []byte{0x38}},
		{"INX", func(a *Emitter) { a.INX() }, []byte{0xE8}},
		{"INY", func(a *Emitter) { a.INY() }, []byte{0xC8}},
		{"TAY", func(a *Emitter) { a.TAY() }, []byte{0xA8}},
		{"TYA", func(a *Emitter) { a.TYA() }, []byte{0x98}},
		{"ADC_imm16_w", func(a *Emitter) { a.ADC_imm16_w(0x1234) }, []byte{0x69, 0x34, 0x12}},
		{"SBC_dp", func(a *Emitter) { a.SBC_dp(0x08) }, []byte{0xE5, 0x08}},
		{"SBC_sr", func(a *Emitter) { a.SBC_sr(0x01) }, []byte{0xE3, 0x01}},
		{"CMP_dp", func(a *Emitter) { a.CMP_dp(0x0E) }, []byte{0xC5, 0x0E}},
		{"CMP_long_x", func(a *Emitter) { a.CMP_long_x(0x701234) }, []byte{0xDF, 0x34, 0x12, 0x70}},
		{"CPX_imm16_w", func(a *Emitter) { a.CPX_imm16_w(0x0020) }, []byte{0xE0, 0x20, 0x00}},
		{"EOR_imm16_w", func(a *Emitter) { a.EOR_imm16_w(0x0001) }, []byte{0x49, 0x01, 0x00}},
		{"STA_long_x", func(a *Emitter) { a.STA_long_x(0x7E2000) }, []byte{0x9F, 0x00, 0x20, 0x7E}},
		{"PEA", func(a *Emitter) { a.PEA(0xABCD) }, []byte{0xF4, 0xCD, 0xAB}},
		{"SBC_abs", func(a *Emitter) { a.SBC_abs(0x0352) }, []byte{0xED, 0x52, 0x03}},
		{"SBC_long", func(a *Emitter) { a.SBC_long(0x7E011A) }, []byte{0xEF, 0x1A, 0x01, 0x7E}},
		{"CMP_abs", func(a *Emitter) { a.CMP_abs(0x6234) }, []byte{0xCD, 0x34, 0x62}},
		{"ORA_abs", func(a *Emitter) { a.ORA_abs(0x6234) }, []byte{0x0D, 0x34, 0x62}},
		{"ORA_abs_y", func(a *Emitter) { a.ORA_abs_y(0x6234) }, []byte{0x19, 0x34, 0x62}},
		{"LDA_abs_y", func(a *Emitter) { a.LDA_abs_y(0x0800) }, []byte{0xB9, 0x00, 0x08}},
		{"STA_abs_y", func(a *Emitter) { a.STA_abs_y(0x6040) }, []byte{0x99, 0x40, 0x60}},
		{"CPY_imm16_w", func(a *Emitter) { a.CPY_imm16_w(0x0200) }, []byte{0xC0, 0x00, 0x02}},
		{"BIT_imm16_w", func(a *Emitter) { a.BIT_imm16_w(0x0100) }, []byte{0x89, 0x00, 0x01}},
		{"EOR_imm8_b", func(a *Emitter) { a.SEP(0x20); a.EOR_imm8_b(0x01) }, []byte{0xE2, 0x20, 0x49, 0x01}},
		{"TSB_abs", func(a *Emitter) { a.TSB_abs(0x62FE) }, []byte{0x0C, 0xFE, 0x62}},
		{"AND_long", func(a *Emitter) { a.AND_long(0x7062FE) }, []byte{0x2F, 0xFE, 0x62, 0x70}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewEmitter(make([]byte, 16), true)
			a.SetBase(0x708000)
			a.AssumeREP(0x30)
			tt.emit(a)
			if err := a.Finalize(); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(a.Bytes(), tt.expected) {
				t.Errorf("got % x; expected % x", a.Bytes(), tt.expected)
			}
		})
	}
}

func TestEmitter_BNE_far(t *testing.T) {
	a := NewEmitter(make([]byte, 0x200), true)
	a.SetBase(0x708000)
	a.BNE_far("far")
	for i := 0; i < 0x100; i++ {
		a.NOP()
	}
	a.Label("far")
	if err := a.Finalize(); err != nil {
		t.Fatal(err)
	}
	// BEQ +3 skips over JMP far:
	if expected := []byte{0xF0, 0x03, 0x4C, 0x05, 0x81}; !bytes.Equal(a.Bytes()[:5], expected) {
		t.Errorf("got % x; expected % x", a.Bytes()[:5], expected)
	}
}

func TestEmitter_JSRBanked(t *testing.T) {
	a := NewEmitter(make([]byte, 0x20), true)
	a.SetBase(0x708000)
	a.JSRBanked(0x01D828, 0x01D827)
	ret := a.PC()
	if err := a.Finalize(); err != nil {
		t.Fatal(err)
	}
	expected := []byte{
		0x4B,                                      // phk
		0xF4, byte(ret - 1), byte((ret - 1) >> 8), // pea ret-1
		0xF4, 0x26, 0xD8, // pea rtl-1
		0x5C, 0x28, 0xD8, 0x01, // jml routine
	}
	if !bytes.Equal(a.Bytes(), expected) {
		t.Errorf("got % x; expected % x", a.Bytes(), expected)
	}

	defer func() {
		if recover() == nil {
			t.Error("expected panic for routine and rtl in different banks")
		}
	}()
	a.JSRBanked(0x01D828, 0x02D827)
}
