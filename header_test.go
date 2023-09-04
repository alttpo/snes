package snes

import (
	"bytes"
	"reflect"
	"testing"
)

func TestHeader_ReadHeader(t *testing.T) {
	type args struct {
		b *bytes.Reader
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantHeader Header
	}{
		{
			name: "Parse VT FastROM header",
			args: args{
				b: bytes.NewReader([]byte{
					0x01, 0x8D, 0x24, 0x01, 0xE2, 0x30, 0x6B, 0x5C, 0x9C, 0xB1, 0xA1, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0x56, 0x54, 0x20, 0x71, 0x5A, 0x47, 0x4C, 0x72, 0x6D, 0x52, 0x76, 0x6B, 0x36, 0x20, 0x20, 0x20,
					0x20, 0x20, 0x20, 0x20, 0x20, 0x30, 0x02, 0x0B, 0x05, 0x00, 0x01, 0x00, 0x1F, 0x29, 0xE0, 0xD6,
					0x01, 0x00, 0x04, 0x00, 0xB7, 0xFF, 0xB7, 0xFF, 0x2C, 0x82, 0xAB, 0x98, 0x00, 0x80, 0xAF, 0x98,
					0xFF, 0xFF, 0xFF, 0xFF, 0xB7, 0xFF, 0x2C, 0x82, 0x2C, 0x82, 0x2C, 0x82, 0x00, 0x80, 0xD8, 0x82,
				}),
			},
			wantErr: false,
			wantHeader: Header{
				version:            1,
				MakerCode:          0x0,
				GameCode:           0x0,
				Fixed1:             [6]uint8{0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
				FlashSize:          0x0,
				ExpansionRAMSize:   0x0,
				SpecialVersion:     0x0,
				CoCPUType:          0x0,
				Title:              [21]uint8{0x56, 0x54, 0x20, 0x71, 0x5a, 0x47, 0x4c, 0x72, 0x6d, 0x52, 0x76, 0x6b, 0x36, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20},
				MapMode:            0x30,
				CartridgeType:      0x2,
				ROMSize:            0xb,
				RAMSize:            0x5,
				DestinationCode:    0x0,
				OldMakerCode:       0x1,
				MaskROMVersion:     0x0,
				ComplementCheckSum: 0x291f,
				CheckSum:           0xd6e0,
				NativeVectors: NativeVectors{
					Unused1: [4]uint8{0x1, 0x0, 0x4, 0x0},
					COP:     0xffb7,
					BRK:     0xffb7,
					ABORT:   0x822c,
					NMI:     0x98ab,
					Unused2: 0x8000,
					IRQ:     0x98af,
				},
				EmulatedVectors: EmulatedVectors{
					Unused1: [4]uint8{0xff, 0xff, 0xff, 0xff},
					COP:     0xffb7,
					Unused2: 0x822c,
					ABORT:   0x822c,
					NMI:     0x822c,
					RESET:   0x8000,
					IRQBRK:  0x82d8,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := Header{}
			if err := h.ReadHeader(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("ReadHeader() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.wantHeader, h) {
				t.Errorf("ReadHeader() header:\n%#vwantHeader\n%#v", h, tt.wantHeader)
			}
		})
	}
}
