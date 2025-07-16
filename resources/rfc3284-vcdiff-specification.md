# RFC 3284 - The VCDIFF Generic Differencing and Compression Data Format

**Network Working Group**  
**Request for Comments: 3284**  
**Category: Standards Track**  
**June 2002**

## Authors
- D. Korn (AT&T Labs)
- J. MacDonald (UC Berkeley)
- J. Mogul (Hewlett-Packard Company)
- K. Vo (AT&T Labs)

## Status of this Memo
This document specifies an Internet standards track protocol for the Internet community, and requests discussion and suggestions for improvements.

## Abstract
The VCDIFF Generic Differencing and Compression Data Format is a general, efficient and portable data format suitable for encoding compressed and/or differencing data so that they can be easily transported among computers.

## 1. Introduction

VCDIFF is a format and an algorithm for delta encoding. Delta encoding is a way of expressing one data stream as a variant of another data stream. The format is designed to be portable across different machine architectures and to support efficient decoding.

### Key Features:
- **Portability**: Works across different computer architectures
- **Efficiency**: Provides compact representation and fast decoding
- **Flexibility**: Supports various encoding algorithms while maintaining consistent decoding

## 2. Data Model

VCDIFF treats compression as a special case of differencing. The format partitions large files into non-overlapping "target windows" that can be encoded independently.

### 2.1 Basic Concepts

**Source Data**: The reference data used for differencing
**Target Data**: The data being encoded as differences from the source
**Window**: A contiguous segment of the target data
**Delta**: The encoded representation of differences

### 2.2 Three Primary Delta Instructions

1. **ADD**: Copy a sequence of bytes from the delta encoding
2. **COPY**: Copy a substring from the source data or previously decoded target data
3. **RUN**: Repeat a specific byte multiple times

## 3. Format Structure

### 3.1 Overall Structure
A VCDIFF delta file consists of:
- Header
- One or more windows

### 3.2 Header Format
- Magic bytes: 0xD6, 0xC3, 0xC4 (VCD in ASCII with high bit set)
- Version byte: 0x00 for this specification
- Indicator byte: Flags for compression and application-specific data

### 3.3 Window Format
Each window contains:
- Window indicator
- Source segment size and position (if applicable)
- Target window size
- Delta encoding size
- Delta encoding data

## 4. Encoding Algorithms

### 4.1 Variable-Sized Integer Encoding
VCDIFF uses a portable variable-sized integer encoding:
- 7 bits per byte for the integer value
- High bit indicates continuation
- Little-endian byte order

### 4.2 Address Encoding
VCDIFF implements sophisticated address caching:
- **Near cache**: Recently used addresses
- **Same cache**: Addresses used with specific here modes
- Multiple encoding modes for optimal compression

### 4.3 Instruction Code Table
- 256-entry table defining instruction combinations
- Default table provided in specification
- Allows application-specific optimizations
- Supports single and paired instruction encodings

## 5. Decoding Algorithm

### 5.1 Overview
Decoding is performed in linear time with working space proportional to window size.

### 5.2 Process
1. Read window header
2. Initialize address caches
3. Process delta instructions sequentially
4. Apply ADD, COPY, and RUN operations
5. Verify target window checksum (if present)

## 6. Performance Characteristics

### 6.1 Compression Efficiency
- Comparable to gzip for similar content
- Excellent for files with common subsequences
- Efficient delta encoding for version control

### 6.2 Decoding Speed
- Linear time complexity O(n)
- Minimal memory requirements
- Fast random access to encoded data

## 7. Secondary Compression

VCDIFF supports optional secondary compression applied to:
- Delta encoding data
- Instruction sequences
- Address encodings

Common secondary compressors include:
- gzip
- bzip2
- Application-specific algorithms

## 8. Application Examples

### 8.1 HTTP Delta Encoding
Used in RFC 3229 "Delta encoding in HTTP" for efficient web content transmission.

### 8.2 Version Control Systems
Efficient storage of file versions with minimal space overhead.

### 8.3 Software Updates
Compact representation of software patches and updates.

## 9. Security Considerations

### 9.1 Decoder Vulnerabilities
- Malformed delta files could cause buffer overflows
- Implementations must validate all input parameters
- Careful bounds checking required

### 9.2 Resource Consumption
- Large window sizes could exhaust memory
- Implementations should limit resource usage
- Consider timeout mechanisms for long decoding operations

## 10. Implementation Guidelines

### 10.1 Encoder Considerations
- String matching algorithms (e.g., suffix arrays, hash tables)
- Window size optimization
- Address cache management
- Secondary compression integration

### 10.2 Decoder Requirements
- Robust error handling
- Efficient memory management
- Compliance with format specifications
- Performance optimization

## 11. IANA Considerations

The VCDIFF format defines:
- Magic number registration
- MIME type: application/vcdiff
- File extension: .vcdiff

## 12. References

### Normative References
- RFC 2119: Key words for use in RFCs
- RFC 1950: ZLIB Compressed Data Format
- RFC 1951: DEFLATE Compressed Data Format

### Informative References
- Bentley, J. and McIlroy, D. "Data Compression Using Long Common Strings"
- Various compression and differencing algorithms

## Appendices

### Appendix A: Default Instruction Code Table
[Detailed 256-entry table with instruction combinations]

### Appendix B: Interoperability Test Suite
[Test cases for implementation verification]

### Appendix C: Sample Implementation
[Reference implementation guidelines]

---

## Technical Details Summary

### Data Types
- **Integers**: Variable-sized, 7-bit encoding
- **Addresses**: Cached with near/same modes
- **Instructions**: Encoded via code table

### Limits
- Maximum window size: Implementation dependent
- Maximum file size: No theoretical limit
- Address cache sizes: Configurable

### Error Handling
- Checksum verification
- Bounds checking
- Format validation

### Performance
- **Encoding**: O(n) to O(n²) depending on algorithm
- **Decoding**: O(n) linear time
- **Memory**: Proportional to window size

---

This specification provides a complete framework for implementing VCDIFF encoders and decoders with consistent interoperability across different systems and applications.