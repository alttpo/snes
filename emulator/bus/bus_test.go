package bus

import (
	"github.com/alttpo/snes/emulator/memory"
	"testing"
)

func TestBus_EaDump_2(t *testing.T) {
	var err error

	var b *Bus
	b, err = New()
	if err != nil {
		t.Fatal(err)
	}

	var rom []byte
	rom = make([]byte, 65536)

	b.Attach(memory.NewRAM(rom, 0), "rom", 0x00_0000, 0x7D_FFFF)

	fill := make([]byte, 2)
	n := b.EaDump(0x00_0000, 0x00_0001, fill)
	if n != len(fill) {
		t.Fatal(n)
	}
}

func TestBus_EaDump_15(t *testing.T) {
	var err error

	var b *Bus
	b, err = New()
	if err != nil {
		t.Fatal(err)
	}

	var rom []byte
	rom = make([]byte, 65536)

	b.Attach(memory.NewRAM(rom, 0), "rom", 0x00_0000, 0x7D_FFFF)

	fill := make([]byte, 0xF)
	n := b.EaDump(0x00_0000, 0x00_000E, fill)
	if n != len(fill) {
		t.Fatal(n)
	}
}

func TestBus_EaDump_16(t *testing.T) {
	var err error

	var b *Bus
	b, err = New()
	if err != nil {
		t.Fatal(err)
	}

	var rom []byte
	rom = make([]byte, 65536)

	b.Attach(memory.NewRAM(rom, 0), "rom", 0x00_0000, 0x7D_FFFF)

	fill := make([]byte, 0x10)
	n := b.EaDump(0x00_0000, 0x00_000F, fill)
	if n != len(fill) {
		t.Fatal(n)
	}
}

func TestBus_EaDump_17(t *testing.T) {
	var err error

	var b *Bus
	b, err = New()
	if err != nil {
		t.Fatal(err)
	}

	var rom []byte
	rom = make([]byte, 65536)

	b.Attach(memory.NewRAM(rom, 0), "rom", 0x00_0000, 0x7D_FFFF)

	fill := make([]byte, 0x11)
	n := b.EaDump(0x00_0000, 0x00_0010, fill)
	if n != len(fill) {
		t.Fatal(n)
	}
}

func TestBus_EaDump_32(t *testing.T) {
	var err error

	var b *Bus
	b, err = New()
	if err != nil {
		t.Fatal(err)
	}

	var rom []byte
	rom = make([]byte, 65536)

	b.Attach(memory.NewRAM(rom, 0), "rom", 0x00_0000, 0x7D_FFFF)

	fill := make([]byte, 0x20)
	n := b.EaDump(0x00_0000, 0x00_001F, fill)
	if n != len(fill) {
		t.Fatal(n)
	}
}
