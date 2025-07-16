package vcdiff

import (
	"bytes"
	"testing"
)

// FuzzDecode tests the main Decode function with random inputs
func FuzzDecode(f *testing.F) {
	// Seed with known valid VCDIFF data
	f.Add([]byte("ABCDE"), []byte{0xd6, 0xc3, 0xc4, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00})
	f.Add([]byte(""), []byte{0xd6, 0xc3, 0xc4, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00})
	f.Add([]byte("TEST"), []byte{0xd6, 0xc3, 0xc4, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x04, 0x01, 0x00, 0x04, 0x01, 0x54, 0x45, 0x53, 0x54})

	// Seed with some malformed data that should be rejected
	f.Add([]byte("SOURCE"), []byte{0xff, 0xff, 0xff})       // Invalid magic
	f.Add([]byte("SOURCE"), []byte{0xd6, 0xc3, 0xc4})       // Truncated
	f.Add([]byte("SOURCE"), []byte{0xd6, 0xc3, 0xc4, 0x99}) // Invalid version

	f.Fuzz(func(t *testing.T, source []byte, delta []byte) {
		// The decoder should never panic, regardless of input
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Decode panicked with source len=%d, delta len=%d: %v", len(source), len(delta), r)
			}
		}()

		result, err := Decode(source, delta)

		// If decode succeeds, result should be valid
		if err == nil {
			// Basic sanity checks on successful decode
			if result == nil {
				t.Error("Decode returned nil result with nil error")
			}
			// Result length should be reasonable (not massive)
			if len(result) > 10*1024*1024 { // 10MB limit
				t.Errorf("Decode returned suspiciously large result: %d bytes", len(result))
			}
		}

		// If decode fails, error should be non-nil and descriptive
		if err != nil && len(err.Error()) == 0 {
			t.Error("Decode returned empty error message")
		}
	})
}

// FuzzReadVarint tests varint parsing with random byte sequences
func FuzzReadVarint(f *testing.F) {
	// Seed with valid varints
	f.Add([]byte{0x00})       // 0
	f.Add([]byte{0x7f})       // 127
	f.Add([]byte{0x80, 0x01}) // 128
	f.Add([]byte{0xff, 0x7f}) // Maximum 2-byte

	// Seed with invalid varints
	f.Add([]byte{0x80})                         // Incomplete
	f.Add([]byte{0x80, 0x80, 0x80, 0x80, 0x80}) // Too long
	f.Add([]byte{})                             // Empty

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ReadVarint panicked with data %v: %v", data, r)
			}
		}()

		reader := bytes.NewReader(data)
		result, err := ReadVarint(reader)

		if err == nil {
			// If successful, result should be reasonable
			if result > 0xFFFFFFFF {
				t.Errorf("ReadVarint returned value exceeding uint32: %d", result)
			}
		}
	})
}

// FuzzParseDelta tests the ParseDelta function with random inputs
func FuzzParseDelta(f *testing.F) {
	// Seed with minimal valid VCDIFF headers
	f.Add([]byte{0xd6, 0xc3, 0xc4, 0x00, 0x00})                                                 // Header only
	f.Add([]byte{0xd6, 0xc3, 0xc4, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00}) // Complete minimal

	// Seed with invalid data
	f.Add([]byte{0x00, 0x00, 0x00})
	f.Add([]byte{0xd6, 0xc3})

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ParseDelta panicked with data len=%d: %v", len(data), r)
			}
		}()

		parsed, err := ParseDelta(data)

		if err == nil && parsed == nil {
			t.Error("ParseDelta returned nil result with nil error")
		}

		if parsed != nil {
			// Sanity checks on parsed structure
			if len(parsed.Windows) > 1000 {
				t.Errorf("ParseDelta returned suspicious number of windows: %d", len(parsed.Windows))
			}
			if len(parsed.Instructions) > 10000 {
				t.Errorf("ParseDelta returned suspicious number of instructions: %d", len(parsed.Instructions))
			}
		}
	})
}

// FuzzAddressCache tests address cache operations
func FuzzAddressCache(f *testing.F) {
	// Seed with various address data and modes
	f.Add([]byte{0x00}, uint32(0), byte(0))          // Self mode
	f.Add([]byte{0x64}, uint32(100), byte(1))        // Near mode
	f.Add([]byte{0xff}, uint32(255), byte(8))        // Same mode
	f.Add([]byte{0x00}, uint32(0xFFFFFFFF), byte(9)) // Invalid mode

	f.Fuzz(func(t *testing.T, addressData []byte, here uint32, mode byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("AddressCache panicked with addressData=%v, here=%d, mode=%d: %v", addressData, here, mode, r)
			}
		}()

		cache := NewAddressCache(4, 3) // Standard cache sizes
		cache.Reset(addressData)

		// Test DecodeAddress - should not panic
		_, err := cache.DecodeAddress(here, mode)

		// Invalid modes should return errors, not panic
		if mode > 8 && err == nil {
			t.Errorf("DecodeAddress should reject invalid mode %d", mode)
		}
	})
}

// FuzzInstructionParsing tests instruction parsing with malformed data
func FuzzInstructionParsing(f *testing.F) {
	// Seed with various instruction patterns
	f.Add([]byte{0x01}, []byte{0x41}, []byte{}) // ADD instruction
	f.Add([]byte{0x00}, []byte{0x42}, []byte{}) // RUN instruction
	f.Add([]byte{0x13}, []byte{}, []byte{0x0A}) // COPY instruction
	f.Add([]byte{0xFF}, []byte{}, []byte{})     // Invalid instruction

	f.Fuzz(func(t *testing.T, instructionData []byte, dataSection []byte, addressSection []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("parseInstructions panicked: %v", r)
			}
		}()

		cache := NewAddressCache(4, 3)
		cache.Reset(addressSection)

		// This should not panic regardless of input
		_, err := parseInstructions(instructionData, dataSection, cache)

		// We don't care about the specific error, just that it doesn't panic
		_ = err
	})
}
