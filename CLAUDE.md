# VCDIFF Go Decoder - Development Guidelines

## Coding Standards

### Numeric Values
- **Never use hard-coded numeric literals in the code**
- **All numeric values must be defined as named constants**
- **All constants must include comments indicating their origin** (RFC section, specification, etc.)
- Example:
  ```go
  // VCDIFF magic bytes - RFC 3284 Section 4.1
  const (
      VCDIFFMagic1 = 0xD6 // First magic byte: 'V' with high bit set
      VCDIFFMagic2 = 0xC3 // Second magic byte: 'C' with high bit set
      VCDIFFMagic3 = 0xC4 // Third magic byte: 'D' with high bit set
  )
  ```

## Implementation Status
- **Parser**: Window structure parsing is complete and working correctly
- **Decoder**: VCDIFF delta decoding is fully implemented with all instruction types (ADD, COPY, RUN)
- **Checksum**: Adler32 checksum validation is fully implemented and working
- **Error Handling**: Comprehensive validation with detailed error messages for malformed inputs
- **Testing**: Extensive test coverage including positive/negative tests and fuzz testing
- **CLI**: Uses Cobra framework with proper subcommands, flags, and help text

## CLI Commands
- **apply**: Apply VCDIFF delta to base document (flags: -b/--base, -d/--delta, -o/--output)
- **parse**: Parse and display VCDIFF delta structure (flags: -d/--delta)
- **analyze**: Analyze VCDIFF delta with base document context (flags: -b/--base, -d/--delta)
- **completion**: Generate shell completion scripts (bash, zsh, fish, powershell)
- **help**: Built-in help system with detailed usage information

## Test Setup
- Use `xdelta3 -e -S -A` to generate compatible test files
- VCDIFF test suite is included as git submodule in `submodules/vcdiff-tests/`
- Run comprehensive test suite: `cd submodules/vcdiff-tests && ./run_tests.sh ../../vcdiff`
- Module name: `github.com/ably/vcdiff-go`
- Copyright holder: Ably Realtime Limited

## Key Limitations
- Application headers not supported
- Secondary compression not supported
- Custom code tables not supported

## Build & Test Commands
- **Build CLI**: `go build -o vcdiff ./cmd/vcdiff`
- **Run all tests**: `go test ./...`
- **Run with coverage**: `./coverage.sh`
- **Run fuzz tests**: `go test -fuzz=.`
- **Lint/typecheck**: Run these commands to ensure code quality before committing

## Development Notes
- All 57 positive tests pass
- All 37 negative tests pass with proper error validation
- Current code coverage: 84.8%
- Fuzz testing implemented for robustness
- Error messages include specific context and suggestions for debugging
- Project structure follows Go conventions (source files at root, package name `vcdiff`)

## Repository Structure
```
github.com/ably/vcdiff-go/
├── go.mod (module: github.com/ably/vcdiff-go)
├── *.go (package vcdiff)
├── cmd/vcdiff/ (CLI tool)
├── submodules/vcdiff-tests/ (test suite submodule)
└── coverage.sh, FUZZING.md, COVERAGE.md
```