package cpualt

type BusReader = func(addr uint32) uint8
type BusWriter = func(addr uint32, val uint8)

type Bus struct {
	M uint8 // last data access

	// 2^10 because segments are 4bits length
	Read  [1048576]BusReader
	Write [1048576]BusWriter
}

func (b *Bus) Init() {
	for i := range b.Read {
		b.Read[i] = func(addr uint32) uint8 { return b.M }
		b.Write[i] = func(addr uint32, val uint8) {}
	}
}

func (b *Bus) AttachReader(start, end uint32, r BusReader) {
	aStart := start >> 4
	aEnd := end >> 4
	for a := aStart; a <= aEnd; a++ {
		b.Read[a] = r
	}
}

func (b *Bus) AttachWriter(start, end uint32, w BusWriter) {
	aStart := start >> 4
	aEnd := end >> 4
	for a := aStart; a <= aEnd; a++ {
		b.Write[a] = w
	}
}

func (b *Bus) EaRead(addr uint32) uint8 {
	b.M = b.Read[addr>>4](addr)
	return b.M
}

func (b *Bus) EaWrite(addr uint32, val uint8) {
	b.Write[addr>>4](addr, val)
	b.M = val
}

func (b *Bus) nWrite(bank byte, addr uint16, val byte) {
	ea := uint32(bank)<<16 | uint32(addr)

	b.Write[ea>>4](ea, val)
	b.M = val
}

func (b *Bus) nWrite16_cross(bank byte, addr uint16, value uint16) {
	ea := uint32(bank)<<16 | uint32(addr)
	ll := byte(value)
	hh := byte(value >> 8)
	b.Write[ea>>4](ea, ll)
	ea++
	b.Write[ea>>4](ea, hh)
	b.M = hh
}

func (b *Bus) eaWrite16_cross(ea uint32, value uint16) {
	ll := byte(value)
	hh := byte(value >> 8)
	b.Write[ea>>4](ea, ll)
	ea++
	b.Write[ea>>4](ea, hh)
	b.M = hh
}

func (b *Bus) nWrite16_wrap(bank byte, addr uint16, value uint16) {
	bank32 := uint32(bank) << 16
	ll := byte(value)
	hh := byte(value >> 8)

	ea := bank32 | uint32(addr)
	b.Write[ea>>4](ea, ll)
	ea = bank32 | uint32(addr+1)
	b.Write[ea>>4](ea, hh)
	b.M = hh
}

func (b *Bus) nRead(bank byte, addr uint16) uint8 {
	ea := uint32(bank)<<16 | uint32(addr)
	b.M = b.Read[ea>>4](ea)
	return b.M
}

func (b *Bus) nRead16_wrap(bank byte, addr uint16) uint16 {
	bank32 := uint32(bank) << 16

	ea := bank32 | uint32(addr)
	ll := b.Read[ea>>4](ea)
	ea = bank32 | uint32(addr+1)
	hh := b.Read[ea>>4](ea)
	b.M = hh

	return uint16(hh)<<8 | uint16(ll)
}

func (b *Bus) nRead16_cross(bank byte, addr uint16) uint16 {
	ea := uint32(bank)<<16 | uint32(addr)
	ll := b.Read[ea>>4](ea)
	ea = (ea + 1) & 0x00ffffff // wrap on 24bits
	hh := b.Read[ea>>4](ea)
	b.M = hh

	return uint16(hh)<<8 | uint16(ll)
}

func (b *Bus) eaRead16_cross(ea uint32) uint16 {
	ll := b.Read[ea>>4](ea)
	ea = (ea + 1) & 0x00ffffff // wrap on 24bits
	hh := b.Read[ea>>4](ea)
	b.M = hh

	return uint16(hh)<<8 | uint16(ll)
}

func (b *Bus) nRead24_wrap(bank byte, addr uint16) uint32 {
	bank32 := uint32(bank) << 16

	ea := bank32 | uint32(addr+0)
	ll := b.Read[ea>>4](ea)
	ea = bank32 | uint32(addr+1)
	mm := b.Read[ea>>4](ea)
	ea = bank32 | uint32(addr+2)
	hh := b.Read[ea>>4](ea)
	b.M = hh

	return uint32(hh)<<16 | uint32(mm)<<8 | uint32(ll)
}
