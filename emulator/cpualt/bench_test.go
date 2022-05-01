package cpualt

import (
	"encoding/binary"
	"github.com/alttpo/snes/asm"
	"testing"
)

func BenchmarkCPU_Step(b *testing.B) {
	b.StopTimer()
	var err error

	var c CPU
	c.Init()

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

	c.Bus.AttachReader(0x00_8000, 0x00_FFFF, func(addr uint32) uint8 { return linearRom[addr-0x8000] })
	c.Bus.AttachWriter(0x00_8000, 0x00_FFFF, func(addr uint32, val uint8) { linearRom[addr-0x8000] = val })

	wram := make([]byte, 0x20000)
	c.Bus.AttachReader(0x00_0000, 0x00_1FFF, func(addr uint32) uint8 { return wram[addr&0x1FFF] })
	c.Bus.AttachWriter(0x00_0000, 0x00_1FFF, func(addr uint32, val uint8) { wram[addr&0x1FFF] = val })

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
