// based on pda/go6502/bus/bus.go
// Copyright 2013–2014 Paul Annesley, released under MIT license
// Copyright 2019      Piotr Meyer

// XXX - should I define eaRead(uint32) and bankRead(uint8, uint16)?
// Or maybe Read(uint32)

package bus

import (
	"fmt"
	"github.com/alttpo/snes/emulator/memory"
)

type busEntry struct {
	mem   memory.Memory
	name  string
	start uint32
	end   uint32
}

// original algorithm was built around array of busEntry segments.
// iterating over declared segments gives us a smallest memory usage
// at expense of O(n) cost when number of declared segments (n) grows

// there are two acceptable alternatives:
// 1 - arbitraly set an smallest, allowed segment to 16 bytes (4 bits)
//     and it gives us a O(1) computational cost at expense of max
//     2^20 array slots (4 bytes each - pointer to function size?)

// 2 - made the same as above but single byte pointer that selects
//     index from pointers table

// 3 - split above into two tables: [00-ff] for banks. Particular index
//     may be nil (no mapping or no memory present) or contains array
//     of fixed size [4096 pointers for example, when combined with 1]

// Bus is a 24-bit address passed via 32-bit variable, 8-bit data bus,
// which maps reads and writes at different locations to different backend
// Memory. For example the lower 32K could be RAM, the upper 8KB ROM, and
// some I/O in the middle.

// implement variant 1 - static table of 2^20 pointers to 16-bytes segments
// simplest, fastest and with greater memory usage
//
type Bus struct {
	EA      uint32                 // last memory access - r/w
	Write   bool                   // is write op?
	segment [1048576]memory.Memory // 2^10 because segments are 4bits length
	entries []busEntry
}

func (b *Bus) String() string {
	return fmt.Sprintf("Address bus (TODO: describe)")
}

func (b *Bus) Clear() {
	//b.Mem.Clear()        // XXX - not implemented yet
}

func New() (*Bus, error) {
	b := &Bus{}
	b.Init(10)
	return b, nil
}

func NewWithSizeHint(cap int) (*Bus, error) {
	b := &Bus{}
	b.Init(cap)
	return b, nil
}

func (b *Bus) Init(cap int) {
	b.entries = make([]busEntry, 0, cap)
}

// There are two variants possible:
// handler, "name", start, size
// handler, "name", start, end      <- currently selected
func (b *Bus) Attach(mem memory.Memory, name string, start uint32, end uint32) error {

	if (start & 0xf) != 0 {
		//mylog.Logger.Log(fmt.Sprintf("start are not 4-bit aligned: %06X", start))
		return fmt.Errorf("start are not 4-bit aligned: %06X", start)
	}

	if ((end + 1) & 0xf) != 0 {
		//mylog.Logger.Log(fmt.Sprintf("bus: end are not 4-bit aligned: %06X", end))
		return fmt.Errorf("end are not 4-bit aligned: %06X", end)
	}

	for x := (start >> 4); x <= (end >> 4); x++ {
		//fmt.Printf("%v", x)
		b.segment[x] = mem
	}
	//fmt.Printf("0x3ffff: %v\n", b.segment[0x3ffff>>4])

	entry := busEntry{mem: mem, name: name, start: start, end: end}
	//mylog.Logger.Log(fmt.Sprintf("bus attach: %-20v %06x %06x", mem, start, end))
	b.entries = append(b.entries, entry)
	return nil
}

// Shutdown tells the address bus a shutdown is occurring, and to pass the
// message on to subordinates.
func (b *Bus) Shutdown() {
	for _, be := range b.entries {
		be.mem.Shutdown()
	}
}

// Read returns the byte from memory mapped to the given address.
// e.g. if ROM is mapped to 0xC000, then Read(0xC0FF) returns the byte at
// 0x00FF in that RAM device.
func (b *Bus) EaRead(a uint32) byte {
	mem := b.segment[a>>4]
	if mem == nil {
		panic(fmt.Errorf("No backend for address 0x%06X index %06x", a, a>>4))
	}
	b.EA = a // for debug interface
	b.Write = false
	value := mem.Read(a)
	return value
}

func (b *Bus) EaRead24_wrap(bank byte, addr uint16) uint32 {
	bank32 := uint32(bank) << 16
	offs := uint32(addr)

	m0 := b.segment[(bank32|offs+0)>>4]
	m1 := b.segment[(bank32|offs+1)>>4]
	m2 := b.segment[(bank32|offs+2)>>4]
	if m0 == nil || m1 == nil || m2 == nil {
		a := bank32 | (offs + 0)
		panic(fmt.Errorf("No backend for address 0x%06X index %06x", a, a>>4))
	}

	b.Write = false
	b.EA = bank32 | (offs + 0) // for debug interface
	ll := m0.Read(b.EA)
	b.EA = bank32 | (offs + 1)
	mm := m1.Read(b.EA)
	b.EA = bank32 | (offs + 2)
	hh := m2.Read(b.EA)

	return uint32(hh)<<16 | uint32(mm)<<8 | uint32(ll)
}

// Write the byte to the device mapped to the given address.
func (b *Bus) EaWrite(a uint32, value byte) {
	mem := b.segment[a>>4]
	if mem == nil {
		panic(fmt.Errorf("No backend for address 0x%06X index %06x", a, a>>4))
	}
	b.EA = a // for debug interface
	b.Write = true
	mem.Write(a, value)
}

func (b *Bus) EaDump(start uint32, end uint32, data []byte) int {
	// determine start and end segments:
	startK := (start & 0xff_fff0) >> 4
	endK := (end & 0xff_fff0) >> 4

	a := start
	i := 0

	// move segment by segment:
	for k := startK; k <= endK; k++ {
		s := b.segment[k]
		if s == nil {
			// skip the whole segment:
			for n := 0; a <= end && n < 16; n++ {
				a++
				i++
			}
			continue
		}

		// copy the whole segment:
		for n := 0; a <= end && n < 16; n++ {
			data[i] = s.Read(a)
			a++
			i++
		}
	}

	return i
}
