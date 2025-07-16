package vcdiff

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidMagic    = errors.New("invalid VCDIFF magic bytes")
	ErrInvalidVersion  = errors.New("unsupported VCDIFF version")
	ErrInvalidFormat   = errors.New("invalid VCDIFF format")
	ErrCorruptedData   = errors.New("corrupted VCDIFF data")
	ErrInvalidChecksum = errors.New("invalid checksum")
)

// Enhanced error functions for detailed reporting
func errUnexpectedEOF(context string, bytesNeeded int) error {
	return fmt.Errorf("unexpected EOF while reading %s: need %d bytes", context, bytesNeeded)
}

func errDataOverrun(instruction string, offset int, needed int, available int) error {
	return fmt.Errorf("%s instruction at offset %d requires %d bytes but only %d available in data section",
		instruction, offset, needed, available)
}

func errInvalidValue(field string, offset int, value interface{}, reason string) error {
	return fmt.Errorf("invalid %s at offset %d: value %v, %s", field, offset, value, reason)
}

func errOutOfBounds(instruction string, address uint32, size uint32, maxBound uint32) error {
	return fmt.Errorf("%s instruction address %d + size %d exceeds bounds (max %d)",
		instruction, address, size, maxBound)
}

type Decoder interface {
	Decode(delta []byte) ([]byte, error)
}

type decoder struct {
	source []byte
}

func NewDecoder(source []byte) Decoder {
	return &decoder{
		source: source,
	}
}

func (d *decoder) Decode(delta []byte) ([]byte, error) {
	// Parse the delta to get structured information
	parsed, err := ParseDelta(delta)
	if err != nil {
		return nil, err
	}

	// Process all windows and accumulate target data
	target := make([]byte, 0)

	for _, window := range parsed.Windows {
		// Decode this window's target data
		windowTarget, err := d.decodeWindow(&window, d.source)
		if err != nil {
			return nil, err
		}

		// Append to overall target
		target = append(target, windowTarget...)
	}

	return target, nil
}

func Decode(source []byte, delta []byte) ([]byte, error) {
	decoder := NewDecoder(source)
	return decoder.Decode(delta)
}

// decodeWindow decodes a single window using the source data and window instructions
func (d *decoder) decodeWindow(window *Window, source []byte) ([]byte, error) {
	// Initialize address cache
	addressCache := NewAddressCache(NearCacheSize, SameCacheSize)
	addressCache.Reset(window.AddressSection)

	// Create target buffer
	target := make([]byte, 0, window.TargetWindowLength)

	// Get source segment for this window
	var sourceSegment []byte
	sourceLength := 0
	if window.WinIndicator&VCDSource != 0 {
		// Use source data
		start := window.SourceSegmentPosition
		end := start + window.SourceSegmentSize
		if end > uint32(len(source)) {
			return nil, ErrInvalidFormat
		}
		sourceSegment = source[start:end]
		sourceLength = len(sourceSegment)
	}

	// Parse and execute the actual instructions
	instructions, err := parseInstructions(window.InstructionSection, window.DataSection, addressCache)
	if err != nil {
		return nil, err
	}

	// Execute each instruction
	for _, instruction := range instructions {
		switch instruction.Type {
		case NoOp:
			// Skip
			continue

		case Add:
			// Add data from the instruction's data
			if len(instruction.Data) != int(instruction.Size) {
				return nil, ErrInvalidFormat
			}
			target = append(target, instruction.Data...)

		case Copy:
			// Decode the address using the address cache
			here := len(target) + sourceLength
			addr, err := addressCache.DecodeAddress(uint32(here), instruction.Mode)
			if err != nil {
				return nil, err
			}

			// Determine if copying from source or target
			if addr < uint32(sourceLength) {
				// Copy from source segment
				end := addr + instruction.Size
				if end > uint32(sourceLength) {
					return nil, errOutOfBounds("COPY", addr, instruction.Size, uint32(sourceLength))
				}
				target = append(target, sourceSegment[addr:end]...)
			} else {
				// Copy from target data (self-referential copy)
				targetAddr := addr - uint32(sourceLength)
				if targetAddr >= uint32(len(target)) {
					return nil, fmt.Errorf("COPY instruction address %d references target position %d but target only has %d bytes",
						addr, targetAddr, len(target))
				}

				// Handle overlapping copies byte by byte
				for i := uint32(0); i < instruction.Size; i++ {
					if targetAddr+i >= uint32(len(target)) {
						return nil, fmt.Errorf("COPY instruction would read beyond target bounds: position %d, target size %d",
							targetAddr+i, len(target))
					}
					target = append(target, target[targetAddr+i])
				}
			}

		case Run:
			// Repeat a single byte
			if len(instruction.Data) != 1 {
				return nil, ErrInvalidFormat
			}
			runByte := instruction.Data[0]
			for i := uint32(0); i < instruction.Size; i++ {
				target = append(target, runByte)
			}

		default:
			return nil, ErrInvalidFormat
		}
	}

	// Validate Adler32 checksum if present
	if window.HasChecksum {
		computed := ComputeChecksum(1, target) // Adler32 starts with initial value 1
		if computed != window.Checksum {
			return nil, fmt.Errorf("checksum validation failed: expected 0x%08x, got 0x%08x", window.Checksum, computed)
		}
	}

	return target, nil
}

// ParseDelta parses a VCDIFF delta and returns a structured representation
func ParseDelta(delta []byte) (*ParsedDelta, error) {
	if len(delta) < MinimumFileSize {
		return nil, ErrInvalidFormat
	}

	parsed := &ParsedDelta{}
	reader := bytes.NewReader(delta)

	if err := parseHeader(reader, &parsed.Header); err != nil {
		return nil, err
	}

	// Debug: check reader length after header

	for reader.Len() > 0 {
		window := Window{}
		if err := parseWindow(reader, &window); err != nil {
			if err == io.EOF {
				// If we still have bytes remaining but got EOF, the delta is malformed
				if reader.Len() > 0 {
					return nil, fmt.Errorf("malformed VCDIFF delta: %d bytes remain but cannot form valid window", reader.Len())
				}
				break
			}
			return nil, err
		}
		parsed.Windows = append(parsed.Windows, window)

		// Create address cache for this window
		addressCache := NewAddressCache(NearCacheSize, SameCacheSize)
		addressCache.Reset(window.AddressSection)

		// Parse instructions using the instruction section and data section
		instructions, err := parseInstructions(window.InstructionSection, window.DataSection, addressCache)
		if err != nil {
			return nil, err
		}
		parsed.Instructions = append(parsed.Instructions, instructions...)
	}

	return parsed, nil
}

// parseHeader parses the VCDIFF header section
func parseHeader(reader *bytes.Reader, header *Header) error {
	startPos := reader.Len()

	var magic [3]byte // Read 3 magic bytes as defined in RFC 3284
	n, err := reader.Read(magic[:])
	if err != nil {
		if err == io.EOF {
			return errUnexpectedEOF("VCDIFF magic bytes", 3-n)
		}
		return fmt.Errorf("error reading magic bytes at offset %d: %v", startPos-reader.Len(), err)
	}
	if n < 3 {
		return errUnexpectedEOF("VCDIFF magic bytes", 3-n)
	}

	// Compare magic bytes using bytes.Equal - RFC 3284 Section 4.1
	if !bytes.Equal(magic[:], VCDIFFMagic[:]) {
		return fmt.Errorf("invalid VCDIFF magic bytes at offset 0: expected %02x%02x%02x but got %02x%02x%02x",
			VCDIFFMagic[0], VCDIFFMagic[1], VCDIFFMagic[2], magic[0], magic[1], magic[2])
	}

	version, err := reader.ReadByte()
	if err != nil {
		if err == io.EOF {
			return errUnexpectedEOF("version byte", 1)
		}
		return fmt.Errorf("error reading version at offset 3: %v", err)
	}
	if version != VCDIFFVersion {
		return errInvalidValue("version", 3, version, fmt.Sprintf("only version %d is supported", VCDIFFVersion))
	}

	indicator, err := reader.ReadByte()
	if err != nil {
		if err == io.EOF {
			return errUnexpectedEOF("header indicator", 1)
		}
		return fmt.Errorf("error reading header indicator at offset 4: %v", err)
	}

	// Check for reserved bits in header indicator
	validHeaderBits := byte(VCDDecompress | VCDCodetable | VCDAppHeader)
	if indicator & ^validHeaderBits != 0 {
		return errInvalidValue("header indicator", 4, indicator, "reserved bits must be zero")
	}

	header.Magic = magic
	header.Version = version
	header.Indicator = indicator

	return nil
}

// parseWindow parses a single VCDIFF window
func parseWindow(reader *bytes.Reader, window *Window) error {
	if reader.Len() == 0 {
		return io.EOF
	}
	startLen := reader.Len()

	indicator, err := reader.ReadByte()
	if err != nil {
		if err == io.EOF {
			return errUnexpectedEOF("window indicator", 1)
		}
		return fmt.Errorf("error reading window indicator at offset %d: %v", startLen-reader.Len(), err)
	}

	// Check for reserved bits in window indicator
	validBits := byte(VCDSource | VCDTarget | VCDAdler32)
	if indicator & ^validBits != 0 {
		return errInvalidValue("window indicator", startLen-reader.Len()-1, indicator, "reserved bits must be zero")
	}

	window.WinIndicator = indicator

	if indicator&VCDSource != 0 {
		sourceSize, err := ReadVarint(reader)
		if err != nil {
			return err
		}
		window.SourceSegmentSize = sourceSize

		sourcePos, err := ReadVarint(reader)
		if err != nil {
			return err
		}
		window.SourceSegmentPosition = sourcePos
	}

	// Read the length of the delta encoding
	deltaSize, err := ReadVarint(reader)
	if err != nil {
		return err
	}
	window.DeltaEncodingLength = deltaSize

	// Read the delta encoding section - RFC 3284 Section 4.3
	deltaData := make([]byte, deltaSize)
	if _, err := reader.Read(deltaData); err != nil {
		return err
	}

	// Parse the delta encoding according to RFC 3284 Section 4.3
	deltaReader := bytes.NewReader(deltaData)

	// 1. Length of the target window
	targetSize, err := ReadVarint(deltaReader)
	if err != nil {
		return err
	}
	window.TargetWindowLength = targetSize

	// 2. Delta_Indicator byte
	deltaIndicator, err := deltaReader.ReadByte()
	if err != nil {
		return err
	}
	window.DeltaIndicator = deltaIndicator

	// 3. Length of data for ADDs and RUNs
	dataLength, err := ReadVarint(deltaReader)
	if err != nil {
		return err
	}
	window.DataSectionLength = dataLength

	// 4. Length of instructions section
	instructionLength, err := ReadVarint(deltaReader)
	if err != nil {
		return err
	}
	window.InstructionSectionLength = instructionLength

	// 5. Length of addresses for COPYs
	addressLength, err := ReadVarint(deltaReader)
	if err != nil {
		return err
	}
	window.AddressSectionLength = addressLength

	// Handle VCD_ADLER32 extension - checksum comes AFTER section lengths but BEFORE data sections
	if indicator&VCDAdler32 != 0 {
		window.HasChecksum = true
		// Read the 4-byte checksum from the delta encoding data
		checksumBytes := make([]byte, 4)
		if _, err := deltaReader.Read(checksumBytes); err != nil {
			return err
		}
		// Convert to uint32 (big-endian)
		window.Checksum = uint32(checksumBytes[0])<<24 |
			uint32(checksumBytes[1])<<16 |
			uint32(checksumBytes[2])<<8 |
			uint32(checksumBytes[3])
	}

	// 6. Data section for ADDs and RUNs
	window.DataSection = make([]byte, dataLength)
	if _, err := deltaReader.Read(window.DataSection); err != nil {
		return err
	}

	// 7. Instructions and sizes section
	window.InstructionSection = make([]byte, instructionLength)
	if _, err := deltaReader.Read(window.InstructionSection); err != nil {
		return err
	}

	// 8. Addresses section for COPYs
	window.AddressSection = make([]byte, addressLength)
	if addressLength > 0 {
		if _, err := deltaReader.Read(window.AddressSection); err != nil {
			return err
		}
	}

	return nil
}

// parseInstructions parses the instruction data from a window using the code table
func parseInstructions(instructionData []byte, dataSection []byte, addressCache *AddressCache) ([]RuntimeInstruction, error) {
	stream := bytes.NewReader(instructionData)
	var instructions []RuntimeInstruction
	dataIndex := 0
	instructionOffset := 0

	for {
		code, err := stream.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading instruction code at offset %d: %v", instructionOffset, err)
		}

		// Each code can have up to 2 instructions
		for slot := 0; slot < 2; slot++ {
			instruction := DefaultCodeTable.Get(code, slot)
			if instruction.Type == NoOp {
				continue
			}

			size := uint32(instruction.Size)
			if size == 0 && instruction.Type != NoOp {
				size, err = ReadVarint(stream)
				if err != nil {
					return nil, fmt.Errorf("error reading size for %s instruction at offset %d: %v",
						instruction.Type, instructionOffset, err)
				}
			}

			runtimeInst := RuntimeInstruction{
				Type: instruction.Type,
				Size: size,
				Mode: instruction.Mode,
			}

			// Handle instruction-specific data
			switch instruction.Type {
			case Add:
				if dataIndex+int(size) > len(dataSection) {
					return nil, errDataOverrun("ADD", instructionOffset, int(size), len(dataSection)-dataIndex)
				}
				runtimeInst.Data = make([]byte, size)
				copy(runtimeInst.Data, dataSection[dataIndex:dataIndex+int(size)])
				dataIndex += int(size)

			case Run:
				if dataIndex >= len(dataSection) {
					return nil, fmt.Errorf("RUN instruction at offset %d requires 1 byte but no data available in data section", instructionOffset)
				}
				runtimeInst.Data = []byte{dataSection[dataIndex]}
				dataIndex++

			case Copy:
				// Address will be decoded when needed during execution
				runtimeInst.Mode = instruction.Mode
			}

			instructions = append(instructions, runtimeInst)
		}
		instructionOffset++
	}

	return instructions, nil
}
