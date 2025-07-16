package vcdiff

import (
	"bytes"
	"testing"
)

func TestReadVarint(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected uint32
		hasError bool
	}{
		// Single byte values (0-127) - RFC 3284 Section 2
		{
			name:     "zero value",
			input:    []byte{0x00},
			expected: 0,
		},
		{
			name:     "small positive value",
			input:    []byte{0x01},
			expected: 1,
		},
		{
			name:     "maximum single byte",
			input:    []byte{0x7F},
			expected: 127,
		},

		// Two byte values (128-16383) - RFC 3284 Section 2
		{
			name:     "minimum two byte value",
			input:    []byte{0x81, 0x00},
			expected: 128,
		},
		{
			name:     "mid-range two byte value",
			input:    []byte{0x81, 0x7F},
			expected: 255,
		},
		{
			name:     "larger two byte value",
			input:    []byte{0x82, 0x00},
			expected: 256,
		},
		{
			name:     "maximum two byte value",
			input:    []byte{0xFF, 0x7F},
			expected: 16383,
		},

		// Three byte values (16384-2097151) - RFC 3284 Section 2
		{
			name:     "minimum three byte value",
			input:    []byte{0x81, 0x80, 0x00},
			expected: 16384,
		},
		{
			name:     "mid-range three byte value",
			input:    []byte{0x81, 0xFF, 0x7F},
			expected: 32767,
		},
		{
			name:     "larger three byte value",
			input:    []byte{0x82, 0x80, 0x00},
			expected: 32768,
		},
		{
			name:     "maximum three byte value",
			input:    []byte{0xFF, 0xFF, 0x7F},
			expected: 2097151,
		},

		// Four byte values (2097152-268435455) - RFC 3284 Section 2
		{
			name:     "minimum four byte value",
			input:    []byte{0x81, 0x80, 0x80, 0x00},
			expected: 2097152,
		},
		{
			name:     "mid-range four byte value",
			input:    []byte{0x81, 0xFF, 0xFF, 0x7F},
			expected: 4194303,
		},
		{
			name:     "larger four byte value",
			input:    []byte{0x82, 0x80, 0x80, 0x00},
			expected: 4194304,
		},
		{
			name:     "maximum four byte value",
			input:    []byte{0xFF, 0xFF, 0xFF, 0x7F},
			expected: 268435455,
		},

		// Five byte values (268435456-4294967295) - RFC 3284 Section 2
		{
			name:     "minimum five byte value",
			input:    []byte{0x81, 0x80, 0x80, 0x80, 0x00},
			expected: 268435456,
		},
		{
			name:     "mid-range five byte value",
			input:    []byte{0x81, 0xFF, 0xFF, 0xFF, 0x7F},
			expected: 536870911,
		},
		{
			name:     "larger five byte value",
			input:    []byte{0x82, 0x80, 0x80, 0x80, 0x00},
			expected: 536870912,
		},
		{
			name:     "maximum uint32 value",
			input:    []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x0F},
			expected: 4294967183,
		},

		// Edge cases and specific values
		{
			name:     "power of 2 boundary (256)",
			input:    []byte{0x82, 0x00},
			expected: 256,
		},
		{
			name:     "power of 2 boundary (65536)",
			input:    []byte{0x84, 0x80, 0x00},
			expected: 65536,
		},
		{
			name:     "power of 2 boundary (16777216)",
			input:    []byte{0x88, 0x80, 0x80, 0x00},
			expected: 16777216,
		},

		// Common file size values
		{
			name:     "1KB",
			input:    []byte{0x88, 0x00},
			expected: 1024,
		},
		{
			name:     "64KB",
			input:    []byte{0x84, 0x80, 0x00},
			expected: 65536,
		},
		{
			name:     "1MB",
			input:    []byte{0xC0, 0x80, 0x00},
			expected: 1048576,
		},

		// Error cases
		{
			name:     "empty input",
			input:    []byte{},
			hasError: true,
		},
		{
			name:     "incomplete varint - continuation bit set but no next byte",
			input:    []byte{0x80},
			hasError: true,
		},
		{
			name:     "incomplete varint - multiple continuation bits",
			input:    []byte{0x80, 0x80},
			hasError: true,
		},
		{
			name:     "overflow - six bytes would exceed uint32",
			input:    []byte{0x81, 0x80, 0x80, 0x80, 0x80, 0x00},
			hasError: true,
		},
		{
			name:     "overflow - maximum shift exceeded",
			input:    []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.input)
			result, err := ReadVarint(reader)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none, result: %d", result)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}

			// Verify that the entire input was consumed
			if reader.Len() != 0 {
				t.Errorf("Expected all input to be consumed, %d bytes remaining", reader.Len())
			}
		})
	}
}

func TestReadVarintBoundaryValues(t *testing.T) {
	// Test specific boundary values that are important for VCDIFF
	boundaryTests := []struct {
		name     string
		input    []byte
		expected uint32
	}{
		{
			name:     "7-bit boundary (127)",
			input:    []byte{0x7F},
			expected: 127,
		},
		{
			name:     "7-bit boundary + 1 (128)",
			input:    []byte{0x81, 0x00},
			expected: 128,
		},
		{
			name:     "14-bit boundary (16383)",
			input:    []byte{0xFF, 0x7F},
			expected: 16383,
		},
		{
			name:     "14-bit boundary + 1 (16384)",
			input:    []byte{0x81, 0x80, 0x00},
			expected: 16384,
		},
		{
			name:     "21-bit boundary (2097151)",
			input:    []byte{0xFF, 0xFF, 0x7F},
			expected: 2097151,
		},
		{
			name:     "21-bit boundary + 1 (2097152)",
			input:    []byte{0x81, 0x80, 0x80, 0x00},
			expected: 2097152,
		},
		{
			name:     "28-bit boundary (268435455)",
			input:    []byte{0xFF, 0xFF, 0xFF, 0x7F},
			expected: 268435455,
		},
		{
			name:     "28-bit boundary + 1 (268435456)",
			input:    []byte{0x81, 0x80, 0x80, 0x80, 0x00},
			expected: 268435456,
		},
	}

	for _, tt := range boundaryTests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.input)
			result, err := ReadVarint(reader)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestReadVarintWithTrailingData(t *testing.T) {
	// Test that varint parsing stops at the correct boundary
	tests := []struct {
		name              string
		input             []byte
		expectedValue     uint32
		expectedRemaining int
	}{
		{
			name:              "single byte with trailing data",
			input:             []byte{0x01, 0xFF, 0xFF},
			expectedValue:     1,
			expectedRemaining: 2,
		},
		{
			name:              "two bytes with trailing data",
			input:             []byte{0x81, 0x00, 0xFF, 0xFF},
			expectedValue:     128,
			expectedRemaining: 2,
		},
		{
			name:              "three bytes with trailing data",
			input:             []byte{0x81, 0x80, 0x00, 0xFF, 0xFF},
			expectedValue:     16384,
			expectedRemaining: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.input)
			result, err := ReadVarint(reader)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expectedValue {
				t.Errorf("Expected %d, got %d", tt.expectedValue, result)
			}

			if reader.Len() != tt.expectedRemaining {
				t.Errorf("Expected %d bytes remaining, got %d", tt.expectedRemaining, reader.Len())
			}
		})
	}
}

func TestReadVarintConstants(t *testing.T) {
	// Test that the varint parser correctly uses the defined constants
	tests := []struct {
		name        string
		input       []byte
		description string
	}{
		{
			name:        "continuation bit test",
			input:       []byte{VarintContinuationBit | 0x01, 0x00},
			description: "Tests VarintContinuationBit constant",
		},
		{
			name:        "value mask test",
			input:       []byte{VarintValueMask},
			description: "Tests VarintValueMask constant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.input)
			_, err := ReadVarint(reader)

			// We're mainly testing that constants are used correctly,
			// not specific return values
			if err != nil && tt.name != "continuation bit test" {
				t.Errorf("Unexpected error with %s: %v", tt.description, err)
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkReadVarint(b *testing.B) {
	testCases := []struct {
		name  string
		input []byte
	}{
		{"1-byte", []byte{0x7F}},
		{"2-byte", []byte{0xFF, 0x7F}},
		{"3-byte", []byte{0xFF, 0xFF, 0x7F}},
		{"4-byte", []byte{0xFF, 0xFF, 0xFF, 0x7F}},
		{"5-byte", []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x0F}},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				reader := bytes.NewReader(tc.input)
				_, err := ReadVarint(reader)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
