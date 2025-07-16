package vcdiff

import (
	"bytes"
	"fmt"
	"io"
)

// ReadVarint reads a variable-length integer as defined in RFC 3284 Section 2
// Follows the same algorithm as the C# MiscUtil reference implementation
func ReadVarint(reader *bytes.Reader) (uint32, error) {
	var result uint32
	startLen := reader.Len()

	for i := 0; i < 5; i++ { // Maximum 5 bytes for 32-bit integer
		b, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				bytesRead := startLen - reader.Len()
				return 0, fmt.Errorf("unexpected EOF while reading varint at offset %d: expected continuation or termination byte", bytesRead)
			}
			return 0, err
		}

		// Shift previous result left by 7 bits and add the new 7-bit value
		// This matches the C# reference: ret = (ret << 7) | (b&0x7f);
		result = (result << 7) | uint32(b&VarintValueMask)

		// Check if continuation bit is clear (end of varint)
		if b&VarintContinuationBit == 0 {
			return result, nil
		}
	}

	// If we've read 5 bytes without finding the end, the data is invalid
	startOffset := startLen - reader.Len() - 5
	return 0, fmt.Errorf("invalid varint at offset %d: exceeds maximum 5-byte encoding (continuation bit never cleared)", startOffset)
}
