# Change Log

## [0.0.2](https://github.com/ably/vcdiff-go/tree/v0.0.2)

[Full Changelog](https://github.com/ably/vcdiff-go/compare/v0.0.1...v0.0.2)

**Implemented enhancements:**
- Added GitHub Actions CI workflow for automated testing across Go versions 1.18-1.25
- Set minimum Go version requirement to 1.18 for broader compatibility

## [0.0.1](https://github.com/ably/vcdiff-go/tree/v0.0.1)

[Full Changelog](https://github.com/ably/vcdiff-go/releases/tag/v0.0.1)

**Implemented enhancements:**

- RFC 3284 compliant VCDIFF decoding with clean, idiomatic Go API
- Command-line tool for applying deltas and inspecting VCDIFF file structure
- Comprehensive validation with support for all VCDIFF instruction types (ADD, COPY, RUN)
- Efficient address caching implementation for optimized decoding performance
- Full Adler-32 checksum validation support (non-standard extension)
- Robust error handling with detailed error messages for debugging malformed files
- Extensive testing suite with 94 test cases and reference implementation validation
- Three CLI commands: `apply` for delta application, `parse` for structure inspection, and `analyze` for source context analysis

**Limitations:**

- Application headers are not supported in this implementation
- Secondary compression (e.g., gzip, bzip2) is not supported
- Compatibility is limited to VCDIFF deltas created using `xdelta3 -e -S -A` (no secondary compression, no application header)
