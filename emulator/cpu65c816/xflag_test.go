package cpu65c816

import (
	"encoding/binary"
	"testing"

	"github.com/alttpo/snes/asm"
	"github.com/alttpo/snes/emulator/bus"
	"github.com/alttpo/snes/emulator/memory"
)

// Setting the x flag clears the high bytes of X and Y; clearing it again leaves them zero.
func TestCPU_XFlagClearsIndexHighBytes(t *testing.T) {
	u, err := bus.New()
	if err != nil {
		t.Fatal(err)
	}
	linearRom := make([]byte, 0x8000)
	binary.LittleEndian.PutUint16(linearRom[0x7FFC:], 0x8000)
	a := asm.NewEmitter(linearRom, false)
	a.SetBase(0x00_8000)
	a.REP(0x30)
	a.LDX_imm16_w(0x1234)
	a.LDY_imm16_w(0x5678)
	a.SEP(0x10)
	a.REP(0x10)
	if err = a.Finalize(); err != nil {
		t.Fatal(err)
	}
	if err = u.Attach(memory.NewRAM(make([]byte, 0x2000), 0), "wram", 0x00_0000, 0x00_1FFF); err != nil {
		t.Fatal(err)
	}
	if err = u.Attach(memory.NewROM(linearRom, 0x8000), "rom", 0x00_8000, 0x00_FFFF); err != nil {
		t.Fatal(err)
	}
	c, err := New(u)
	if err != nil {
		t.Fatal(err)
	}
	c.Reset()
	// emulation mode after reset; leave it first:
	c.E = 0
	for i := 0; i < 5; i++ {
		c.Step()
	}
	if c.RX != 0x0034 || c.RY != 0x0078 {
		t.Errorf("X = $%04x, Y = $%04x; expected $0034 and $0078", c.RX, c.RY)
	}
}
