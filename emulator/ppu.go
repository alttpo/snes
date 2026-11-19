package emulator

// PPU is a Device that models the PPU well enough for code that writes VRAM, OAM and CGRAM through their ports
// (directly or with DMA), and for code that times itself by the scanline: it counts master clocks into the H and V
// counters, and raises NMI at the start of vblank and the H/V timer IRQ. It also handles the CPU registers tied to that
// timing: NMITIMEN ($4200), WRIO ($4201, for latching the counters), HTIME and VTIME ($4207..$420A), RDNMI ($4210),
// TIMEUP ($4211) and HVBJOY ($4212). It draws nothing.
//
// The timing is simplified: every scanline is 1,364 master clocks (341 dots of 4 clocks), a frame is 262 lines (312 if
// PAL) with no short, long or interlaced lines, the counters advance after each CPU instruction by the clocks it took,
// and events take effect at the instruction boundary at or after their dot.
type PPU struct {
	// PAL selects 312 lines per frame instead of 262.
	PAL bool

	VRAM  [0x10000]byte
	OAM   [544]byte
	CGRAM [512]byte

	// Regs are the last values written to $2100..$213F.
	Regs [0x40]byte

	// HClock is the master clock within the current scanline, 0..1363, V the scanline, 0..261 (or 311), and Frame the
	// number of frames begun since the PPU was attached or reset.
	HClock int
	V      int
	Frame  uint64

	s *System

	// VRAM port:
	vmain   byte
	vmadd   uint16
	vramBuf uint16
	// OAM port: the byte address (0..$3FF; $200 and up is the high table, mirrored every $20 bytes), the address reloaded at the start of vblank, and the low table's latched
	// even byte:
	oamAddr   uint16
	oamReload uint16
	oamLatch  byte
	// CGRAM port: the byte address (0..$1FF) and the latched even byte:
	cgAddr  uint16
	cgLatch byte

	// counter latches and their read flip-flops:
	hLatch, vLatch uint16
	latched        bool
	hFlip, vFlip   bool

	// CPU timing registers:
	nmitimen byte
	wrio     byte
	htime    uint16
	vtime    uint16
	nmiFlag  bool // RDNMI bit 7
	irqFlag  bool // TIMEUP bit 7; the IRQ line is asserted while set
}

const (
	ppuClocksPerLine = 1364
	ppuClocksPerDot  = 4
	// hblank, in dots: from 274 to the end of the line, and dot 0:
	ppuHBlankStart = 274
)

func (p *PPU) linesPerFrame() int {
	if p.PAL {
		return 312
	}
	return 262
}

// VBlankLine is the first line of vblank, when NMI comes: 225, or 240 with overscan (SETINI bit 2).
func (p *PPU) VBlankLine() int {
	if p.Regs[0x33]&0x04 != 0 {
		return 240
	}
	return 225
}

// HDot is the dot of the H counter, 0..340.
func (p *PPU) HDot() int { return p.HClock / ppuClocksPerDot }

// InVBlank reports whether the V counter is in vblank.
func (p *PPU) InVBlank() bool { return p.V >= p.VBlankLine() }

func (p *PPU) Attach(s *System) error {
	p.s = s
	p.Reset()
	s.MapIO(0x2100, 0x213F, p.readPPU, p.writePPU)
	s.MapIO(0x4200, 0x4201, nil, p.writeCPU)
	s.MapIO(0x4207, 0x420A, nil, p.writeCPU)
	s.MapIO(0x4210, 0x4212, p.readCPU, nil)
	return nil
}

// Reset puts the counters at the top of a frame and the registers as at power on (WRIO $FF, NMI and IRQ off). VRAM,
// OAM and CGRAM keep their contents.
func (p *PPU) Reset() {
	p.HClock, p.V, p.Frame = 0, 0, 0
	p.Regs = [0x40]byte{}
	p.vmain, p.vmadd, p.vramBuf = 0, 0, 0
	p.oamAddr, p.oamReload, p.oamLatch = 0, 0, 0
	p.cgAddr, p.cgLatch = 0, 0
	p.hLatch, p.vLatch, p.latched, p.hFlip, p.vFlip = 0, 0, false, false, false
	p.nmitimen, p.wrio, p.htime, p.vtime = 0, 0xFF, 0x1FF, 0x1FF
	p.nmiFlag, p.irqFlag = false, false
}

// Advance runs the counters forward by clocks master clocks, raising NMI and IRQ as they pass the dots for them.
func (p *PPU) Advance(clocks int) {
	for clocks > 0 {
		n := ppuClocksPerLine - p.HClock
		if clocks < n {
			n = clocks
		}
		from := p.HClock
		p.HClock += n
		clocks -= n
		p.hTimer(from, p.HClock)
		if p.HClock == ppuClocksPerLine {
			p.HClock = 0
			p.V++
			if p.V == p.linesPerFrame() {
				p.V = 0
				p.Frame++
			}
			p.lineStart()
		}
	}
	if p.irqFlag {
		p.s.CPU.TriggerIRQ()
	}
}

// lineStart is what happens at dot 0 of a line: vblank starts or ends, the OAM address reloads, and the V timer IRQ.
func (p *PPU) lineStart() {
	switch p.V {
	case p.VBlankLine():
		p.nmiFlag = true
		p.oamAddr = p.oamReload
		if p.nmitimen&0x80 != 0 {
			p.s.CPU.TriggerNMI()
		}
	case 0:
		p.nmiFlag = false
	}
	// the V timer alone fires at the start of line VTIME:
	if p.nmitimen&0x30 == 0x20 && uint16(p.V) == p.vtime {
		p.irqFlag = true
	}
}

// hTimer fires the H timer IRQ (alone, or with the V timer on line VTIME) if the H counter passed dot HTIME in from..to
// (master clocks of the current line).
func (p *PPU) hTimer(from, to int) {
	if p.nmitimen&0x10 == 0 {
		return
	}
	if p.nmitimen&0x20 != 0 && uint16(p.V) != p.vtime {
		return
	}
	if t := int(p.htime) * ppuClocksPerDot; from <= t && t < to {
		p.irqFlag = true
	}
}

// now returns the scanline and master clock within it at the access the CPU is making: the counters plus the clocks
// of the instruction so far (System.StepClocks), which they don't count until it ends.
func (p *PPU) now() (v, hclock int) {
	v, hclock = p.V, p.HClock+p.s.StepClocks()
	for hclock >= ppuClocksPerLine {
		hclock -= ppuClocksPerLine
		if v++; v == p.linesPerFrame() {
			v = 0
		}
	}
	return
}

// latchCounters latches the H and V counters for OPHCT and OPVCT.
func (p *PPU) latchCounters() {
	v, hclock := p.now()
	p.hLatch, p.vLatch = uint16(hclock/ppuClocksPerDot), uint16(v)
	p.latched = true
}

// vramAddr is the VRAM word address the port reads and writes: VMADD with VMAIN's address remapping applied.
func (p *PPU) vramAddr() uint16 {
	a := p.vmadd
	if n := uint16(p.vmain>>2&3) + 7; n > 7 {
		// remapping n = 8, 9 or 10 bits: aaaaaaaaBBBccccc -> aaaaaaaacccccBBB (for 8)
		mask := uint16(1)<<n - 1
		x := a & mask
		a = a&^mask | (x<<3|x>>(n-3))&mask
	}
	return a & 0x7FFF
}

func (p *PPU) vramIncrement() {
	switch p.vmain & 3 {
	case 0:
		p.vmadd++
	case 1:
		p.vmadd += 32
	default:
		p.vmadd += 128
	}
}

func (p *PPU) vramPrefetch() {
	a := uint32(p.vramAddr()) << 1
	p.vramBuf = uint16(p.VRAM[a]) | uint16(p.VRAM[a+1])<<8
}

func (p *PPU) writePPU(offs uint16, value byte) {
	p.Regs[offs-0x2100] = value
	switch offs {
	case 0x2102: // OAMADDL
		p.oamReload = p.oamReload&0x200 | uint16(value)<<1
		p.oamAddr = p.oamReload
	case 0x2103: // OAMADDH
		p.oamReload = p.oamReload&0x1FE | uint16(value&1)<<9
		p.oamAddr = p.oamReload
	case 0x2104: // OAMDATA
		a := p.oamAddr
		switch {
		case a >= 0x200:
			p.OAM[0x200|a&0x1F] = value
		case a&1 == 0:
			p.oamLatch = value
		default:
			p.OAM[a-1] = p.oamLatch
			p.OAM[a] = value
		}
		p.oamAddr = (a + 1) & 0x3FF
	case 0x2115: // VMAIN
		p.vmain = value
	case 0x2116: // VMADDL
		p.vmadd = p.vmadd&0xFF00 | uint16(value)
		p.vramPrefetch()
	case 0x2117: // VMADDH
		p.vmadd = p.vmadd&0x00FF | uint16(value)<<8
		p.vramPrefetch()
	case 0x2118: // VMDATAL
		p.VRAM[uint32(p.vramAddr())<<1] = value
		if p.vmain&0x80 == 0 {
			p.vramIncrement()
		}
	case 0x2119: // VMDATAH
		p.VRAM[uint32(p.vramAddr())<<1|1] = value
		if p.vmain&0x80 != 0 {
			p.vramIncrement()
		}
	case 0x2121: // CGADD
		p.cgAddr = uint16(value) << 1
	case 0x2122: // CGDATA
		if p.cgAddr&1 == 0 {
			p.cgLatch = value
		} else {
			p.CGRAM[p.cgAddr-1] = p.cgLatch
			p.CGRAM[p.cgAddr] = value & 0x7F
		}
		p.cgAddr = (p.cgAddr + 1) & 0x1FF
	}
}

func (p *PPU) readPPU(offs uint16) (value byte) {
	switch offs {
	case 0x2137: // SLHV: latch the counters if WRIO bit 7 is set
		if p.wrio&0x80 != 0 {
			p.latchCounters()
		}
	case 0x2138: // OAMDATAREAD
		a := p.oamAddr
		if a >= 0x200 {
			a = 0x200 | a&0x1F
		}
		value = p.OAM[a]
		p.oamAddr = (p.oamAddr + 1) & 0x3FF
	case 0x2139: // VMDATALREAD
		value = byte(p.vramBuf)
		if p.vmain&0x80 == 0 {
			p.vramPrefetch()
			p.vramIncrement()
		}
	case 0x213A: // VMDATAHREAD
		value = byte(p.vramBuf >> 8)
		if p.vmain&0x80 != 0 {
			p.vramPrefetch()
			p.vramIncrement()
		}
	case 0x213B: // CGDATAREAD
		value = p.CGRAM[p.cgAddr]
		p.cgAddr = (p.cgAddr + 1) & 0x1FF
	case 0x213C: // OPHCT: low byte, then bit 8
		value = byte(p.hLatch)
		if p.hFlip {
			value = byte(p.hLatch >> 8 & 1)
		}
		p.hFlip = !p.hFlip
	case 0x213D: // OPVCT: low byte, then bit 8
		value = byte(p.vLatch)
		if p.vFlip {
			value = byte(p.vLatch >> 8 & 1)
		}
		p.vFlip = !p.vFlip
	case 0x213E: // STAT77: PPU1 version 1
		value = 0x01
	case 0x213F: // STAT78: counter latched, PAL, PPU2 version 3; reading clears the latch flag and the flip-flops
		value = 0x03
		if p.latched {
			value |= 0x40
		}
		if p.PAL {
			value |= 0x10
		}
		p.latched, p.hFlip, p.vFlip = false, false, false
	default:
		// write-only registers read as the last value written (open bus, more or less):
		value = p.Regs[offs-0x2100]
	}
	return
}

func (p *PPU) writeCPU(offs uint16, value byte) {
	switch offs {
	case 0x4200: // NMITIMEN
		prev := p.nmitimen
		p.nmitimen = value
		// enabling NMI during vblank, before RDNMI is read, raises it at once:
		if prev&0x80 == 0 && value&0x80 != 0 && p.nmiFlag {
			p.s.CPU.TriggerNMI()
		}
		if value&0x30 == 0 {
			p.irqFlag = false
		}
	case 0x4201: // WRIO: bit 7 going from 1 to 0 latches the counters
		if p.wrio&0x80 != 0 && value&0x80 == 0 {
			p.latchCounters()
		}
		p.wrio = value
	case 0x4207: // HTIMEL
		p.htime = p.htime&0x100 | uint16(value)
	case 0x4208: // HTIMEH
		p.htime = p.htime&0xFF | uint16(value&1)<<8
	case 0x4209: // VTIMEL
		p.vtime = p.vtime&0x100 | uint16(value)
	case 0x420A: // VTIMEH
		p.vtime = p.vtime&0xFF | uint16(value&1)<<8
	}
}

func (p *PPU) readCPU(offs uint16) (value byte) {
	switch offs {
	case 0x4210: // RDNMI: vblank started since the last read; CPU version 2
		value = 0x02
		if p.nmiFlag {
			value |= 0x80
		}
		p.nmiFlag = false
	case 0x4211: // TIMEUP: the timer IRQ fired since the last read; reading acknowledges it
		if p.irqFlag {
			value = 0x80
		}
		p.irqFlag = false
	case 0x4212: // HVBJOY: vblank, hblank, auto joypad read busy (the first 3 lines of vblank, if enabled)
		v, hclock := p.now()
		if v >= p.VBlankLine() {
			value |= 0x80
			if p.nmitimen&1 != 0 && v < p.VBlankLine()+3 {
				value |= 0x01
			}
		}
		if d := hclock / ppuClocksPerDot; d == 0 || d >= ppuHBlankStart {
			value |= 0x40
		}
	}
	return
}
