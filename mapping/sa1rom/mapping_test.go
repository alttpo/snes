package sa1rom

import "testing"

func TestPakAddressToBus(t *testing.T) {
	type args struct {
		pakAddr uint32
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		// ROM header:
		{
			name: "ROM header",
			args: args{
				pakAddr: 0x007FC0,
			},
			want: 0x00FFC0,
		},
		// ROM banks:
		// CX, DX:
		{
			name: "ROM $000000",
			args: args{
				pakAddr: 0x000000,
			},
			want: 0x008000,
		},
		{
			name: "ROM $007FFF",
			args: args{
				pakAddr: 0x007FFF,
			},
			want: 0x00FFFF,
		},
		{
			name: "ROM $008000",
			args: args{
				pakAddr: 0x008000,
			},
			want: 0x018000,
		},
		{
			name: "ROM $00FFFF",
			args: args{
				pakAddr: 0x00FFFF,
			},
			want: 0x01FFFF,
		},
		{
			name: "ROM $1F0000",
			args: args{
				pakAddr: 0x1F0000,
			},
			want: 0x3E8000,
		},
		{
			name: "ROM $1F7FFF",
			args: args{
				pakAddr: 0x1F7FFF,
			},
			want: 0x3EFFFF,
		},
		{
			name: "ROM $1F8000",
			args: args{
				pakAddr: 0x1F8000,
			},
			want: 0x3F8000,
		},
		{
			name: "ROM $1FFFFF",
			args: args{
				pakAddr: 0x1FFFFF,
			},
			want: 0x3FFFFF,
		},
		// EX, FX:
		{
			name: "ROM $200000",
			args: args{
				pakAddr: 0x200000,
			},
			want: 0x808000,
		},
		{
			name: "ROM $207FFF",
			args: args{
				pakAddr: 0x207FFF,
			},
			want: 0x80FFFF,
		},
		{
			name: "ROM $208000",
			args: args{
				pakAddr: 0x208000,
			},
			want: 0x818000,
		},
		{
			name: "ROM $20FFFF",
			args: args{
				pakAddr: 0x20FFFF,
			},
			want: 0x81FFFF,
		},
		{
			name: "ROM $3F0000",
			args: args{
				pakAddr: 0x3F0000,
			},
			want: 0xBE8000,
		},
		{
			name: "ROM $3F7FFF",
			args: args{
				pakAddr: 0x3F7FFF,
			},
			want: 0xBEFFFF,
		},
		{
			name: "ROM $3F8000",
			args: args{
				pakAddr: 0x3F8000,
			},
			want: 0xBF8000,
		},
		{
			name: "ROM $3FFFFF",
			args: args{
				pakAddr: 0x3FFFFF,
			},
			want: 0xBFFFFF,
		},
		// BW-RAM:
		{
			name: "BW-RAM $00000",
			args: args{
				pakAddr: 0xE00000,
			},
			want: 0x400000,
		},
		{
			name: "BW-RAM $07FFF",
			args: args{
				pakAddr: 0xE07FFF,
			},
			want: 0x407FFF,
		},
		{
			name: "BW-RAM $08000",
			args: args{
				pakAddr: 0xE08000,
			},
			want: 0x408000,
		},
		{
			name: "BW-RAM $0FFFF",
			args: args{
				pakAddr: 0xE0FFFF,
			},
			want: 0x40FFFF,
		},
		{
			name: "BW-RAM $10000",
			args: args{
				pakAddr: 0xE10000,
			},
			want: 0x410000,
		},
		{
			name: "BW-RAM $1FFFF",
			args: args{
				pakAddr: 0xE1FFFF,
			},
			want: 0x41FFFF,
		},
		{
			name: "BW-RAM $30000",
			args: args{
				pakAddr: 0xE30000,
			},
			want: 0x430000,
		},
		{
			name: "BW-RAM $3FFFF",
			args: args{
				pakAddr: 0xE3FFFF,
			},
			want: 0x43FFFF,
		},
		// mirrored:
		{
			name: "BW-RAM $40000",
			args: args{
				pakAddr: 0xE40000,
			},
			want: 0x400000,
		},
		{
			name: "BW-RAM $47FFF",
			args: args{
				pakAddr: 0xE47FFF,
			},
			want: 0x407FFF,
		},
		{
			name: "BW-RAM $48000",
			args: args{
				pakAddr: 0xE48000,
			},
			want: 0x408000,
		},
		{
			name: "BW-RAM $4FFFF",
			args: args{
				pakAddr: 0xE4FFFF,
			},
			want: 0x40FFFF,
		},
		{
			name: "BW-RAM $50000",
			args: args{
				pakAddr: 0xE50000,
			},
			want: 0x410000,
		},
		{
			name: "BW-RAM $5FFFF",
			args: args{
				pakAddr: 0xE5FFFF,
			},
			want: 0x41FFFF,
		},
		{
			name: "BW-RAM $70000",
			args: args{
				pakAddr: 0xE70000,
			},
			want: 0x430000,
		},
		{
			name: "BW-RAM $7FFFF",
			args: args{
				pakAddr: 0xE7FFFF,
			},
			want: 0x43FFFF,
		},
		// WRAM:
		{
			name: "WRAM $00000",
			args: args{
				pakAddr: 0xF50000,
			},
			want: 0x7E0000,
		},
		{
			name: "WRAM $01000",
			args: args{
				pakAddr: 0xF51000,
			},
			want: 0x7E1000,
		},
		{
			name: "WRAM $02000",
			args: args{
				pakAddr: 0xF52000,
			},
			want: 0x7E2000,
		},
		{
			name: "WRAM $0FFFF",
			args: args{
				pakAddr: 0xF5FFFF,
			},
			want: 0x7EFFFF,
		},
		{
			name: "WRAM $10000",
			args: args{
				pakAddr: 0xF60000,
			},
			want: 0x7F0000,
		},
		{
			name: "WRAM $1FFFF",
			args: args{
				pakAddr: 0xF6FFFF,
			},
			want: 0x7FFFFF,
		},
		// WRAM mirrors:
		{
			name: "WRAM mirror 1",
			args: args{
				pakAddr: 0xF70000,
			},
			want: 0x7E0000,
		},
		{
			name: "WRAM mirror 2",
			args: args{
				pakAddr: 0xF90000,
			},
			want: 0x7E0000,
		},
		{
			name: "WRAM mirror 3",
			args: args{
				pakAddr: 0xFB0000,
			},
			want: 0x7E0000,
		},
		{
			name: "WRAM mirror 4",
			args: args{
				pakAddr: 0xFD0000,
			},
			want: 0x7E0000,
		},
		{
			name: "WRAM mirror 5",
			args: args{
				pakAddr: 0xFF0000,
			},
			want: 0x7E0000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := PakAddressToBus(tt.args.pakAddr); got != tt.want {
				t.Errorf("PakAddressToBus() = 0x%06x, want 0x%06x", got, tt.want)
			}
		})
	}
}

func TestBusAddressToPak(t *testing.T) {
	type args struct {
		busAddr uint32
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		// WRAM banks 00..3F:
		{
			name: "WRAM bank $00:0000",
			args: args{
				busAddr: 0x000000,
			},
			want: 0xF50000,
		},
		{
			name: "WRAM bank $00:1FFF",
			args: args{
				busAddr: 0x001FFF,
			},
			want: 0xF51FFF,
		},
		{
			name: "WRAM bank $1F:0000",
			args: args{
				busAddr: 0x1F0000,
			},
			want: 0xF50000,
		},
		{
			name: "WRAM bank $1F:1FFF",
			args: args{
				busAddr: 0x1F1FFF,
			},
			want: 0xF51FFF,
		},
		{
			name: "WRAM bank $20:0000",
			args: args{
				busAddr: 0x200000,
			},
			want: 0xF50000,
		},
		{
			name: "WRAM bank $20:1FFF",
			args: args{
				busAddr: 0x201FFF,
			},
			want: 0xF51FFF,
		},
		{
			name: "WRAM bank $3F:0000",
			args: args{
				busAddr: 0x3F0000,
			},
			want: 0xF50000,
		},
		{
			name: "WRAM bank $3F:1FFF",
			args: args{
				busAddr: 0x3F1FFF,
			},
			want: 0xF51FFF,
		},
		{
			name: "ROM $00:8000",
			args: args{
				busAddr: 0x008000,
			},
			want: 0x000000,
		},
		{
			name: "ROM $00:FFFF",
			args: args{
				busAddr: 0x00FFFF,
			},
			want: 0x007FFF,
		},
		{
			name: "ROM $01:8000",
			args: args{
				busAddr: 0x018000,
			},
			want: 0x008000,
		},
		{
			name: "ROM $01:FFFF",
			args: args{
				busAddr: 0x01FFFF,
			},
			want: 0x00FFFF,
		},
		{
			name: "ROM $3E:8000",
			args: args{
				busAddr: 0x3E8000,
			},
			want: 0x1F0000,
		},
		{
			name: "ROM $3E:FFFF",
			args: args{
				busAddr: 0x3EFFFF,
			},
			want: 0x1F7FFF,
		},
		{
			name: "ROM $3F:8000",
			args: args{
				busAddr: 0x3F8000,
			},
			want: 0x1F8000,
		},
		{
			name: "ROM $3F:FFFF",
			args: args{
				busAddr: 0x3FFFFF,
			},
			want: 0x1FFFFF,
		},
		// WRAM banks 7E-7F:
		{
			name: "WRAM bank $7E:0000",
			args: args{
				busAddr: 0x7E0000,
			},
			want: 0xF50000,
		},
		{
			name: "WRAM bank $7E:1FFF",
			args: args{
				busAddr: 0x7E1FFF,
			},
			want: 0xF51FFF,
		},
		{
			name: "WRAM bank $7E:2000",
			args: args{
				busAddr: 0x7E2000,
			},
			want: 0xF52000,
		},
		{
			name: "WRAM bank $7E:3FFF",
			args: args{
				busAddr: 0x7E3FFF,
			},
			want: 0xF53FFF,
		},
		{
			name: "WRAM bank $7E:FFFF",
			args: args{
				busAddr: 0x7EFFFF,
			},
			want: 0xF5FFFF,
		},
		{
			name: "WRAM bank $7F:0000",
			args: args{
				busAddr: 0x7F0000,
			},
			want: 0xF60000,
		},
		{
			name: "WRAM bank $7F:FFFF",
			args: args{
				busAddr: 0x7FFFFF,
			},
			want: 0xF6FFFF,
		},
		// WRAM banks 80..BF:
		{
			name: "WRAM bank $80:0000",
			args: args{
				busAddr: 0x800000,
			},
			want: 0xF50000,
		},
		{
			name: "WRAM bank $80:1FFF",
			args: args{
				busAddr: 0x801FFF,
			},
			want: 0xF51FFF,
		},
		{
			name: "WRAM bank $9F:0000",
			args: args{
				busAddr: 0x9F0000,
			},
			want: 0xF50000,
		},
		{
			name: "WRAM bank $9F:1FFF",
			args: args{
				busAddr: 0x9F1FFF,
			},
			want: 0xF51FFF,
		},
		{
			name: "WRAM bank $A0:0000",
			args: args{
				busAddr: 0xA00000,
			},
			want: 0xF50000,
		},
		{
			name: "WRAM bank $A0:1FFF",
			args: args{
				busAddr: 0xA01FFF,
			},
			want: 0xF51FFF,
		},
		{
			name: "WRAM bank $BF:0000",
			args: args{
				busAddr: 0xBF0000,
			},
			want: 0xF50000,
		},
		{
			name: "WRAM bank $BF:1FFF",
			args: args{
				busAddr: 0xBF1FFF,
			},
			want: 0xF51FFF,
		},
		{
			name: "ROM $80:8000",
			args: args{
				busAddr: 0x808000,
			},
			want: 0x200000,
		},
		{
			name: "ROM $80:FFFF",
			args: args{
				busAddr: 0x80FFFF,
			},
			want: 0x207FFF,
		},
		{
			name: "ROM $81:8000",
			args: args{
				busAddr: 0x818000,
			},
			want: 0x208000,
		},
		{
			name: "ROM $81:FFFF",
			args: args{
				busAddr: 0x81FFFF,
			},
			want: 0x20FFFF,
		},
		{
			name: "ROM $BE:8000",
			args: args{
				busAddr: 0xBE8000,
			},
			want: 0x3F0000,
		},
		{
			name: "ROM $BE:FFFF",
			args: args{
				busAddr: 0xBEFFFF,
			},
			want: 0x3F7FFF,
		},
		{
			name: "ROM $BF:8000",
			args: args{
				busAddr: 0xBF8000,
			},
			want: 0x3F8000,
		},
		{
			name: "ROM $BF:FFFF",
			args: args{
				busAddr: 0xBFFFFF,
			},
			want: 0x3FFFFF,
		},
		// BW-RAM banks 40..43:
		{
			name: "BW-RAM $40:0000",
			args: args{
				busAddr: 0x400000,
			},
			want: 0xE00000,
		},
		{
			name: "BW-RAM $40:7FFF",
			args: args{
				busAddr: 0x407FFF,
			},
			want: 0xE07FFF,
		},
		{
			name: "BW-RAM $40:8000",
			args: args{
				busAddr: 0x408000,
			},
			want: 0xE08000,
		},
		{
			name: "BW-RAM $40:FFFF",
			args: args{
				busAddr: 0x40FFFF,
			},
			want: 0xE0FFFF,
		},
		{
			name: "BW-RAM $43:0000",
			args: args{
				busAddr: 0x430000,
			},
			want: 0xE30000,
		},
		{
			name: "BW-RAM $43:7FFF",
			args: args{
				busAddr: 0x437FFF,
			},
			want: 0xE37FFF,
		},
		{
			name: "BW-RAM $43:8000",
			args: args{
				busAddr: 0x438000,
			},
			want: 0xE38000,
		},
		{
			name: "BW-RAM $43:FFFF",
			args: args{
				busAddr: 0x43FFFF,
			},
			want: 0xE3FFFF,
		},
		// BW-RAM image banks 44..4F
		{
			name: "BW-RAM $44:0000",
			args: args{
				busAddr: 0x440000,
			},
			want: 0xE00000,
		},
		{
			name: "BW-RAM $44:1FFF",
			args: args{
				busAddr: 0x441FFF,
			},
			want: 0xE01FFF,
		},
		{
			name: "BW-RAM $44:2000",
			args: args{
				busAddr: 0x442000,
			},
			want: 0xE00000,
		},
		{
			name: "BW-RAM $44:7FFF",
			args: args{
				busAddr: 0x447FFF,
			},
			want: 0xE01FFF,
		},
		{
			name: "BW-RAM $45:0000",
			args: args{
				busAddr: 0x450000,
			},
			want: 0xE00000,
		},
		{
			name: "BW-RAM $45:7FFF",
			args: args{
				busAddr: 0x457FFF,
			},
			want: 0xE01FFF,
		},
		{
			name: "BW-RAM $4F:0000",
			args: args{
				busAddr: 0x4F0000,
			},
			want: 0xE00000,
		},
		{
			name: "BW-RAM $4F:FFFF",
			args: args{
				busAddr: 0x4FFFFF,
			},
			want: 0xE01FFF,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := BusAddressToPak(tt.args.busAddr); got != tt.want {
				t.Errorf("BusAddressToPak() = 0x%06x, want 0x%06x", got, tt.want)
			}
		})
	}
}
