package asm

import (
	"bytes"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/alttpo/snes/emulator"
)

// The tests here run each emitter method's instruction on the CPU emulator and check its effect on the registers,
// flags and memory, as well as its encoding and its text.
//
// The machine starts in native mode with the code at $00:8000 (ROM), D = $0100, DB = $7E, S = $01FF, so that:
//   - direct page $10 is $00:0110 (WRAM);
//   - absolute $1234 is $7E:1234 (WRAM);
//   - long $7F:4000 is in WRAM too.
// The harness emits STP after the code under test and runs until the CPU stops.

const (
	testCodeBase = 0x008000
	testD        = 0x0100
	testDB       = 0x7E
	testS        = 0x01FF
)

type cpuRun struct {
	t   *testing.T
	s   *emulator.System
	end uint32
	// the address of the STP that stopped the CPU:
	stop uint32
}

func (r *cpuRun) setA(v uint16) {
	c := &r.s.CPU
	c.RA, c.RAl, c.RAh = v, uint8(v), uint8(v>>8)
}

func (r *cpuRun) setX(v uint16) {
	c := &r.s.CPU
	if c.X == 1 {
		v &= 0xFF
	}
	c.RX, c.RXl = v, uint8(v)
}

func (r *cpuRun) setY(v uint16) {
	c := &r.s.CPU
	if c.X == 1 {
		v &= 0xFF
	}
	c.RY, c.RYl = v, uint8(v)
}

func (r *cpuRun) getA() uint16 {
	c := &r.s.CPU
	if c.M == 1 {
		return uint16(c.RAh)<<8 | uint16(c.RAl)
	}
	return c.RA
}

func (r *cpuRun) getX() uint16 {
	if r.s.CPU.X == 1 {
		return uint16(r.s.CPU.RXl)
	}
	return r.s.CPU.RX
}

func (r *cpuRun) getY() uint16 {
	if r.s.CPU.X == 1 {
		return uint16(r.s.CPU.RYl)
	}
	return r.s.CPU.RY
}

func (r *cpuRun) poke(addr uint32, b ...byte) {
	for i, v := range b {
		r.s.Bus.EaWrite(addr+uint32(i), v)
	}
}

func (r *cpuRun) poke16(addr uint32, v uint16) {
	r.poke(addr, uint8(v), uint8(v>>8))
}

func (r *cpuRun) peek16(addr uint32) uint16 {
	return uint16(r.s.Bus.EaRead(addr)) | uint16(r.s.Bus.EaRead(addr+1))<<8
}

func (r *cpuRun) expectA(want uint16) {
	r.t.Helper()
	if got := r.getA(); got != want {
		r.t.Errorf("A = $%04x; want $%04x", got, want)
	}
}

func (r *cpuRun) expectX(want uint16) {
	r.t.Helper()
	if got := r.getX(); got != want {
		r.t.Errorf("X = $%04x; want $%04x", got, want)
	}
}

func (r *cpuRun) expectY(want uint16) {
	r.t.Helper()
	if got := r.getY(); got != want {
		r.t.Errorf("Y = $%04x; want $%04x", got, want)
	}
}

func (r *cpuRun) expectS(want uint16) {
	r.t.Helper()
	if got := r.s.CPU.SP; got != want {
		r.t.Errorf("S = $%04x; want $%04x", got, want)
	}
}

func (r *cpuRun) expectD(want uint16) {
	r.t.Helper()
	if got := r.s.CPU.RD; got != want {
		r.t.Errorf("D = $%04x; want $%04x", got, want)
	}
}

func (r *cpuRun) expectDB(want uint8) {
	r.t.Helper()
	if got := r.s.CPU.RDBR; got != want {
		r.t.Errorf("DB = $%02x; want $%02x", got, want)
	}
}

// expectP checks the flags in mask.
func (r *cpuRun) expectP(mask, want uint8) {
	r.t.Helper()
	if got := r.s.CPU.Flags() & mask; got != want&mask {
		r.t.Errorf("P & $%02x = $%02x; want $%02x", mask, got, want&mask)
	}
}

func (r *cpuRun) expectMem(addr uint32, want ...byte) {
	r.t.Helper()
	got := make([]byte, len(want))
	for i := range got {
		got[i] = r.s.Bus.EaRead(addr + uint32(i))
	}
	if !bytes.Equal(got, want) {
		r.t.Errorf("[$%06x] = % x; want % x", addr, got, want)
	}
}

const (
	flagC = 0x01
	flagZ = 0x02
	flagI = 0x04
	flagD = 0x08
	flagX = 0x10
	flagM = 0x20
	flagV = 0x40
	flagN = 0x80
)

type cpuTest struct {
	// name is the emitter method tested, optionally followed by "/" and a variant:
	name string
	// the initial flags; m and x set the register sizes for the emitter too:
	p     uint8
	init  func(r *cpuRun)
	emit  func(a *Emitter)
	bytes []byte
	check func(r *cpuRun)
	// the address of the STP expected to stop the CPU; zero for the one after the emitted code:
	stop uint32
}

func runCPUTest(t *testing.T, tt cpuTest) {
	s := &emulator.System{}
	if err := s.CreateEmulator(); err != nil {
		t.Fatal(err)
	}
	r := &cpuRun{t: t, s: s}

	a := NewEmitter(s.ROM[0:0x8000], true)
	a.SetBase(testCodeBase)
	a.AssumeREP(0x30)
	a.AssumeSEP(Flags(tt.p & 0x30))
	tt.emit(a)
	r.end = a.PC()
	a.STP()
	if err := a.Finalize(); err != nil {
		t.Fatal(err)
	}
	if got := a.Bytes()[:r.end-testCodeBase]; !bytes.Equal(got, tt.bytes) {
		t.Errorf("emitted % x; want % x", got, tt.bytes)
	}
	var text strings.Builder
	if err := a.WriteTextTo(&text); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(text.String(), "%!") || strings.Contains(text.String(), "!!") {
		t.Errorf("bad text:\n%s", text.String())
	}

	c := &s.CPU
	c.E = 0
	c.SetFlags(tt.p)
	c.SP = testS
	c.RD = testD
	c.RDBR = testDB
	s.SetPC(testCodeBase)
	if tt.init != nil {
		tt.init(r)
	}

	for i := 0; i < 100000 && !c.Stopped; i++ {
		c.Step()
	}
	if !c.Stopped {
		t.Fatalf("CPU did not stop; PC = $%06x", s.GetPC())
	}
	r.stop = uint32(c.PRK)<<16 | uint32(c.PPC)
	wantStop := tt.stop
	if wantStop == 0 {
		wantStop = r.end
	}
	if r.stop != wantStop {
		t.Errorf("stopped at $%06x; want $%06x", r.stop, wantStop)
	}
	if tt.check != nil {
		tt.check(r)
	}
}

// aluEmitters emit the instructions of the ORA, AND, EOR, ADC, STA, LDA, CMP and SBC group with the operands that the
// modes in aluModes expect:
var aluEmitters = map[string]func(a *Emitter){
	"ORA_dp_x_indirect":      func(a *Emitter) { a.ORA_dp_x_indirect(0x10) },
	"ORA_sr":                 func(a *Emitter) { a.ORA_sr(0x03) },
	"ORA_dp":                 func(a *Emitter) { a.ORA_dp(0x10) },
	"ORA_dp_indirect_long":   func(a *Emitter) { a.ORA_dp_indirect_long(0x10) },
	"ORA_imm16_w":            func(a *Emitter) { a.ORA_imm16_w(0x3C55) },
	"ORA_imm8_b":             func(a *Emitter) { a.ORA_imm8_b(0x55) },
	"ORA_abs":                func(a *Emitter) { a.ORA_abs(0x1234) },
	"ORA_long":               func(a *Emitter) { a.ORA_long(0x7F4000) },
	"ORA_dp_indirect_y":      func(a *Emitter) { a.ORA_dp_indirect_y(0x10) },
	"ORA_dp_indirect":        func(a *Emitter) { a.ORA_dp_indirect(0x10) },
	"ORA_sr_indirect_y":      func(a *Emitter) { a.ORA_sr_indirect_y(0x03) },
	"ORA_dp_x":               func(a *Emitter) { a.ORA_dp_x(0x10) },
	"ORA_dp_indirect_long_y": func(a *Emitter) { a.ORA_dp_indirect_long_y(0x10) },
	"ORA_abs_y":              func(a *Emitter) { a.ORA_abs_y(0x1234) },
	"ORA_abs_x":              func(a *Emitter) { a.ORA_abs_x(0x1234) },
	"ORA_long_x":             func(a *Emitter) { a.ORA_long_x(0x7F4000) },
	"AND_dp_x_indirect":      func(a *Emitter) { a.AND_dp_x_indirect(0x10) },
	"AND_sr":                 func(a *Emitter) { a.AND_sr(0x03) },
	"AND_dp":                 func(a *Emitter) { a.AND_dp(0x10) },
	"AND_dp_indirect_long":   func(a *Emitter) { a.AND_dp_indirect_long(0x10) },
	"AND_imm16_w":            func(a *Emitter) { a.AND_imm16_w(0x3C55) },
	"AND_imm8_b":             func(a *Emitter) { a.AND_imm8_b(0x55) },
	"AND_abs":                func(a *Emitter) { a.AND_abs(0x1234) },
	"AND_long":               func(a *Emitter) { a.AND_long(0x7F4000) },
	"AND_dp_indirect_y":      func(a *Emitter) { a.AND_dp_indirect_y(0x10) },
	"AND_dp_indirect":        func(a *Emitter) { a.AND_dp_indirect(0x10) },
	"AND_sr_indirect_y":      func(a *Emitter) { a.AND_sr_indirect_y(0x03) },
	"AND_dp_x":               func(a *Emitter) { a.AND_dp_x(0x10) },
	"AND_dp_indirect_long_y": func(a *Emitter) { a.AND_dp_indirect_long_y(0x10) },
	"AND_abs_y":              func(a *Emitter) { a.AND_abs_y(0x1234) },
	"AND_abs_x":              func(a *Emitter) { a.AND_abs_x(0x1234) },
	"AND_long_x":             func(a *Emitter) { a.AND_long_x(0x7F4000) },
	"EOR_dp_x_indirect":      func(a *Emitter) { a.EOR_dp_x_indirect(0x10) },
	"EOR_sr":                 func(a *Emitter) { a.EOR_sr(0x03) },
	"EOR_dp":                 func(a *Emitter) { a.EOR_dp(0x10) },
	"EOR_dp_indirect_long":   func(a *Emitter) { a.EOR_dp_indirect_long(0x10) },
	"EOR_imm16_w":            func(a *Emitter) { a.EOR_imm16_w(0x3C55) },
	"EOR_imm8_b":             func(a *Emitter) { a.EOR_imm8_b(0x55) },
	"EOR_abs":                func(a *Emitter) { a.EOR_abs(0x1234) },
	"EOR_long":               func(a *Emitter) { a.EOR_long(0x7F4000) },
	"EOR_dp_indirect_y":      func(a *Emitter) { a.EOR_dp_indirect_y(0x10) },
	"EOR_dp_indirect":        func(a *Emitter) { a.EOR_dp_indirect(0x10) },
	"EOR_sr_indirect_y":      func(a *Emitter) { a.EOR_sr_indirect_y(0x03) },
	"EOR_dp_x":               func(a *Emitter) { a.EOR_dp_x(0x10) },
	"EOR_dp_indirect_long_y": func(a *Emitter) { a.EOR_dp_indirect_long_y(0x10) },
	"EOR_abs_y":              func(a *Emitter) { a.EOR_abs_y(0x1234) },
	"EOR_abs_x":              func(a *Emitter) { a.EOR_abs_x(0x1234) },
	"EOR_long_x":             func(a *Emitter) { a.EOR_long_x(0x7F4000) },
	"ADC_dp_x_indirect":      func(a *Emitter) { a.ADC_dp_x_indirect(0x10) },
	"ADC_sr":                 func(a *Emitter) { a.ADC_sr(0x03) },
	"ADC_dp":                 func(a *Emitter) { a.ADC_dp(0x10) },
	"ADC_dp_indirect_long":   func(a *Emitter) { a.ADC_dp_indirect_long(0x10) },
	"ADC_imm16_w":            func(a *Emitter) { a.ADC_imm16_w(0x3C55) },
	"ADC_imm8_b":             func(a *Emitter) { a.ADC_imm8_b(0x55) },
	"ADC_abs":                func(a *Emitter) { a.ADC_abs(0x1234) },
	"ADC_long":               func(a *Emitter) { a.ADC_long(0x7F4000) },
	"ADC_dp_indirect_y":      func(a *Emitter) { a.ADC_dp_indirect_y(0x10) },
	"ADC_dp_indirect":        func(a *Emitter) { a.ADC_dp_indirect(0x10) },
	"ADC_sr_indirect_y":      func(a *Emitter) { a.ADC_sr_indirect_y(0x03) },
	"ADC_dp_x":               func(a *Emitter) { a.ADC_dp_x(0x10) },
	"ADC_dp_indirect_long_y": func(a *Emitter) { a.ADC_dp_indirect_long_y(0x10) },
	"ADC_abs_y":              func(a *Emitter) { a.ADC_abs_y(0x1234) },
	"ADC_abs_x":              func(a *Emitter) { a.ADC_abs_x(0x1234) },
	"ADC_long_x":             func(a *Emitter) { a.ADC_long_x(0x7F4000) },
	"STA_dp_x_indirect":      func(a *Emitter) { a.STA_dp_x_indirect(0x10) },
	"STA_sr":                 func(a *Emitter) { a.STA_sr(0x03) },
	"STA_dp":                 func(a *Emitter) { a.STA_dp(0x10) },
	"STA_dp_indirect_long":   func(a *Emitter) { a.STA_dp_indirect_long(0x10) },
	"STA_abs":                func(a *Emitter) { a.STA_abs(0x1234) },
	"STA_long":               func(a *Emitter) { a.STA_long(0x7F4000) },
	"STA_dp_indirect_y":      func(a *Emitter) { a.STA_dp_indirect_y(0x10) },
	"STA_dp_indirect":        func(a *Emitter) { a.STA_dp_indirect(0x10) },
	"STA_sr_indirect_y":      func(a *Emitter) { a.STA_sr_indirect_y(0x03) },
	"STA_dp_x":               func(a *Emitter) { a.STA_dp_x(0x10) },
	"STA_dp_indirect_long_y": func(a *Emitter) { a.STA_dp_indirect_long_y(0x10) },
	"STA_abs_y":              func(a *Emitter) { a.STA_abs_y(0x1234) },
	"STA_abs_x":              func(a *Emitter) { a.STA_abs_x(0x1234) },
	"STA_long_x":             func(a *Emitter) { a.STA_long_x(0x7F4000) },
	"LDA_dp_x_indirect":      func(a *Emitter) { a.LDA_dp_x_indirect(0x10) },
	"LDA_sr":                 func(a *Emitter) { a.LDA_sr(0x03) },
	"LDA_dp":                 func(a *Emitter) { a.LDA_dp(0x10) },
	"LDA_dp_indirect_long":   func(a *Emitter) { a.LDA_dp_indirect_long(0x10) },
	"LDA_imm16_w":            func(a *Emitter) { a.LDA_imm16_w(0x3C55) },
	"LDA_imm8_b":             func(a *Emitter) { a.LDA_imm8_b(0x55) },
	"LDA_abs":                func(a *Emitter) { a.LDA_abs(0x1234) },
	"LDA_long":               func(a *Emitter) { a.LDA_long(0x7F4000) },
	"LDA_dp_indirect_y":      func(a *Emitter) { a.LDA_dp_indirect_y(0x10) },
	"LDA_dp_indirect":        func(a *Emitter) { a.LDA_dp_indirect(0x10) },
	"LDA_sr_indirect_y":      func(a *Emitter) { a.LDA_sr_indirect_y(0x03) },
	"LDA_dp_x":               func(a *Emitter) { a.LDA_dp_x(0x10) },
	"LDA_dp_indirect_long_y": func(a *Emitter) { a.LDA_dp_indirect_long_y(0x10) },
	"LDA_abs_y":              func(a *Emitter) { a.LDA_abs_y(0x1234) },
	"LDA_abs_x":              func(a *Emitter) { a.LDA_abs_x(0x1234) },
	"LDA_long_x":             func(a *Emitter) { a.LDA_long_x(0x7F4000) },
	"CMP_dp_x_indirect":      func(a *Emitter) { a.CMP_dp_x_indirect(0x10) },
	"CMP_sr":                 func(a *Emitter) { a.CMP_sr(0x03) },
	"CMP_dp":                 func(a *Emitter) { a.CMP_dp(0x10) },
	"CMP_dp_indirect_long":   func(a *Emitter) { a.CMP_dp_indirect_long(0x10) },
	"CMP_imm16_w":            func(a *Emitter) { a.CMP_imm16_w(0x3C55) },
	"CMP_imm8_b":             func(a *Emitter) { a.CMP_imm8_b(0x55) },
	"CMP_abs":                func(a *Emitter) { a.CMP_abs(0x1234) },
	"CMP_long":               func(a *Emitter) { a.CMP_long(0x7F4000) },
	"CMP_dp_indirect_y":      func(a *Emitter) { a.CMP_dp_indirect_y(0x10) },
	"CMP_dp_indirect":        func(a *Emitter) { a.CMP_dp_indirect(0x10) },
	"CMP_sr_indirect_y":      func(a *Emitter) { a.CMP_sr_indirect_y(0x03) },
	"CMP_dp_x":               func(a *Emitter) { a.CMP_dp_x(0x10) },
	"CMP_dp_indirect_long_y": func(a *Emitter) { a.CMP_dp_indirect_long_y(0x10) },
	"CMP_abs_y":              func(a *Emitter) { a.CMP_abs_y(0x1234) },
	"CMP_abs_x":              func(a *Emitter) { a.CMP_abs_x(0x1234) },
	"CMP_long_x":             func(a *Emitter) { a.CMP_long_x(0x7F4000) },
	"SBC_dp_x_indirect":      func(a *Emitter) { a.SBC_dp_x_indirect(0x10) },
	"SBC_sr":                 func(a *Emitter) { a.SBC_sr(0x03) },
	"SBC_dp":                 func(a *Emitter) { a.SBC_dp(0x10) },
	"SBC_dp_indirect_long":   func(a *Emitter) { a.SBC_dp_indirect_long(0x10) },
	"SBC_imm16_w":            func(a *Emitter) { a.SBC_imm16_w(0x3C55) },
	"SBC_imm8_b":             func(a *Emitter) { a.SBC_imm8_b(0x55) },
	"SBC_abs":                func(a *Emitter) { a.SBC_abs(0x1234) },
	"SBC_long":               func(a *Emitter) { a.SBC_long(0x7F4000) },
	"SBC_dp_indirect_y":      func(a *Emitter) { a.SBC_dp_indirect_y(0x10) },
	"SBC_dp_indirect":        func(a *Emitter) { a.SBC_dp_indirect(0x10) },
	"SBC_sr_indirect_y":      func(a *Emitter) { a.SBC_sr_indirect_y(0x03) },
	"SBC_dp_x":               func(a *Emitter) { a.SBC_dp_x(0x10) },
	"SBC_dp_indirect_long_y": func(a *Emitter) { a.SBC_dp_indirect_long_y(0x10) },
	"SBC_abs_y":              func(a *Emitter) { a.SBC_abs_y(0x1234) },
	"SBC_abs_x":              func(a *Emitter) { a.SBC_abs_x(0x1234) },
	"SBC_long_x":             func(a *Emitter) { a.SBC_long_x(0x7F4000) },
}

type aluMode struct {
	name    string
	offset  byte
	operand []byte
	// init sets up the registers and memory for the operand to be at ea:
	init func(r *cpuRun)
	ea   uint32
}

var aluModes = []aluMode{
	{"dp_x_indirect", 0x01, []byte{0x10}, func(r *cpuRun) { r.setX(2); r.poke16(0x000112, 0x2000) }, 0x7E2000},
	{"sr", 0x03, []byte{0x03}, func(r *cpuRun) { r.s.CPU.SP = 0x01F0 }, 0x0001F3},
	{"dp", 0x05, []byte{0x10}, nil, 0x000110},
	{"dp_indirect_long", 0x07, []byte{0x10}, func(r *cpuRun) { r.poke(0x000110, 0x00, 0x30, 0x7F) }, 0x7F3000},
	{"abs", 0x0D, []byte{0x34, 0x12}, nil, 0x7E1234},
	{"long", 0x0F, []byte{0x00, 0x40, 0x7F}, nil, 0x7F4000},
	{"dp_indirect_y", 0x11, []byte{0x10}, func(r *cpuRun) { r.setY(4); r.poke16(0x000110, 0x2000) }, 0x7E2004},
	{"dp_indirect", 0x12, []byte{0x10}, func(r *cpuRun) { r.poke16(0x000110, 0x2000) }, 0x7E2000},
	{"sr_indirect_y", 0x13, []byte{0x03}, func(r *cpuRun) {
		r.s.CPU.SP = 0x01F0
		r.setY(4)
		r.poke16(0x0001F3, 0x2000)
	}, 0x7E2004},
	{"dp_x", 0x15, []byte{0x10}, func(r *cpuRun) { r.setX(2) }, 0x000112},
	{"dp_indirect_long_y", 0x17, []byte{0x10}, func(r *cpuRun) { r.setY(4); r.poke(0x000110, 0x00, 0x30, 0x7F) }, 0x7F3004},
	{"abs_y", 0x19, []byte{0x34, 0x12}, func(r *cpuRun) { r.setY(4) }, 0x7E1238},
	{"abs_x", 0x1D, []byte{0x34, 0x12}, func(r *cpuRun) { r.setX(2) }, 0x7E1236},
	{"long_x", 0x1F, []byte{0x00, 0x40, 0x7F}, func(r *cpuRun) { r.setX(2) }, 0x7F4002},
}

// aluOps are the results of A = $0F0F op $3C55 with the carry set (16-bit), and A = $AB0F op $55 (8-bit):
var aluOps = []struct {
	name     string
	base     byte
	a16, a8  uint16
	carry    uint8
	setCarry bool
}{
	{"ORA", 0x00, 0x3F5F, 0xAB5F, 1, false},
	{"AND", 0x20, 0x0C05, 0xAB05, 1, false},
	{"EOR", 0x40, 0x335A, 0xAB5A, 1, false},
	{"ADC", 0x60, 0x4B65, 0xAB65, 0, true},
	{"STA", 0x80, 0x0F0F, 0xAB0F, 1, false},
	{"LDA", 0xA0, 0x3C55, 0xAB55, 1, false},
	{"CMP", 0xC0, 0x0F0F, 0xAB0F, 0, true},
	{"SBC", 0xE0, 0xD2BA, 0xABBA, 0, true},
}

func aluTests() (tests []cpuTest) {
	for _, op := range aluOps {
		op := op
		for _, m := range aluModes {
			m := m
			tests = append(tests, cpuTest{
				name: op.name + "_" + m.name,
				p:    flagC,
				init: func(r *cpuRun) {
					r.setA(0x0F0F)
					if m.init != nil {
						m.init(r)
					}
					r.poke16(m.ea, 0x3C55)
				},
				emit:  aluEmitters[op.name+"_"+m.name],
				bytes: append([]byte{op.base + m.offset}, m.operand...),
				check: func(r *cpuRun) {
					r.expectA(op.a16)
					if op.name == "STA" {
						r.expectMem(m.ea, 0x0F, 0x0F)
					} else {
						r.expectMem(m.ea, 0x55, 0x3C)
					}
					if op.setCarry {
						r.expectP(flagC, op.carry)
					}
				},
			})
		}
		if op.name == "STA" {
			continue
		}
		tests = append(tests,
			cpuTest{
				name:  op.name + "_imm16_w",
				p:     flagC,
				init:  func(r *cpuRun) { r.setA(0x0F0F) },
				emit:  aluEmitters[op.name+"_imm16_w"],
				bytes: []byte{op.base + 0x09, 0x55, 0x3C},
				check: func(r *cpuRun) {
					r.expectA(op.a16)
					if op.setCarry {
						r.expectP(flagC, op.carry)
					}
				},
			},
			cpuTest{
				name:  op.name + "_imm8_b",
				p:     flagC | flagM,
				init:  func(r *cpuRun) { r.setA(0xAB0F) },
				emit:  aluEmitters[op.name+"_imm8_b"],
				bytes: []byte{op.base + 0x09, 0x55},
				check: func(r *cpuRun) {
					r.expectA(op.a8)
					if op.setCarry {
						r.expectP(flagC, op.carry)
					}
				},
			},
		)
	}
	return
}

// rmwEmitters emit the read-modify-write instructions with the operands that rmwModes expect:
var rmwEmitters = map[string]func(a *Emitter){
	"ASL":       func(a *Emitter) { a.ASL() },
	"ASL_dp":    func(a *Emitter) { a.ASL_dp(0x10) },
	"ASL_dp_x":  func(a *Emitter) { a.ASL_dp_x(0x10) },
	"ASL_abs":   func(a *Emitter) { a.ASL_abs(0x1234) },
	"ASL_abs_x": func(a *Emitter) { a.ASL_abs_x(0x1234) },
	"LSR":       func(a *Emitter) { a.LSR() },
	"LSR_dp":    func(a *Emitter) { a.LSR_dp(0x10) },
	"LSR_dp_x":  func(a *Emitter) { a.LSR_dp_x(0x10) },
	"LSR_abs":   func(a *Emitter) { a.LSR_abs(0x1234) },
	"LSR_abs_x": func(a *Emitter) { a.LSR_abs_x(0x1234) },
	"ROL":       func(a *Emitter) { a.ROL() },
	"ROL_dp":    func(a *Emitter) { a.ROL_dp(0x10) },
	"ROL_dp_x":  func(a *Emitter) { a.ROL_dp_x(0x10) },
	"ROL_abs":   func(a *Emitter) { a.ROL_abs(0x1234) },
	"ROL_abs_x": func(a *Emitter) { a.ROL_abs_x(0x1234) },
	"ROR":       func(a *Emitter) { a.ROR() },
	"ROR_dp":    func(a *Emitter) { a.ROR_dp(0x10) },
	"ROR_dp_x":  func(a *Emitter) { a.ROR_dp_x(0x10) },
	"ROR_abs":   func(a *Emitter) { a.ROR_abs(0x1234) },
	"ROR_abs_x": func(a *Emitter) { a.ROR_abs_x(0x1234) },
	"INC":       func(a *Emitter) { a.INC() },
	"INC_dp":    func(a *Emitter) { a.INC_dp(0x10) },
	"INC_dp_x":  func(a *Emitter) { a.INC_dp_x(0x10) },
	"INC_abs":   func(a *Emitter) { a.INC_abs(0x1234) },
	"INC_abs_x": func(a *Emitter) { a.INC_abs_x(0x1234) },
	"DEC":       func(a *Emitter) { a.DEC() },
	"DEC_dp":    func(a *Emitter) { a.DEC_dp(0x10) },
	"DEC_dp_x":  func(a *Emitter) { a.DEC_dp_x(0x10) },
	"DEC_abs":   func(a *Emitter) { a.DEC_abs(0x1234) },
	"DEC_abs_x": func(a *Emitter) { a.DEC_abs_x(0x1234) },
}

var rmwModes = []struct {
	name   string
	offset byte // from the dp opcode
	bytes  []byte
	init   func(r *cpuRun)
	ea     uint32
}{
	{"dp", 0x00, []byte{0x10}, nil, 0x000110},
	{"abs", 0x08, []byte{0x34, 0x12}, nil, 0x7E1234},
	{"dp_x", 0x10, []byte{0x10}, func(r *cpuRun) { r.setX(2) }, 0x000112},
	{"abs_x", 0x18, []byte{0x34, 0x12}, func(r *cpuRun) { r.setX(2) }, 0x7E1236},
}

// rmwOps are the results of op $8421 with the carry set, and the carry after:
var rmwOps = []struct {
	name   string
	dp     byte // the dp opcode
	acc    byte // the accumulator opcode
	result uint16
	carry  uint8
}{
	{"ASL", 0x06, 0x0A, 0x0842, 1},
	{"ROL", 0x26, 0x2A, 0x0843, 1},
	{"LSR", 0x46, 0x4A, 0x4210, 1},
	{"ROR", 0x66, 0x6A, 0xC210, 1},
	{"DEC", 0xC6, 0x3A, 0x8420, 1},
	{"INC", 0xE6, 0x1A, 0x8422, 1},
}

func rmwTests() (tests []cpuTest) {
	for _, op := range rmwOps {
		op := op
		tests = append(tests, cpuTest{
			name:  op.name,
			p:     flagC,
			init:  func(r *cpuRun) { r.setA(0x8421) },
			emit:  rmwEmitters[op.name],
			bytes: []byte{op.acc},
			check: func(r *cpuRun) {
				r.expectA(op.result)
				r.expectP(flagC, op.carry)
			},
		})
		for _, m := range rmwModes {
			m := m
			tests = append(tests, cpuTest{
				name: op.name + "_" + m.name,
				p:    flagC,
				init: func(r *cpuRun) {
					if m.init != nil {
						m.init(r)
					}
					r.poke16(m.ea, 0x8421)
				},
				emit:  rmwEmitters[op.name+"_"+m.name],
				bytes: append([]byte{op.dp + m.offset}, m.bytes...),
				check: func(r *cpuRun) {
					r.expectMem(m.ea, uint8(op.result), uint8(op.result>>8))
					r.expectP(flagC, op.carry)
				},
			})
		}
	}
	return
}

func TestEmitter_CPU(t *testing.T) {
	tests := append(aluTests(), rmwTests()...)
	tests = append(tests, otherTests()...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.emit == nil {
				t.Fatal("no emitter")
			}
			runCPUTest(t, tt)
		})
	}
}

// TestEmitter_CPU_Coverage checks that each instruction emitting method of Emitter has a test in TestEmitter_CPU.
func TestEmitter_CPU_Coverage(t *testing.T) {
	tested := map[string]bool{}
	tests := append(aluTests(), rmwTests()...)
	tests = append(tests, otherTests()...)
	for _, tt := range tests {
		tested[strings.SplitN(tt.name, "/", 2)[0]] = true
	}
	opMethod := regexp.MustCompile(`^[A-Z]{3}([A-Z]|_|$)`)
	et := reflect.TypeOf(&Emitter{})
	n := 0
	for i := 0; i < et.NumMethod(); i++ {
		name := et.Method(i).Name
		if !opMethod.MatchString(name) {
			continue
		}
		n++
		if !tested[name] {
			t.Errorf("%s has no test", name)
		}
	}
	if n < 280 {
		t.Errorf("only %d instruction methods found", n)
	}
}
