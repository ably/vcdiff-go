package vcdiff

import (
	"bytes"
	"testing"
)

func TestNewDecoder(t *testing.T) {
	source := []byte("hello world")
	decoder := NewDecoder(source)
	
	if decoder == nil {
		t.Fatal("NewDecoder returned nil")
	}
}

func TestDecode(t *testing.T) {
	source := []byte("hello world")
	delta := bytes.NewReader([]byte{})
	
	decoder := NewDecoder(source)
	result, err := decoder.Decode(delta)
	
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	
	if result == nil {
		t.Fatal("Decode returned nil result")
	}
}

func TestDecodeFunction(t *testing.T) {
	source := []byte("hello world")
	delta := bytes.NewReader([]byte{})
	
	result, err := Decode(source, delta)
	
	if err != nil {
		t.Fatalf("Decode function failed: %v", err)
	}
	
	if result == nil {
		t.Fatal("Decode function returned nil result")
	}
}