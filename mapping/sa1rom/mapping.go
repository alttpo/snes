// Package sa1rom
//
// This implementation has a caveat for SA-1 ROM access via SNES A-bus.
//
// SA-1 ROM mapping on SNES A-bus is dynamic, and we don't have a clean way to alter the mapping temporarily
// and reset it back. Using the SNES A-bus is generally a bad idea to access dynamically mapped memory.
//
// A compromise is to just assume CX, DX, EX, FX are linearly mapped to banks $00..3F of linear ROM.
//
// CX : SNES A-bus banks $00..1F
// DX : SNES A-bus banks $20..3F
// EX : SNES A-bus banks $80..9F
// FX : SNES A-bus banks $A0..BF
package sa1rom

import (
	"github.com/alttpo/snes/mapping/util"
)

// https://archive.org/details/SNESDevManual/book2/page/n15/mode/1up?view=theater

func BusAddressToPak(busAddr uint32) (pakAddr uint32, err error) {
	bank := busAddr >> 16
	offs := busAddr & 0xFFFF

	if bank >= 0xC0 {
		// C0..FF
		// ROM area CX, DX, EX, FX:
		pakAddr = (bank-0xC0)<<16 | offs
		return
	} else if bank >= 0x80 {
		// 80..BF
		if offs >= 0x8000 {
			// ROM for CX, DX:
			pakAddr = ((bank - 0x80 + 0x40) << 15) | (offs & 0x7FFF)
		} else if offs >= 0x6000 {
			// BW-RAM image dynamically selects a single $2000 sized block:
			pakAddr = 0xE00000 | (offs - 0x6000)
		} else if offs < 0x2000 {
			// WRAM
			pakAddr = 0xF50000 | offs
		} else {
			// SA-1 I-RAM or registers
			err = util.ErrUnmappedAddress
		}
		return
	} else if bank >= 0x7E {
		// 7E..7F
		// WRAM access:
		pakAddr = (busAddr - 0x7E0000) + 0xF50000
		return
	} else if bank >= 0x50 {
		// inaccessible?
		err = util.ErrUnmappedAddress
	} else if bank >= 0x44 {
		// 44..4F
		// BW-RAM image dynamically selects a single $2000 sized block:
		abs := (busAddr - 0x440000) & 0x1FFF
		pakAddr = 0xE00000 | abs
		return
	} else if bank >= 0x40 {
		// 40..43
		// BW-RAM area: linearly mapped
		pakAddr = (bank-0x40+0xE0)<<16 | offs
		return
	} else if bank < 0x40 {
		// 00..3F
		if offs >= 0x8000 {
			// ROM for CX, DX:
			pakAddr = (bank << 15) | (offs & 0x7FFF)
		} else if offs >= 0x6000 {
			// BW-RAM image dynamically selects a single $2000 sized block:
			pakAddr = 0xE00000 | (offs - 0x6000)
		} else if offs < 0x2000 {
			// WRAM
			pakAddr = 0xF50000 | offs
		} else {
			// SA-1 I-RAM or registers
			err = util.ErrUnmappedAddress
		}
		return
	}
	return 0, util.ErrUnmappedAddress
}

func PakAddressToBus(pakAddr uint32) (busAddr uint32, err error) {
	if pakAddr >= 0xF50000 {
		// WRAM is easy:
		// mirror fxpakpro banks $F7..FF back down into WRAM $F5..F6 because these banks in FX Pak Pro space
		// are not available on the SNES bus; they are copies of otherwise inaccessible memory
		// like VRAM, CGRAM, OAM, etc.:
		busAddr = ((pakAddr - 0xF50000) & 0x01FFFF) + 0x7E0000
		return
	} else if pakAddr >= 0xE00000 && pakAddr < 0xF00000 {
		// BW-RAM is available up to 2Mbits in size at SNES A-bus banks $40..43:
		abs := pakAddr - 0xE00000
		// mask to only allow for 4 full banks, mirroring the higher banks to the first 4:
		bank := (abs >> 16) & 0x03
		offs := abs & 0xFFFF

		busAddr = ((0x40 + bank) << 16) | offs
		return
	} else if pakAddr < 0xE00000 {
		// ROM access:

		// Mirror the top $40..DF fxpakpro banks down to $00..3F because the SNES A-bus can only map in $40 full banks
		// of ROM:
		abs := pakAddr & 0x3FFFFF
		bank := abs >> 15
		offs := abs & 0x7FFF

		if bank >= 0x40 {
			// EX and FX map to SNES A-bus banks $80 and $A0:
			busAddr = (0x80+(bank-0x40))<<16 | (0x8000 + offs)
		} else {
			// CX and DX map to SNES A-bus banks $00 and $20:
			busAddr = bank<<16 | (0x8000 + offs)
		}

		return
	}
	return 0, util.ErrUnmappedAddress
}
