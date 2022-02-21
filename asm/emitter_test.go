package asm

import (
	"fmt"
	"log"
	"testing"
)

func TestEmitter_LabelBackwards(t *testing.T) {
	a := NewEmitter(make([]byte, 0x100), true)
	a.Label("loop")
	a.BNE("loop")
	if err := a.Finalize(); err != nil {
		t.Error(err)
	}
	if err := a.WriteTextTo(log.Writer()); err != nil {
		t.Error(err)
	}
}

func TestEmitter_LabelForwards_InRange(t *testing.T) {
	a := NewEmitter(make([]byte, 0x100), true)
	a.SEP(0x30)
	a.LDA_imm8_b(0x01)
	a.BNE("next")
	a.RTS()
	a.Label("next")
	a.CMP_imm8_b(0x02)
	a.RTS()
	if err := a.Finalize(); err != nil {
		t.Error(err)
	}
	if err := a.WriteTextTo(log.Writer()); err != nil {
		t.Error(err)
	}
}

func TestEmitter_LabelForwards_NoRef(t *testing.T) {
	a := NewEmitter(make([]byte, 0x100), true)
	a.SEP(0x30)
	a.LDA_imm8_b(0x01)
	a.BNE("next")
	a.RTS()
	a.Label("next")
	a.CMP_imm8_b(0x02)
	a.BNE("next2")
	a.RTS()
	err := a.Finalize()
	expectedErrStr := "could not resolve label 'next2'"
	if err == nil || err.Error() != expectedErrStr {
		t.Errorf("Finalize() error=%v, want=%v", err.Error(), expectedErrStr)
	}
	if err := a.WriteTextTo(log.Writer()); err != nil {
		t.Error(err)
	}
}

func TestEmitter_LabelForwards_OutOfRange(t *testing.T) {
	a := NewEmitter(make([]byte, 0x100), true)
	a.SEP(0x30)
	a.LDA_imm8_b(0x01)
	a.BNE("next")
	a.RTS()
	for i := 0; i < 127; i++ {
		a.NOP()
	}
	a.Label("next")
	a.CMP_imm8_b(0x02)
	a.RTS()
	err := a.Finalize()
	expectedErrStr := "branch from 0x000006 to 0x000086 too far for signed 8-bit; diff=128"
	if err == nil || err.Error() != expectedErrStr {
		t.Errorf("Finalize() error=%v, want=%v", err.Error(), expectedErrStr)
	}
	if err := a.WriteTextTo(log.Writer()); err != nil {
		t.Error(err)
	}
}

func TestEmitter_OpenHCPortal(t *testing.T) {
	a := NewEmitter(make([]byte, 0x200), true)

	a.SetBase(0x707c00)

	{
		a.Comment("check if in HC overworld:")
		a.SEP(0x30)

		// check if in dungeon:
		a.LDA_dp(0x1B)
		a.BNE_imm8(0x6F - 0x06) // exit
		// check if in HC overworld:
		a.LDA_dp(0x8A)
		a.CMP_imm8_b(0x1B)
		a.BNE_imm8(0x6F - 0x0C) // exit

		a.Comment("find free sprite slot:")
		a.LDX_imm8_b(0x0f)      //   LDX   #$0F
		_ = 0                   // loop:
		a.LDA_abs_x(0x0DD0)     //   LDA.w $0DD0,X
		a.BEQ_imm8(0x05)        //   BEQ   found
		a.DEX()                 //   DEX
		a.BPL_imm8(-8)          //   BPL   loop
		a.BRA_imm8(0x6F - 0x18) //   BRA   exit
		_ = 0                   // found:

		a.Comment("open portal at HC:")
		// Y:
		a.LDA_imm8_b(0x50)
		a.STA_abs_x(0x0D00)
		a.LDA_imm8_b(0x08)
		a.STA_abs_x(0x0D20)
		// X:
		a.LDA_imm8_b(0xe0)
		a.STA_abs_x(0x0D10)
		a.LDA_imm8_b(0x07)
		a.STA_abs_x(0x0D30)
		// zeros:
		a.STZ_abs_x(0x0D40)
		a.STZ_abs_x(0x0D50)
		a.STZ_abs_x(0x0D60)
		a.STZ_abs_x(0x0D70)
		a.STZ_abs_x(0x0D80)
		// gfx?
		a.LDA_imm8_b(0x01)
		a.STA_abs_x(0x0D90)
		// hitbox/persist:
		a.STA_abs_x(0x0F60)
		// zeros:
		a.STZ_abs_x(0x0DA0)
		a.STZ_abs_x(0x0DB0)
		a.STZ_abs_x(0x0DC0)
		// active
		a.LDA_imm8_b(0x09)
		a.STA_abs_x(0x0DD0)
		// zeros:
		a.STZ_abs_x(0x0DE0)
		a.STZ_abs_x(0x0DF0)
		a.STZ_abs_x(0x0E00)
		a.STZ_abs_x(0x0E10)
		// whirlpool
		a.LDA_imm8_b(0xBA)
		a.STA_abs_x(0x0E20)
		// zeros:
		a.STZ_abs_x(0x0E30)
		// harmless
		a.LDA_imm8_b(0x80)
		a.STA_abs_x(0x0E40)
		// OAM:
		a.LDA_imm8_b(0x04)
		a.STA_abs_x(0x0F50)
		// exit:
		a.REP(0x30)
	}

	a.WriteTextTo(log.Writer())
}

func TestEmitter_CopyMem(t *testing.T) {
	writes := []struct {
		Address uint32
		Data    []byte
	}{{0x7E0000, []byte{0x10}}}

	a := NewEmitter(make([]byte, 0x200), true)

	// codeSize represents the total size of ASM code below without MVN blocks:
	const codeSize = 0x1B

	a.SetBase(0x002C00)

	// this NOP slide is necessary to avoid the problematic $2C00 address itself.
	a.NOP()
	a.NOP()

	a.Comment("preserve registers:")

	a.REP(0x30)
	a.PHA()
	a.PHX()
	a.PHY()
	a.PHD()

	// MVN affects B register:
	a.PHB()
	expectedCodeSize := codeSize + (12 * len(writes))
	srcOffs := uint16(0x2C00 + expectedCodeSize)
	for _, write := range writes {
		data := write.Data
		size := uint16(len(data))
		targetFXPakProAddress := write.Address
		destBank := uint8(0x7E + (targetFXPakProAddress-0xF5_0000)>>16)
		destOffs := uint16(targetFXPakProAddress & 0xFFFF)

		a.Comment(fmt.Sprintf("transfer $%04x bytes from $00:%04x to $%02x:%04x", size, srcOffs, destBank, destOffs))
		// A - Specifies the amount of bytes to transfer, minus 1
		a.LDA_imm16_w(size - 1)
		// X - Specifies the high and low bytes of the data source memory address
		a.LDX_imm16_w(srcOffs)
		// Y - Specifies the high and low bytes of the destination memory address
		a.LDY_imm16_w(destOffs)
		a.MVN(destBank, 0x00)

		srcOffs += size
	}
	a.PLB()

	a.Comment("disable NMI vector override:")
	a.SEP(0x30)
	a.LDA_imm8_b(0x00)
	a.STA_long(0x002C00)

	a.Comment("restore registers:")
	a.REP(0x30)
	a.PLD()
	a.PLY()
	a.PLX()
	a.PLA()

	a.Comment("jump to original NMI:")
	a.JMP_indirect(0xFFEA)

	a.WriteTextTo(log.Writer())

	// bug check: make sure emitted code is the expected size
	if actual, expected := a.Len(), expectedCodeSize; actual != expected {
		panic(fmt.Errorf("bug check: emitted code size %d != %d", actual, expected))
	}

	// copy in the data to be written to WRAM:
	for _, write := range writes {
		a.EmitBytes(write.Data)
	}
}
