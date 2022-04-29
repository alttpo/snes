package memory

type ROM struct {
	data   []byte
	offset uint32
}

func NewROM(data []byte, offset uint32) *ROM {
	return &ROM{data, offset}
}

func (m *ROM) Read(address uint32) byte {
	return m.data[address-m.offset]
}

func (m *ROM) Write(address uint32, value byte) {
}

func (m *ROM) Shutdown() {
	panic("implement me")
}

func (m *ROM) Size() uint32 {
	return uint32(len(m.data))
}

func (m *ROM) Clear() {
	for i := range m.data {
		m.data[i] = 0
	}
}

func (m *ROM) Dump(address uint32) []byte {
	panic("implement me")
}
