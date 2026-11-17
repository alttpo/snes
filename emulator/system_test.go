package emulator

import (
	"testing"
)

func TestSystem_CreateEmulator(t *testing.T) {
	// verify our ROM, SRAM, WRAM mappings in the bus:
	tests := []struct {
		name   string
		verify func(t *testing.T, q *System)
	}{
		// ROM:
		{
			name: "ROM bank 00",
			verify: func(t *testing.T, q *System) {
				q.ROM[0x0000] = 0xFE
				if actual, expected := q.Bus.EaRead(0x00_8000), uint8(0xFE); actual != expected {
					t.Errorf("mapping failed, actual = %v, expected = %v", actual, expected)
				}
				if actual, expected := q.Bus.EaRead(0x80_8000), uint8(0xFE); actual != expected {
					t.Errorf("mapping failed, actual = %v, expected = %v", actual, expected)
				}
			},
		},
		{
			name: "ROM bank 01",
			verify: func(t *testing.T, q *System) {
				q.ROM[0x8000] = 0xFD
				if actual, expected := q.Bus.EaRead(0x01_8000), uint8(0xFD); actual != expected {
					t.Errorf("mapping failed, actual = %v, expected = %v", actual, expected)
				}
				if actual, expected := q.Bus.EaRead(0x81_8000), uint8(0xFD); actual != expected {
					t.Errorf("mapping failed, actual = %v, expected = %v", actual, expected)
				}
			},
		},
		// SRAM:
		{
			name: "SRAM bank 00",
			verify: func(t *testing.T, q *System) {
				q.SRAM[0x0000] = 0xFC
				if actual, expected := q.Bus.EaRead(0x70_0000), uint8(0xFC); actual != expected {
					t.Errorf("mapping failed, actual = %v, expected = %v", actual, expected)
				}
			},
		},
		{
			name: "SRAM bank 01",
			verify: func(t *testing.T, q *System) {
				q.SRAM[0x8000] = 0xFB
				if actual, expected := q.Bus.EaRead(0x71_0000), uint8(0xFB); actual != expected {
					t.Errorf("mapping failed, actual = %v, expected = %v", actual, expected)
				}
			},
		},
		// WRAM:
		{
			name: "WRAM $7E:0000",
			verify: func(t *testing.T, q *System) {
				q.WRAM[0x0000] = 0xFA
				if actual, expected := q.Bus.EaRead(0x7E_0000), uint8(0xFA); actual != expected {
					t.Errorf("mapping failed, actual = %v, expected = %v", actual, expected)
				}
			},
		},
		{
			name: "WRAM $00:0000",
			verify: func(t *testing.T, q *System) {
				q.WRAM[0x0000] = 0xF9
				if actual, expected := q.Bus.EaRead(0x00_0000), uint8(0xF9); actual != expected {
					t.Errorf("mapping failed, actual = %v, expected = %v", actual, expected)
				}
			},
		},
		{
			name: "WRAM $7E:1FFF",
			verify: func(t *testing.T, q *System) {
				q.WRAM[0x1FFF] = 0xF8
				if actual, expected := q.Bus.EaRead(0x7E_1FFF), uint8(0xF8); actual != expected {
					t.Errorf("mapping failed, actual = %v, expected = %v", actual, expected)
				}
			},
		},
		{
			name: "WRAM $00:1FFF",
			verify: func(t *testing.T, q *System) {
				q.WRAM[0x1FFF] = 0xF7
				if actual, expected := q.Bus.EaRead(0x00_1FFF), uint8(0xF7); actual != expected {
					t.Errorf("mapping failed, actual = %v, expected = %v", actual, expected)
				}
			},
		},
		{
			name: "WRAM $7E:2000",
			verify: func(t *testing.T, q *System) {
				q.WRAM[0x2000] = 0xF6
				if actual, expected := q.Bus.EaRead(0x7E_2000), uint8(0xF6); actual != expected {
					t.Errorf("mapping failed, actual = %v, expected = %v", actual, expected)
				}
			},
		},
		{
			name: "WRAM $7F:2000",
			verify: func(t *testing.T, q *System) {
				q.WRAM[0x12000] = 0xF5
				if actual, expected := q.Bus.EaRead(0x7F_2000), uint8(0xF5); actual != expected {
					t.Errorf("mapping failed, actual = %v, expected = %v", actual, expected)
				}
			},
		},
		{
			name: "WRAM $7F:FFFF",
			verify: func(t *testing.T, q *System) {
				q.WRAM[0x1FFFF] = 0xF4
				if actual, expected := q.Bus.EaRead(0x7F_FFFF), uint8(0xF4); actual != expected {
					t.Errorf("mapping failed, actual = %v, expected = %v", actual, expected)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &System{}
			if err := q.CreateEmulator(); err != nil {
				t.Fatal(err)
			}
			tt.verify(t, q)
		})
	}
}

func TestSystem_DMAFromWRAMPort(t *testing.T) {
	s := &System{}
	if err := s.CreateEmulator(); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 0x10; i++ {
		s.WRAM[0x0352+i] = byte(0x40 + i)
	}
	w := func(a uint32, v byte) { s.Bus.EaWrite(a, v) }
	// WMADD = $7E0352; channel 5: B bus $2180 to SRAM $70:6142, $10 bytes:
	w(0x2181, 0x52)
	w(0x2182, 0x03)
	w(0x2183, 0x00)
	w(0x4350, 0x80)
	w(0x4351, 0x80)
	w(0x4352, 0x42)
	w(0x4353, 0x61)
	w(0x4354, 0x70)
	w(0x4355, 0x10)
	w(0x4356, 0x00)
	w(0x420B, 0x20)
	for i := 0; i < 0x10; i++ {
		if s.SRAM[0x6142+i] != byte(0x40+i) {
			t.Fatalf("SRAM[$%04x] = $%02x; expected $%02x", 0x6142+i, s.SRAM[0x6142+i], 0x40+i)
		}
	}
	// the A bus address advanced and the size counted down:
	if s.Bus.EaRead(0x4352) != 0x52 || s.Bus.EaRead(0x4353) != 0x61 || s.Bus.EaRead(0x4355) != 0 {
		t.Errorf("DMA registers not updated")
	}
	// the WRAM port advanced too:
	if v := s.Bus.EaRead(0x2180); v != s.WRAM[0x0362] {
		t.Errorf("WMDATA read $%02x; expected the byte after the transfer", v)
	}
}
