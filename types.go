package vcdiff

const (
	VCDIFFMagic1 = 0xD6
	VCDIFFMagic2 = 0xC3
	VCDIFFMagic3 = 0xC4
	VCDIFFVersion = 0x00
)

const (
	VCDDecompress = 0x01
	VCDCodetable  = 0x02
)

const (
	VCDSource = 0x01
	VCDTarget = 0x02
	VCDAdler32 = 0x04
)

const (
	VCDAdd = iota
	VCDCopy
	VCDRun
	VCDNoop
)

type Header struct {
	Magic     [3]byte
	Version   byte
	Indicator byte
}

type Window struct {
	Indicator     byte
	SourceSize    uint32
	SourcePos     uint32
	TargetSize    uint32
	DeltaSize     uint32
	DeltaData     []byte
	Checksum      uint32
	HasChecksum   bool
}

type Instruction struct {
	Type byte
	Size uint32
	Mode byte
	Addr uint32
	Data []byte
}

type AddressCache struct {
	Near []uint32
	Same []uint32
}

type InstructionTable struct {
	Entries [256]InstructionEntry
}

type InstructionEntry struct {
	Type1 byte
	Size1 byte
	Mode1 byte
	Type2 byte
	Size2 byte
	Mode2 byte
}

func NewAddressCache() *AddressCache {
	return &AddressCache{
		Near: make([]uint32, 4),
		Same: make([]uint32, 3*256),
	}
}