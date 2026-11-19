package emulator

// Joypads is a Device whose controller inputs the caller programs: Input[i] is what the auto joypad read registers
// JOY1..JOY4 ($4218..$421F) read, in their bit order (B Y Select Start Up Down Left Right in the high byte, A X L R in
// the low byte's top nibble), and what the serial ports JOYSER0/1 ($4016/$4017) shift out after a strobe (controllers 1
// and 2 on bit 0, 3 and 4 on bit 1).
type Joypads struct {
	Input [4]uint16

	strobe bool
	shift  [2]int
}

func (j *Joypads) Attach(s *System) error {
	s.MapIO(0x4016, 0x4017, j.readSerial, j.writeSerial)
	s.MapIO(0x4218, 0x421F, j.readAuto, nil)
	return nil
}

func (j *Joypads) Reset() {
	j.strobe, j.shift = false, [2]int{}
}

func (j *Joypads) readAuto(offs uint16) byte {
	i := (offs - 0x4218) >> 1
	if offs&1 == 0 {
		return byte(j.Input[i])
	}
	return byte(j.Input[i] >> 8)
}

func (j *Joypads) writeSerial(offs uint16, value byte) {
	if offs != 0x4016 {
		return
	}
	j.strobe = value&1 != 0
	if j.strobe {
		j.shift = [2]int{}
	}
}

// serialBit returns the serial bit n (0 = B, first) of the controller input in, 1 past the 16 bits.
func serialBit(in uint16, n int) byte {
	if n >= 16 {
		return 1
	}
	return byte(in >> (15 - n) & 1)
}

func (j *Joypads) readSerial(offs uint16) byte {
	port := int(offs - 0x4016)
	n := j.shift[port]
	v := serialBit(j.Input[port], n) | serialBit(j.Input[port+2], n)<<1
	if !j.strobe && n < 16 {
		j.shift[port]++
	}
	if port == 1 {
		// JOYSER1 bits 2..4 read as set:
		v |= 0x1C
	}
	return v
}

// Math is a Device for the CPU's multiplier and divider: WRMPYA/WRMPYB ($4202/$4203) multiply into RDMPY
// ($4216/$4217), and WRDIVL/H/WRDIVB ($4204..$4206) divide into RDDIV ($4214/$4215, the quotient) and RDMPY (the
// remainder). Results are ready at once, rather than after the 8 or 16 cycles the hardware takes.
type Math struct {
	mpya       byte
	divd       uint16
	quot, prod uint16
}

func (m *Math) Attach(s *System) error {
	s.MapIO(0x4202, 0x4206, nil, m.write)
	s.MapIO(0x4214, 0x4217, m.read, nil)
	return nil
}

func (m *Math) write(offs uint16, value byte) {
	switch offs {
	case 0x4202:
		m.mpya = value
	case 0x4203:
		m.prod = uint16(m.mpya) * uint16(value)
	case 0x4204:
		m.divd = m.divd&0xFF00 | uint16(value)
	case 0x4205:
		m.divd = m.divd&0x00FF | uint16(value)<<8
	case 0x4206:
		if value == 0 {
			m.quot, m.prod = 0xFFFF, m.divd
		} else {
			m.quot, m.prod = m.divd/uint16(value), m.divd%uint16(value)
		}
	}
}

func (m *Math) read(offs uint16) byte {
	switch offs {
	case 0x4214:
		return byte(m.quot)
	case 0x4215:
		return byte(m.quot >> 8)
	case 0x4216:
		return byte(m.prod)
	default:
		return byte(m.prod >> 8)
	}
}
