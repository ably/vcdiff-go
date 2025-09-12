# VCDIFF Go Decoder

A Go implementation of a VCDIFF (RFC 3284) decoder library and command-line tool for efficient binary differencing and compression.

## Overview

This repository contains both a Go library and a command-line interface (CLI) for working with VCDIFF delta files. The library provides a VCDIFF decoder that can decode delta files created according to RFC 3284 - The VCDIFF Generic Differencing and Compression Data Format. VCDIFF is a format for expressing one data stream as a variant of another data stream, commonly used for binary differencing, compression, and patch applications.

The CLI tool can be used to apply VCDIFF deltas to reconstruct files, as well as to inspect and analyze the structure of VCDIFF delta files.

## Features

- **Go Library**: RFC 3284 compliant VCDIFF decoding with clean, idiomatic API
- **Command-Line Tool**: Apply deltas and inspect VCDIFF file structure
- **Comprehensive Validation**: Support for all VCDIFF instruction types (ADD, COPY, RUN)
- **Address Caching**: Efficient decoding with proper address cache implementation
- **Checksum Validation**: Full Adler-32 checksum validation support
- **Robust Error Handling**: Detailed error messages for debugging malformed files
- **Extensive Testing**: 94 test cases with reference implementation validation

## Limitations

- **Application Headers**: This implementation does not handle application header information
- **Secondary Compression**: This decoder does not support secondary compression (e.g., gzip, bzip2)
- **Compatibility**: Works with VCDIFF deltas created using `xdelta3 -e -S -A` (no secondary compression, no application header)

## Checksum Support

- **VCD_ADLER32**: This implementation detects and parses the VCD_ADLER32 extension (bit 0x04 in window indicator)
- **Non-standard Extension**: The Adler-32 checksum is not part of RFC 3284 but is supported by some implementations
- **Validation**: Full Adler-32 checksum validation is implemented and performed during decoding
- **Display**: Checksums are displayed in the CLI output as `Adler32: 0x########`

## Installation

```bash
go get github.com/ably/vcdiff-go
```

## Quick Start

### Library Usage

```go
package main

import (
    "fmt"
    "io/ioutil"
    "log"
    
    "github.com/ably/vcdiff-go"
)

func main() {
    // Read the source file
    source, err := ioutil.ReadFile("original.txt")
    if err != nil {
        log.Fatal(err)
    }
    
    // Read the VCDIFF delta file
    deltaData, err := ioutil.ReadFile("changes.vcdiff")
    if err != nil {
        log.Fatal(err)
    }
    
    // Apply the delta to reconstruct the target
    result, err := vcdiff.Decode(source, deltaData)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Decoded result: %s\n", result)
}
```

### CLI Usage

Build the CLI tool:

```bash
go build -o vcdiff ./cmd/vcdiff
```

Apply a VCDIFF delta:

```bash
./vcdiff apply -b source.txt -d changes.vcdiff -o result.txt
```

Inspect a VCDIFF delta file:

```bash
./vcdiff parse -d changes.vcdiff
```

Analyze a VCDIFF delta with source context:

```bash
./vcdiff analyze -b source.txt -d changes.vcdiff
```

## API Reference

### Core Functions

#### `vcdiff.Decode(source []byte, delta []byte) ([]byte, error)`

Decodes a VCDIFF delta file using the provided source data and returns the reconstructed target data.

**Parameters:**
- `source`: The original source data (may be empty for deltas that don't reference source)
- `delta`: The VCDIFF delta file data

**Returns:**
- Decoded target data as byte slice
- Error if decoding fails (malformed delta, checksum validation failure, etc.)

#### `vcdiff.NewDecoder(source []byte) Decoder`

Creates a new decoder instance with the specified source data. Useful for decoding multiple deltas against the same source.

**Parameters:**
- `source`: The source data for decoding operations

**Returns:**
- A `Decoder` interface that can be used to decode multiple deltas

#### `decoder.Decode(delta []byte) ([]byte, error)`

Decodes a single VCDIFF delta using the decoder's source data.

### Error Handling

The decoder provides detailed error messages for various failure conditions:
- Invalid VCDIFF format or magic bytes
- Malformed varint encoding
- Out-of-bounds memory access attempts
- Checksum validation failures
- Truncated or corrupted delta files

## Command-Line Interface

The CLI provides three main commands:

### `apply` - Apply VCDIFF Delta

Applies a VCDIFF delta to a source file to produce the target file.

```bash
./vcdiff apply -b <source-file> -d <delta-file> -o <output-file>
```

**Flags:**
- `-b, --base`: Source/base file path (required)
- `-d, --delta`: VCDIFF delta file path (required)
- `-o, --output`: Output file path (required)

### `parse` - Inspect VCDIFF Structure

Parses and displays the internal structure of a VCDIFF delta file.

```bash
./vcdiff parse -d <delta-file>
```

**Flags:**
- `-d, --delta`: VCDIFF delta file path (required)

**Output includes:**
- Header information (magic bytes, version, flags)
- Window details (source segments, target length, checksums)
- Instruction breakdown (ADD, COPY, RUN operations)
- Address cache usage
- Data section analysis

### `analyze` - Analyze with Source Context

Analyzes a VCDIFF delta file with access to the source data, providing additional insights.

```bash
./vcdiff analyze -b <source-file> -d <delta-file>
```

**Flags:**
- `-b, --base`: Source/base file path (required)
- `-d, --delta`: VCDIFF delta file path (required)

**Additional features:**
- Validates actual address references
- Shows source data context for COPY operations
- Provides compression ratio analysis

## License

This project is licensed under the Apache License 2.0. See the LICENSE file for details.

## References

- [RFC 3284: The VCDIFF Generic Differencing and Compression Data Format](https://tools.ietf.org/html/rfc3284)
- [xdelta3: VCDIFF binary diff tool](https://github.com/jmacd/xdelta)

