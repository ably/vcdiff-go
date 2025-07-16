package vcdiff

import (
	"bytes"
	"fmt"
)

const (
	SelfMode = 0
	HereMode = 1
)

// AddressCache manages address encoding/decoding for COPY instructions
type AddressCache struct {
	nearSize      int
	sameSize      int
	near          []uint32
	nextNearSlot  int
	same          []uint32
	addressStream *bytes.Reader
}

// NewAddressCache creates a new address cache with the specified sizes
func NewAddressCache(nearSize, sameSize int) *AddressCache {
	return &AddressCache{
		nearSize: nearSize,
		sameSize: sameSize,
		near:     make([]uint32, nearSize),
		same:     make([]uint32, sameSize*256),
	}
}

// Reset resets the address cache for a new window
func (ac *AddressCache) Reset(addresses []byte) {
	ac.nextNearSlot = 0

	// Clear near cache
	for i := range ac.near {
		ac.near[i] = 0
	}

	// Clear same cache
	for i := range ac.same {
		ac.same[i] = 0
	}

	ac.addressStream = bytes.NewReader(addresses)
}

// DecodeAddress decodes an address using the specified mode
func (ac *AddressCache) DecodeAddress(here uint32, mode byte) (uint32, error) {
	var addr uint32
	var err error

	// Validate addressing mode
	if mode > 8 {
		return 0, fmt.Errorf("invalid address cache mode %d: valid modes are 0-8", mode)
	}

	switch mode {
	case SelfMode:
		addr, err = ReadVarint(ac.addressStream)
		if err != nil {
			return 0, fmt.Errorf("error reading address for SELF mode: %v", err)
		}

	case HereMode:
		offset, err := ReadVarint(ac.addressStream)
		if err != nil {
			return 0, fmt.Errorf("error reading offset for HERE mode: %v", err)
		}
		if offset > here {
			return 0, fmt.Errorf("HERE mode offset %d exceeds current position %d", offset, here)
		}
		addr = here - offset

	default:
		if int(mode-2) < ac.nearSize {
			// Near cache
			cacheIndex := mode - 2
			if ac.near[cacheIndex] == 0 {
				return 0, fmt.Errorf("near cache slot %d is uninitialized", cacheIndex)
			}
			offset, err := ReadVarint(ac.addressStream)
			if err != nil {
				return 0, fmt.Errorf("error reading offset for near cache mode %d: %v", mode, err)
			}
			addr = ac.near[cacheIndex] + offset
		} else {
			// Same cache
			m := int(mode) - (2 + ac.nearSize)
			if m >= ac.sameSize {
				return 0, fmt.Errorf("same cache mode %d exceeds available slots (max %d)", mode, 2+ac.nearSize+ac.sameSize-1)
			}
			b, err := ac.addressStream.ReadByte()
			if err != nil {
				return 0, err
			}
			addr = ac.same[m*256+int(b)]
		}
	}

	ac.Update(addr)
	return addr, nil
}

// Update updates the address cache with a new address
func (ac *AddressCache) Update(address uint32) {
	if ac.nearSize > 0 {
		ac.near[ac.nextNearSlot] = address
		ac.nextNearSlot = (ac.nextNearSlot + 1) % ac.nearSize
	}

	if ac.sameSize > 0 {
		ac.same[address%(uint32(ac.sameSize)*256)] = address
	}
}
