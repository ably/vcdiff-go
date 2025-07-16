package main

import (
	"bytes"
	"fmt"
	"log"

	"ably/vcdiff"
)

func main() {
	source := []byte("Hello, World!")
	
	delta := bytes.NewReader([]byte{})
	
	decoder := vcdiff.NewDecoder(source)
	result, err := decoder.Decode(delta)
	if err != nil {
		log.Fatalf("Failed to decode: %v", err)
	}
	
	fmt.Printf("Source: %q\n", source)
	fmt.Printf("Result: %q\n", result)
	
	result2, err := vcdiff.Decode(source, bytes.NewReader([]byte{}))
	if err != nil {
		log.Fatalf("Failed to decode with function: %v", err)
	}
	
	fmt.Printf("Result2: %q\n", result2)
}