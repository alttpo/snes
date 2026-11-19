package emulator

import (
	"fmt"
	"io"

	"github.com/alttpo/snes/emulator/bus"
	"github.com/alttpo/snes/emulator/cpu65c816"
	"github.com/alttpo/snes/emulator/memory"
)

// System is an emulated SNES: a CPU on a bus with LoROM-mapped ROM, WRAM and SRAM, and the IO registers. Out of the
// box the IO registers are a plain register file (what is written reads back) with general purpose DMA and the WRAM
// port working, which is all that code needing only a CPU uses. Further hardware is added as Devices, set before
// CreateEmulator: a PPU (its IO registers, VRAM, OAM and CGRAM, the H/V counters and NMI and IRQ timing), Joypads, or
// a Math unit, or any other Device. The CPU and the clocked devices run in step on master clocks (see Step).
type System struct {
	// emulated system:
	Bus bus.Bus
	CPU cpu65c816.CPU

	ROM  [0x1000000]byte
	WRAM [0x20000]byte
	SRAM [0x10000]byte

	// Devices are the hardware beyond the above, attached by CreateEmulator in order.
	Devices []Device

	// Clocks is the master clocks Step has run.
	Clocks uint64

	Logger io.Writer

	hwio *hwIO
	// while Step runs the CPU: the bus's clocks when it started (see StepClocks):
	inStep     bool
	stepClocks uint64
	clocked    []Clocked
	resets     []Resetter
	// the master clocks of an access, per 256-byte page (see bus.Bus.Speeds):
	speeds [0x10000]uint8
}

// Device is hardware that a System has beyond its CPU, memory, DMA and WRAM port. CreateEmulator calls Attach, in
// which the device maps the IO registers it handles with System.MapIO (they take over from the plain register file)
// and keeps the System to raise interrupts on. A device that implements Clocked runs alongside the CPU, and one that
// implements Resetter is reset by System.Reset.
type Device interface {
	Attach(s *System) error
}

// Clocked is a Device that runs in step with the CPU.
type Clocked interface {
	// Advance runs the device forward by the master clocks the CPU just took.
	Advance(clocks int)
}

// Resetter is a Device that System.Reset resets.
type Resetter interface {
	Reset()
}

type Committer interface {
	Commit()
}

type Reserver interface {
	Reserve(n int)
}

func (s *System) CreateEmulator() (err error) {
	// Create CPU:
	s.CPU.Init(&s.Bus)

	// master clocks per access, for the CPU and the clocked devices to run in step:
	s.setAccessSpeeds(false)
	s.Bus.Speeds = &s.speeds

	// map in ROM to Bus; parts of this mapping will be overwritten:
	for b := uint32(0); b < 0x40; b++ {
		halfBank := b << 15
		bank := b << 16
		err = s.Bus.Attach(
			memory.NewRAM(s.ROM[halfBank:halfBank+0x8000], bank|0x8000),
			"rom",
			bank|0x8000,
			bank|0xFFFF,
		)
		if err != nil {
			return
		}

		// mirror:
		err = s.Bus.Attach(
			memory.NewRAM(s.ROM[halfBank:halfBank+0x8000], (bank+0x80_0000)|0x8000),
			"rom",
			(bank+0x80_0000)|0x8000,
			(bank+0x80_0000)|0xFFFF,
		)
		if err != nil {
			return
		}
	}

	// SRAM (banks 70-7D,F0-FF) (7E,7F) will be overwritten with WRAM:
	for b := uint32(0); b < uint32(len(s.SRAM)>>15); b++ {
		bank := b << 16
		halfBank := b << 15
		err = s.Bus.Attach(
			memory.NewRAM(s.SRAM[halfBank:halfBank+0x8000], bank+0x70_0000),
			"sram",
			bank+0x70_0000,
			bank+0x70_7FFF,
		)
		if err != nil {
			return
		}

		// mirror:
		err = s.Bus.Attach(
			memory.NewRAM(s.SRAM[halfBank:halfBank+0x8000], bank+0xF0_0000),
			"sram",
			bank+0xF0_0000,
			bank+0xF0_7FFF,
		)
		if err != nil {
			return
		}
	}

	// WRAM:
	{
		err = s.Bus.Attach(
			memory.NewRAM(s.WRAM[0:0x20000], 0x7E0000),
			"wram",
			0x7E_0000,
			0x7F_FFFF,
		)
		if err != nil {
			return
		}

		// map in first $2000 of each bank as a mirror of WRAM:
		for b := uint32(0); b < 0x40; b++ {
			bank := b << 16
			err = s.Bus.Attach(
				memory.NewRAM(s.WRAM[0:0x2000], bank),
				"wram",
				bank,
				bank|0x1FFF,
			)
			if err != nil {
				return
			}
		}
		for b := uint32(0x80); b < 0xC0; b++ {
			bank := b << 16
			err = s.Bus.Attach(
				memory.NewRAM(s.WRAM[0:0x2000], bank),
				"wram",
				bank,
				bank|0x1FFF,
			)
			if err != nil {
				return
			}
		}
	}

	// Memory-mapped IO registers:
	{
		s.hwio = &hwIO{s: s}
		for b := uint32(0); b < 0x70; b++ {
			bank := b << 16
			err = s.Bus.Attach(
				s.hwio,
				"hwio",
				bank|0x2000,
				bank|0x7FFF,
			)
			if err != nil {
				return
			}

			bank = (b + 0x80) << 16
			err = s.Bus.Attach(
				s.hwio,
				"hwio",
				bank|0x2000,
				bank|0x7FFF,
			)
			if err != nil {
				return
			}
		}
	}

	// devices:
	s.clocked, s.resets = nil, nil
	for _, d := range s.Devices {
		if err = d.Attach(s); err != nil {
			return
		}
		if c, ok := d.(Clocked); ok {
			s.clocked = append(s.clocked, c)
		}
		if r, ok := d.(Resetter); ok {
			s.resets = append(s.resets, r)
		}
	}

	return
}

// MapIO makes read and write handle the IO registers lo..hi (offsets $2000..$7FFF, in every bank the IO registers are
// mapped in), instead of the plain register file. Either may be nil: reads then return the last value written, and
// writes are only stored for that. Called by a Device's Attach.
func (s *System) MapIO(lo, hi uint16, read func(offs uint16) byte, write func(offs uint16, value byte)) {
	if s.hwio == nil {
		panic(fmt.Errorf("MapIO: called before CreateEmulator mapped the IO registers"))
	}
	if lo < 0x2000 || hi > 0x7FFF || lo > hi {
		panic(fmt.Errorf("MapIO: $%04x..$%04x outside of $2000..$7FFF", lo, hi))
	}
	for o := uint32(lo); o <= uint32(hi); o++ {
		if read != nil {
			s.hwio.readers[o-0x2000] = read
		}
		if write != nil {
			s.hwio.writers[o-0x2000] = write
		}
	}
}

// setAccessSpeeds sets the master clocks of an access to each page: 8 for WRAM, SRAM and SlowROM, 6 for the B bus
// and most CPU registers (and FastROM, banks $80..$FF, if fastROM), and 12 for the old joypad registers $4000..$41FF.
func (s *System) setAccessSpeeds(fastROM bool) {
	for p := range s.speeds {
		bank, offs := p>>8, (p&0xFF)<<8
		rom := uint8(8)
		if fastROM && bank >= 0x80 {
			rom = 6
		}
		switch {
		case bank >= 0x40 && bank < 0x80:
			s.speeds[p] = 8
		case bank >= 0xC0:
			s.speeds[p] = rom
		case offs < 0x2000:
			s.speeds[p] = 8
		case offs < 0x4000:
			s.speeds[p] = 6
		case offs < 0x4200:
			s.speeds[p] = 12
		case offs < 0x6000:
			s.speeds[p] = 6
		case offs < 0x8000:
			s.speeds[p] = 8
		default:
			s.speeds[p] = rom
		}
	}
}

// Reset resets the CPU (to the reset vector) and the devices that implement Resetter.
func (s *System) Reset() {
	s.CPU.Reset()
	for _, r := range s.resets {
		r.Reset()
	}
}

func (s *System) SetPC(pc uint32) {
	s.CPU.RK = byte(pc >> 16)
	s.CPU.PC = uint16(pc & 0xFFFF)
}

func (s *System) GetPC() uint32 {
	return uint32(s.CPU.RK)<<16 | uint32(s.CPU.PC)
}

// Step runs one CPU instruction (see cpu65c816.CPU.Step), then the clocked devices for the master clocks it took
// (CPU.Clocks), and returns the CPU cycles it took and whether the CPU is stopped.
func (s *System) Step() (cycles int, stopped bool) {
	s.inStep, s.stepClocks = true, s.Bus.Clocks
	cycles, stopped = s.CPU.Step()
	s.inStep = false
	clocks := s.CPU.Clocks
	s.Clocks += uint64(clocks)
	for _, c := range s.clocked {
		c.Advance(clocks)
	}
	return
}

// StepClocks returns the master clocks of the bus accesses the instruction Step is running has made so far, up to and
// including the one being made, or 0 outside of Step: how far past the clocked devices' state an IO register access
// is. A device that answers with its timing (as the PPU's counter latch does) adds it.
func (s *System) StepClocks() int {
	if !s.inStep {
		return 0
	}
	return int(s.Bus.Clocks - s.stepClocks)
}

func (s *System) RunUntil(targetPC uint32, maxCycles uint64) bool {
	// reserve space in the logger for enough room up to maxCycles if reasonable:
	if reserver, ok := s.Logger.(Reserver); ok {
		n := int(maxCycles)
		if maxCycles > 0x100 {
			n = 0x100
		}
		reserver.Reserve(40 * n / 2)
	}

	var oa [100]byte

	for cycles := uint64(0); cycles < maxCycles; {
		if s.Logger != nil {
			o := oa[:0]
			o = s.CPU.DisassembleCurrentPC(o)
			_, _ = s.Logger.Write(o)
		}
		if s.GetPC() == targetPC {
			break
		}
		nCycles, _ := s.Step()
		cycles += uint64(nCycles)
	}

	// commit to the logger:
	if committer, ok := s.Logger.(Committer); ok {
		committer.Commit()
	}

	return s.GetPC() == targetPC
}
