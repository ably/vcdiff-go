package vcdiff

import (
	"errors"
	"io"
)

var (
	ErrInvalidMagic    = errors.New("invalid VCDIFF magic bytes")
	ErrInvalidVersion  = errors.New("unsupported VCDIFF version")
	ErrInvalidFormat   = errors.New("invalid VCDIFF format")
	ErrCorruptedData   = errors.New("corrupted VCDIFF data")
	ErrInvalidChecksum = errors.New("invalid checksum")
)

type Decoder interface {
	Decode(delta io.Reader) ([]byte, error)
}

type decoder struct {
	source []byte
}

func NewDecoder(source []byte) Decoder {
	return &decoder{
		source: source,
	}
}

func (d *decoder) Decode(delta io.Reader) ([]byte, error) {
	return nil, nil
}

func Decode(source []byte, delta io.Reader) ([]byte, error) {
	decoder := NewDecoder(source)
	return decoder.Decode(delta)
}