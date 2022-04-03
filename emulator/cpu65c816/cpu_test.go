package cpu65c816

import (
	"bytes"
	"fmt"
	"github.com/alttpo/snes/emulator/bus"
	"github.com/alttpo/snes/emulator/memory"
	"os"
	"testing"
)

func TestCPU_Step(t *testing.T) {
	roms := []string{
		//"SNES/CPUTest/CPU/BRA/CPUBRA.sfc",
		//"SNES/CPUTest/CPU/ROR/CPUROR.sfc",
		//"SNES/CPUTest/CPU/CMP/CPUCMP.sfc",
		"SNES/CPUTest/CPU/RET/CPURET.sfc",
		//"SNES/CPUTest/CPU/INC/CPUINC.sfc",
		//"SNES/CPUTest/CPU/TRN/CPUTRN.sfc",
		//"SNES/CPUTest/CPU/SBC/CPUSBC.sfc",
		//"SNES/CPUTest/CPU/BIT/CPUBIT.sfc",
		//"SNES/CPUTest/CPU/ASL/CPUASL.sfc",
		//"SNES/CPUTest/CPU/LDR/CPULDR.sfc",
		//"SNES/CPUTest/CPU/ORA/CPUORA.sfc",
		//"SNES/CPUTest/CPU/JMP/CPUJMP.sfc",
		//"SNES/CPUTest/CPU/PHL/CPUPHL.sfc",
		//"SNES/CPUTest/CPU/AND/CPUAND.sfc",
		//"SNES/CPUTest/CPU/ROL/CPUROL.sfc",
		//"SNES/CPUTest/CPU/ADC/CPUADC.sfc",
		//"SNES/CPUTest/CPU/MSC/CPUMSC.sfc",
		//"SNES/CPUTest/CPU/DEC/CPUDEC.sfc",
		//"SNES/CPUTest/CPU/PSR/CPUPSR.sfc",
		//"SNES/CPUTest/CPU/STR/CPUSTR.sfc",
		//"SNES/CPUTest/CPU/LSR/CPULSR.sfc",
		//"SNES/CPUTest/CPU/EOR/CPUEOR.sfc",
		//"SNES/CPUTest/CPU/MOV/CPUMOV.sfc",
	}
	for _, romName := range roms {
		t.Run(romName, func(t *testing.T) {
			var err error

			var s *snes
			s, err = createSNES()
			if err != nil {
				t.Fatal(err)
			}

			{
				// load file into ROM:
				var f *os.File
				f, err = os.Open(romName)
				if err != nil {
					t.Fatal(err)
				}
				_, err = f.Read(s.ROM[:])
				if err != nil {
					t.Fatal(err)
				}
				err = f.Close()
				if err != nil {
					t.Fatal(err)
				}
			}

			mmio := testMMIO{t: t}
			for bank := uint32(0); bank < 0x100; bank++ {
				err = s.Bus.Attach(
					&mmio,
					"mmio",
					bank<<16|0x2000,
					bank<<16|0x7FFF,
				)
				if err != nil {
					t.Fatal(err)
				}
			}

			s.CPU.Reset()
			for s.CPU.AllCycles < 0x1000_0000 {
				//s.CPU.DisassembleCurrentPC(os.Stdout)
				//fmt.Println()
				s.CPU.Step()
				if s.CPU.PC == s.CPU.PPC {
					break
				}
			}
			fmt.Println()
			mmio.printScreen()

			if mmio.fail {
				t.FailNow()
			}
		})
	}
}

type snes struct {
	// emulated system:
	Bus *bus.Bus
	CPU *CPU

	ROM  [0x1000000]byte
	WRAM [0x20000]byte
	SRAM [0x10000]byte
}

func createSNES() (s *snes, err error) {
	s = &snes{}

	// create primary A bus for SNES:
	s.Bus, err = bus.NewWithSizeHint(0x40*2 + 0x10*2 + 1 + 0x70 + 0x80 + 0x70*2)
	// Create CPU:
	s.CPU, err = New(s.Bus)

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

	return
}

type testMMIO struct {
	t   *testing.T
	nmi byte

	scr      [0x10000]byte
	addr     uint16
	scrWrote bool

	fail bool
}

func (m *testMMIO) Read(address uint32) (value byte) {
	//defer func() { fmt.Printf("[$%06x] -> $%02x\n", address, value) }()
	if address&0xFFFF == 0x4210 {
		m.printScreen()
		if m.nmi == 0x42 {
			m.nmi = 0xc2
			value = m.nmi
			return
		}
		if m.nmi == 0xc2 {
			m.nmi = 0x42
			value = m.nmi
			return
		}

		m.nmi = 0x42
		value = m.nmi
		return
	}
	return 0
}

func (m *testMMIO) Write(address uint32, value byte) {
	offs := address & 0xFFFF
	//fmt.Printf("[$%04x] <- $%02x\n", offs, value)

	//if offs == 0x4210 {
	//	m.nmi = value
	//	return
	//}
	if offs == 0x2116 {
		// VMADDL
		m.addr = uint16(value) | m.addr&0xFF00
		//fmt.Fprintf(os.Stdout, "$%04x\n", m.addr)
	}
	if offs == 0x2117 {
		// VMADDH
		m.addr = uint16(value)<<8 | m.addr&0x00FF
		//fmt.Fprintf(os.Stdout, "$%04x\n", m.addr)
	}
	if offs == 0x2118 {
		//fmt.Fprintf(os.Stdout, "%c", value)
		m.scr[m.addr] = value
		m.addr++
		m.scrWrote = true
	}
}

func (m *testMMIO) Shutdown() {
	//TODO implement me
	panic("implement me")
}

func (m *testMMIO) Size() uint32 {
	//TODO implement me
	panic("implement me")
}

func (m *testMMIO) Clear() {
	//TODO implement me
	panic("implement me")
}

func (m *testMMIO) Dump(address uint32) []byte {
	//TODO implement me
	panic("implement me")
}

func (m *testMMIO) printScreen() {
	if !m.fail {
		failText := []byte("FAIL")
		if bytes.Contains(m.scr[0x7C00:], failText) {
			m.fail = true
		}
	}

	if m.scrWrote {
		for y := 0; y < 32; y++ {
			line := m.scr[0x7C00+y<<5 : 0x7C00+y<<5+32]
			for x := 0; x < 32; x++ {
				fmt.Printf("%c", line[x])
			}
			fmt.Println()
		}
		fmt.Println("--------------------------------")

		m.scrWrote = false
	}
}
