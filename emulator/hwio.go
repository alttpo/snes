package emulator

import "github.com/alttpo/snes/emulator/memory"

// hwIO is the memory-mapped IO registers: a plain register file (memory.FakeHW) plus general purpose DMA ($420B)
// between the A bus and the B bus registers $2100..$21FF, the WRAM port (WMDATA $2180, WMADD $2181..$2183) and
// MEMSEL ($420D, FastROM). The registers a Device maps (System.MapIO) go to its handlers instead; written values are
// stored in the register file either way.
type hwIO struct {
	memory.FakeHW
	s *System
	// wmadd is the WRAM port's address, $00000..$1FFFF:
	wmadd uint32
	// the handlers the devices mapped, per register offset - $2000:
	readers [0x6000]func(offs uint16) byte
	writers [0x6000]func(offs uint16, value byte)
}

func (h *hwIO) Read(address uint32) byte {
	offs := address & 0xFFFF
	if r := h.readers[offs-0x2000]; r != nil {
		return r(uint16(offs))
	}
	switch offs {
	case 0x2180:
		v := h.s.WRAM[h.wmadd]
		h.wmadd = (h.wmadd + 1) & 0x1FFFF
		return v
	}
	return h.FakeHW.Read(address)
}

func (h *hwIO) Write(address uint32, value byte) {
	offs := address & 0xFFFF
	h.FakeHW.Write(address, value)
	if w := h.writers[offs-0x2000]; w != nil {
		w(uint16(offs), value)
		return
	}
	switch offs {
	case 0x2180:
		h.s.WRAM[h.wmadd] = value
		h.wmadd = (h.wmadd + 1) & 0x1FFFF
	case 0x2181:
		h.wmadd = h.wmadd&^0xFF | uint32(value)
	case 0x2182:
		h.wmadd = h.wmadd&^0xFF00 | uint32(value)<<8
	case 0x2183:
		h.wmadd = h.wmadd&^0x10000 | uint32(value&1)<<16
	case 0x420B:
		h.dmaStart(value)
	case 0x420D:
		h.s.setAccessSpeeds(value&1 != 0)
	}
}

// bOffsets are the B bus register offsets of each transfer unit, by DMA transfer pattern ($43x0 bits 0-2):
var bOffsets = [8][]uint32{
	{0}, {0, 1}, {0, 0}, {0, 0, 1, 1}, {0, 1, 2, 3}, {0, 1, 0, 1}, {0, 0}, {0, 0, 1, 1},
}

// DMA timing in master clocks: the overhead of a $420B write that starts any channel, of each channel, and of each
// byte transferred.
const (
	dmaStartClocks   = 18
	dmaChannelClocks = 8
	dmaByteClocks    = 8
)

// dmaStart runs the general purpose DMA transfers of the channels set in value, in order, as a write of value to
// $420B starts them. The CPU is halted meanwhile: the transfers' clocks are added to the bus's (bus.Bus.Clocks) in
// place of those of the accesses they make, which are not the CPU's.
func (h *hwIO) dmaStart(value byte) {
	if value == 0 {
		return
	}
	b := &h.s.Bus
	clocks, accesses := b.Clocks, b.Accesses
	n := uint64(dmaStartClocks)
	for ch := uint32(0); ch < 8; ch++ {
		if value&(1<<ch) != 0 {
			n += dmaChannelClocks + uint64(h.dma(ch))*dmaByteClocks
		}
	}
	b.Clocks, b.Accesses = clocks+n, accesses
}

// dma runs a general purpose DMA transfer on channel ch and returns the number of bytes it transferred.
func (h *hwIO) dma(ch uint32) int {
	reg := func(r uint32) byte { return h.FakeHW.Read(0x4300 | ch<<4 | r) }
	setReg := func(r uint32, v byte) { h.FakeHW.Write(0x4300|ch<<4|r, v) }

	mode := reg(0)
	bAddr := uint32(reg(1))
	aAddr := uint16(reg(2)) | uint16(reg(3))<<8
	aBank := uint32(reg(4)) << 16
	size := int(reg(5)) | int(reg(6))<<8
	if size == 0 {
		size = 0x10000
	}
	pattern := bOffsets[mode&7]
	for i := 0; i < size; i++ {
		b := 0x2100 | (bAddr+pattern[i%len(pattern)])&0xFF
		a := aBank | uint32(aAddr)
		if mode&0x80 != 0 {
			// B bus to A bus:
			h.s.Bus.EaWrite(a, h.s.Bus.EaRead(b))
		} else {
			h.s.Bus.EaWrite(b, h.s.Bus.EaRead(a))
		}
		switch {
		case mode&0x08 != 0:
			// fixed A bus address
		case mode&0x10 != 0:
			aAddr--
		default:
			aAddr++
		}
	}
	setReg(2, uint8(aAddr))
	setReg(3, uint8(aAddr>>8))
	setReg(5, 0)
	setReg(6, 0)
	return size
}
