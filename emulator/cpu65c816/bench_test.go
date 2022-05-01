package cpu65c816

import (
	"encoding/binary"
	"github.com/alttpo/snes/asm"
	"github.com/alttpo/snes/emulator/bus"
	"github.com/alttpo/snes/emulator/memory"
	"testing"
)

func BenchmarkCPU_Step(b *testing.B) {
	b.StopTimer()
	var err error

	var u *bus.Bus
	u, err = bus.New()
	if err != nil {
		b.Fatal(err)
	}

	linearRom := make([]byte, 0x8000)
	binary.LittleEndian.PutUint16(linearRom[0x7FFC:], 0x8000)
	{
		a := asm.NewEmitter(linearRom, false)
		a.SetBase(0x00_8000)
		a.SEP(0x30)

		a.Label("inf")
		a.LDX_imm8_b(0xFF)

		a.Label("loop")
		a.PHX()
		a.PLY()
		a.DEX()
		a.BNE("loop")
		a.BRA("inf")

		err = a.Finalize()
		if err != nil {
			b.Fatal(err)
		}
	}

	wram := make([]byte, 0x20000)
	ram := memory.NewRAM(wram, 0)
	err = u.Attach(ram, "wram", 0x00_0000, 0x00_1FFF)
	if err != nil {
		b.Fatal(err)
	}

	rom := memory.NewROM(linearRom, 0x8000)
	err = u.Attach(rom, "rom", 0x00_8000, 0x00_FFFF)
	if err != nil {
		b.Fatal(err)
	}

	var c *CPU
	c, err = New(u)
	if err != nil {
		b.Fatal(err)
	}
	c.Reset()

	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, abort := c.Step()
		if abort {
			break
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(c.AllCycles), "cycles")
}
