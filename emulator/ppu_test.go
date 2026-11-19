package emulator

import (
	"testing"
)

// newTestSystem returns a System with the devices, the code at $00:8000 and handlers at $00:9000 (NMI) and $00:9100
// (IRQ), which store $01 to $10 or $11 of WRAM and then stop.
func newTestSystem(t *testing.T, code []byte, devices ...Device) *System {
	t.Helper()
	s := &System{Devices: devices}
	if err := s.CreateEmulator(); err != nil {
		t.Fatal(err)
	}
	copy(s.ROM[0x0000:], code)
	// NMI: LDA #$01 : STA $10 : STP; IRQ: LDA #$01 : STA $11 : STP (A is 8-bit in the tests)
	copy(s.ROM[0x1000:], []byte{0xA9, 0x01, 0x85, 0x10, 0xDB})
	copy(s.ROM[0x1100:], []byte{0xA9, 0x01, 0x85, 0x11, 0xDB})
	s.ROM[0x7FEA], s.ROM[0x7FEB] = 0x00, 0x90
	s.ROM[0x7FEE], s.ROM[0x7FEF] = 0x00, 0x91
	s.CPU.Reset()
	s.CPU.SetFlags(0x30) // M and X 8-bit, interrupts enabled
	s.SetPC(0x00_8000)
	return s
}

// runToStop steps until STP or maxClocks pass.
func runToStop(t *testing.T, s *System, maxClocks uint64) {
	t.Helper()
	start := s.Clocks
	for s.Clocks-start < maxClocks {
		if _, stopped := s.Step(); stopped {
			return
		}
	}
	t.Fatalf("did not stop within %d clocks; PC=$%06x", maxClocks, s.GetPC())
}

func TestCPU_StepClocks(t *testing.T) {
	for _, c := range []struct {
		name   string
		code   []byte
		pc     uint32
		clocks int
	}{
		// opcode fetch from SlowROM, then an internal operation:
		{"NOP", []byte{0xEA}, 0x00_8000, 8 + 6},
		// 2 fetches, a WRAM read:
		{"LDA dp", []byte{0xA5, 0x12}, 0x00_8000, 8 + 8 + 8},
		// 3 fetches, an IO register read:
		{"LDA $2137", []byte{0xAD, 0x37, 0x21}, 0x00_8000, 8 + 8 + 8 + 6},
		// 3 fetches, an old joypad register read:
		{"LDA $4016", []byte{0xAD, 0x16, 0x40}, 0x00_8000, 8 + 8 + 8 + 12},
	} {
		s := newTestSystem(t, c.code)
		s.SetPC(c.pc)
		cycles, _ := s.Step()
		if s.CPU.Clocks != c.clocks {
			t.Errorf("%s: %d clocks (%d cycles); expected %d", c.name, s.CPU.Clocks, cycles, c.clocks)
		}
		if s.Clocks != uint64(c.clocks) {
			t.Errorf("%s: System.Clocks %d; expected %d", c.name, s.Clocks, c.clocks)
		}
	}

	// FastROM: STZ $420D is SlowROM; NOP in bank $80 then fetches at 6 clocks once MEMSEL is set:
	s := newTestSystem(t, []byte{0xA9, 0x01, 0x8D, 0x0D, 0x42, 0xEA})
	s.SetPC(0x80_8000)
	s.Step()
	s.Step()
	s.Step()
	if s.CPU.Clocks != 6+6 {
		t.Errorf("FastROM NOP: %d clocks; expected %d", s.CPU.Clocks, 6+6)
	}
}

func TestPPU_CountersAndLatch(t *testing.T) {
	ppu := &PPU{}
	// LDA $2137 : LDA $213F : LDA $213C : STA $00 : LDA $213C : STA $01 : LDA $213D : STA $02 : LDA $213D : STA $03 :
	// LDA $213F : STA $04 : STP
	code := []byte{
		0xAD, 0x37, 0x21,
		0xAD, 0x3C, 0x21, 0x85, 0x00,
		0xAD, 0x3C, 0x21, 0x85, 0x01,
		0xAD, 0x3D, 0x21, 0x85, 0x02,
		0xAD, 0x3D, 0x21, 0x85, 0x03,
		0xAD, 0x3F, 0x21, 0x85, 0x04,
		0xDB,
	}
	s := newTestSystem(t, code, ppu)
	// each line started halts the CPU for DRAM refresh:
	ppu.Advance(300*(ppuClocksPerLine-ppuRefreshClocks) + 100*ppuClocksPerDot)
	if ppu.V != 300-262 || ppu.HDot() != 100 || ppu.Frame != 1 {
		t.Fatalf("after 300 lines and 100 dots: V=%d H=%d frame %d; expected V=38 H=100 frame 1", ppu.V, ppu.HDot(), ppu.Frame)
	}
	runToStop(t, s, 10_000)
	// the latch is at the read of LDA $2137, its last access (30 clocks, 7 dots, in):
	if h := uint16(s.WRAM[0x00]) | uint16(s.WRAM[0x01])<<8; h != 107 {
		t.Errorf("OPHCT = %d; expected 107", h)
	}
	if v := uint16(s.WRAM[0x02]) | uint16(s.WRAM[0x03])<<8; v != 38 {
		t.Errorf("OPVCT = %d; expected 38", v)
	}
	if s.WRAM[0x04]&0x40 != 0x40 {
		t.Errorf("STAT78 = $%02x; expected the latch flag", s.WRAM[0x04])
	}

	// WRIO bit 7 clear: SLHV doesn't latch:
	s = newTestSystem(t, append([]byte{0xA9, 0x00, 0x8D, 0x01, 0x42, 0xAD, 0x3F, 0x21}, code...), ppu)
	ppu.Advance(50 * (ppuClocksPerLine - ppuRefreshClocks))
	runToStop(t, s, 10_000)
	if v := uint16(s.WRAM[0x02]) | uint16(s.WRAM[0x03])<<8; v != 50 {
		t.Errorf("OPVCT after WRIO's latch = %d; expected 50, latched when WRIO bit 7 went low", v)
	}
}

func TestPPU_NMI(t *testing.T) {
	ppu := &PPU{}
	// LDA #$80 : STA $4200 : loop: BRA loop
	s := newTestSystem(t, []byte{0xA9, 0x80, 0x8D, 0x00, 0x42, 0x80, 0xFE}, ppu)
	s.CPU.RK = 0x00
	runToStop(t, s, 2*262*ppuClocksPerLine)
	if s.WRAM[0x10] != 1 {
		t.Fatalf("NMI handler did not run")
	}
	if ppu.V != 225 || ppu.HDot() > 80 {
		t.Errorf("NMI handled at V=%d H=%d; expected right after line 225 starts", ppu.V, ppu.HDot())
	}
	// native mode: PBR, PC (the loop's), flags pushed:
	sp := s.CPU.SP
	if pbr, pcl, pch := s.WRAM[sp+4], s.WRAM[sp+2], s.WRAM[sp+3]; pbr != 0 || pcl != 0x05 || pch != 0x80 {
		t.Errorf("pushed PBR $%02x PC $%02x%02x; expected $00 $8005", pbr, pch, pcl)
	}

	// WAI waits for the NMI:
	s = newTestSystem(t, []byte{0xA9, 0x80, 0x8D, 0x00, 0x42, 0xCB, 0xDB}, ppu)
	runToStop(t, s, 2*262*ppuClocksPerLine)
	if s.WRAM[0x10] != 1 || ppu.V != 225 {
		t.Errorf("after WAI: NMI handler ran %v at V=%d; expected at 225", s.WRAM[0x10] == 1, ppu.V)
	}

	// RDNMI: set from line 225, cleared by reading it:
	ppu.Reset()
	// the clocks to line 225, whose refresh then follows:
	ppu.Advance(225*(ppuClocksPerLine-ppuRefreshClocks) + ppuRefreshClocks)
	if v := s.Bus.EaRead(0x4210); v&0x80 == 0 {
		t.Errorf("RDNMI = $%02x in vblank; expected bit 7 set", v)
	}
	if v := s.Bus.EaRead(0x4210); v&0x80 != 0 {
		t.Errorf("RDNMI = $%02x read again; expected bit 7 clear", v)
	}
	if v := s.Bus.EaRead(0x4212); v&0x80 == 0 {
		t.Errorf("HVBJOY = $%02x in vblank; expected bit 7 set", v)
	}
}

func TestPPU_TimerIRQ(t *testing.T) {
	for _, c := range []struct {
		name     string
		nmitimen byte
		htime    uint16
		vtime    uint16
		v, h     int
	}{
		{"V timer", 0x20, 0, 100, 100, 0},
		{"H and V timer", 0x30, 200, 100, 100, 200},
		{"H timer", 0x10, 200, 0, 0, 200},
	} {
		ppu := &PPU{}
		// STZ $4208 : STZ $420A : LDA #htime : STA $4207 : LDA #vtime : STA $4209 : LDA #nmitimen : STA $4200 : CLI :
		// loop: BRA loop
		s := newTestSystem(t, []byte{
			0x9C, 0x08, 0x42, 0x9C, 0x0A, 0x42,
			0xA9, byte(c.htime), 0x8D, 0x07, 0x42,
			0xA9, byte(c.vtime), 0x8D, 0x09, 0x42,
			0xA9, c.nmitimen, 0x8D, 0x00, 0x42,
			0x58, 0x80, 0xFE,
		}, ppu)
		runToStop(t, s, 262*ppuClocksPerLine)
		if s.WRAM[0x11] != 1 {
			t.Errorf("%s: IRQ handler did not run", c.name)
			continue
		}
		if ppu.V != c.v || ppu.HDot() < c.h || ppu.HDot() > c.h+80 {
			t.Errorf("%s: IRQ at V=%d H=%d; expected V=%d H=%d", c.name, ppu.V, ppu.HDot(), c.v, c.h)
		}
		if v := s.Bus.EaRead(0x4211); v&0x80 == 0 {
			t.Errorf("%s: TIMEUP = $%02x; expected bit 7 set", c.name, v)
		}
	}
}

func TestPPU_VRAMDMA(t *testing.T) {
	ppu := &PPU{}
	// VMAIN = $80 (increment after the high byte), VMADD = $1000; DMA channel 0, mode 1, $7E:2000 -> $2118, 4 bytes:
	s := newTestSystem(t, []byte{
		0xA9, 0x80, 0x8D, 0x15, 0x21,
		0xA9, 0x00, 0x8D, 0x16, 0x21, 0xA9, 0x10, 0x8D, 0x17, 0x21,
		0xA9, 0x01, 0x8D, 0x00, 0x43, 0xA9, 0x18, 0x8D, 0x01, 0x43,
		0xA9, 0x00, 0x8D, 0x02, 0x43, 0xA9, 0x20, 0x8D, 0x03, 0x43, 0xA9, 0x7E, 0x8D, 0x04, 0x43,
		0xA9, 0x04, 0x8D, 0x05, 0x43, 0xA9, 0x00, 0x8D, 0x06, 0x43,
		0xA9, 0x01, 0x8D, 0x0B, 0x42,
		0xDB,
	}, ppu)
	copy(s.WRAM[0x2000:], []byte{0x11, 0x22, 0x33, 0x44})
	var last int
	for {
		_, stopped := s.Step()
		if stopped {
			break
		}
		last = s.CPU.Clocks
	}
	if got := ppu.VRAM[0x2000:0x2004]; string(got) != "\x11\x22\x33\x44" {
		t.Errorf("VRAM $1000 = % x; expected 11 22 33 44", got)
	}
	// STA $420B: 3 fetches and the write (30), then the DMA:
	if expected := 30 + dmaStartClocks + dmaChannelClocks + 4*dmaByteClocks; last != expected {
		t.Errorf("STA $420B with DMA: %d clocks; expected %d", last, expected)
	}
	// VMDATAREAD reads with prefetch: set the address, then read back the first word:
	s.Bus.EaWrite(0x2116, 0x00)
	s.Bus.EaWrite(0x2117, 0x10)
	if lo, hi := s.Bus.EaRead(0x2139), s.Bus.EaRead(0x213A); lo != 0x11 || hi != 0x22 {
		t.Errorf("VMDATAREAD = $%02x%02x; expected $2211", hi, lo)
	}
}

func TestJoypadsAndMath(t *testing.T) {
	j, m := &Joypads{}, &Math{}
	s := newTestSystem(t, nil, j, m)
	j.Input[0] = 0x8040 // B and X
	if lo, hi := s.Bus.EaRead(0x4218), s.Bus.EaRead(0x4219); lo != 0x40 || hi != 0x80 {
		t.Errorf("JOY1 = $%02x%02x; expected $8040", hi, lo)
	}
	s.Bus.EaWrite(0x4016, 1)
	s.Bus.EaWrite(0x4016, 0)
	var bits uint16
	for i := 0; i < 16; i++ {
		bits = bits<<1 | uint16(s.Bus.EaRead(0x4016)&1)
	}
	if bits != 0x8040 {
		t.Errorf("JOYSER0 shifted out $%04x; expected $8040", bits)
	}

	s.Bus.EaWrite(0x4202, 12)
	s.Bus.EaWrite(0x4203, 34)
	if p := uint16(s.Bus.EaRead(0x4216)) | uint16(s.Bus.EaRead(0x4217))<<8; p != 12*34 {
		t.Errorf("product %d; expected %d", p, 12*34)
	}
	s.Bus.EaWrite(0x4204, 0x39)
	s.Bus.EaWrite(0x4205, 0x30)
	s.Bus.EaWrite(0x4206, 100)
	q := uint16(s.Bus.EaRead(0x4214)) | uint16(s.Bus.EaRead(0x4215))<<8
	r := uint16(s.Bus.EaRead(0x4216)) | uint16(s.Bus.EaRead(0x4217))<<8
	if q != 0x3039/100 || r != 0x3039%100 {
		t.Errorf("quotient %d remainder %d; expected %d %d", q, r, 0x3039/100, 0x3039%100)
	}
}

func TestPPU_HDMAClocks(t *testing.T) {
	ppu := &PPU{HDMAClocks: 66}
	newTestSystem(t, nil, ppu)
	// lines 1..224 (line 0 started at reset) lose refresh and HDMA, lines 225.. only refresh; line 235's refresh comes
	// once the CPU reaches it:
	ppu.Advance(235*(ppuClocksPerLine-ppuRefreshClocks) - 224*66 + ppuRefreshClocks)
	if ppu.V != 235 || ppu.HClock != ppuRefreshClocks {
		t.Errorf("V=%d H clock %d; expected V=235 H clock %d", ppu.V, ppu.HClock, ppuRefreshClocks)
	}
}
