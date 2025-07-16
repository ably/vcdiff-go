package vcdiff

// VCDIFF magic bytes and version - RFC 3284 Section 4.1
const (
	VCDIFFMagic1  = 0xD6 // First magic byte: 'V' with high bit set
	VCDIFFMagic2  = 0xC3 // Second magic byte: 'C' with high bit set
	VCDIFFMagic3  = 0xC4 // Third magic byte: 'D' with high bit set
	VCDIFFVersion = 0x00 // Version 0 as defined in RFC 3284
)

// VCDIFFMagic is the expected magic number sequence - RFC 3284 Section 4.1
var VCDIFFMagic = [3]byte{VCDIFFMagic1, VCDIFFMagic2, VCDIFFMagic3}

// Header indicator flags - RFC 3284 Section 4.1
const (
	VCDDecompress = 0x01 // VCD_DECOMPRESS: secondary compression used
	VCDCodetable  = 0x02 // VCD_CODETABLE: custom instruction table used
	VCDAppHeader  = 0x04 // VCD_APPHEADER: application header present
)

// Window indicator flags - RFC 3284 Section 4.2
const (
	VCDSource  = 0x01 // VCD_SOURCE: window uses source data
	VCDTarget  = 0x02 // VCD_TARGET: window uses target data
	VCDAdler32 = 0x04 // VCD_ADLER32: window includes Adler-32 checksum (non-standard extension)
)

// Variable-length integer encoding constants - RFC 3284 Section 2
const (
	VarintContinuationBit = 0x80 // High bit indicates continuation
	VarintValueMask       = 0x7F // Mask for 7-bit value portion
	VarintMaxShift        = 32   // Maximum shift to prevent overflow
	VarintShiftIncrement  = 7    // Bits to shift for each byte
)

// Instruction code ranges - RFC 3284 Section 5
const (
	RunInstructionMin  = 0   // RUN instructions: 0-17
	RunInstructionMax  = 17  // RUN instructions: 0-17
	AddInstructionMin  = 18  // ADD instructions: 18-161
	AddInstructionMax  = 161 // ADD instructions: 18-161
	CopyInstructionMin = 162 // COPY instructions: 162-255
	CopyInstructionMax = 255 // COPY instructions: 162-255
)

// Address cache configuration - RFC 3284 Section 5.3
const (
	NearCacheSize        = 4       // Size of "near" address cache
	SameCacheSize        = 3 * 256 // Size of "same" address cache
	InstructionTableSize = 256     // Size of instruction code table
)

// File format validation constants
const (
	MinimumFileSize = 4 // Minimum VCDIFF file size (magic + version)
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
	WinIndicator             byte   // Win_Indicator - RFC 3284 Section 4.2
	SourceSegmentSize        uint32 // Source segment size - RFC 3284 Section 4.2
	SourceSegmentPosition    uint32 // Source segment position - RFC 3284 Section 4.2
	TargetWindowLength       uint32 // Length of the target window - RFC 3284 Section 4.3
	DeltaEncodingLength      uint32 // Length of the delta encoding - RFC 3284 Section 4.3
	DeltaIndicator           byte   // Delta_Indicator - RFC 3284 Section 4.3
	DataSectionLength        uint32 // Length of data for ADDs and RUNs - RFC 3284 Section 4.3
	InstructionSectionLength uint32 // Length of instructions section - RFC 3284 Section 4.3
	AddressSectionLength     uint32 // Length of addresses for COPYs - RFC 3284 Section 4.3
	DataSection              []byte // Data section for ADDs and RUNs - RFC 3284 Section 4.3
	InstructionSection       []byte // Instructions and sizes section - RFC 3284 Section 4.3
	AddressSection           []byte // Addresses section for COPYs - RFC 3284 Section 4.3
	Checksum                 uint32 // Adler-32 checksum of target window (VCD_ADLER32 extension)
	HasChecksum              bool   // Whether VCD_ADLER32 bit is set in WinIndicator
}

// Legacy instruction type for backwards compatibility
type LegacyInstruction struct {
	Type byte
	Size uint32
	Mode byte
	Addr uint32
	Data []byte
}

type InstructionTable struct {
	Entries [InstructionTableSize]InstructionEntry
}

type InstructionEntry struct {
	Type1 byte
	Size1 byte
	Mode1 byte
	Type2 byte
	Size2 byte
	Mode2 byte
}

type ParsedDelta struct {
	Header       Header
	Windows      []Window
	Instructions []RuntimeInstruction
}
