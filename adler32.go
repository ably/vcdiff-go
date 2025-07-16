package vcdiff

// Adler32 implements the Adler-32 checksum algorithm
type Adler32 struct{}

const (
	// Base for modulo arithmetic
	adler32Base = 65521
	// Number of iterations we can safely do before applying the modulo
	adler32NMax = 5552
)

// ComputeChecksum computes the Adler32 checksum for the given data
func ComputeChecksum(initial uint32, data []byte) uint32 {
	if len(data) == 0 {
		return initial
	}

	s1 := initial & 0xffff
	s2 := (initial >> 16) & 0xffff

	index := 0
	length := len(data)

	for length > 0 {
		k := length
		if k > adler32NMax {
			k = adler32NMax
		}
		length -= k

		for i := 0; i < k; i++ {
			s1 += uint32(data[index])
			s2 += s1
			index++
		}

		s1 %= adler32Base
		s2 %= adler32Base
	}

	return (s2 << 16) | s1
}
