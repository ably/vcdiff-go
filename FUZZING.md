# VCDIFF Fuzz Testing

This project includes comprehensive fuzz tests to find edge cases and potential crashes in the VCDIFF implementation.

## Available Fuzz Tests

### FuzzDecode
Tests the main `Decode()` function with random source and delta inputs.
- **Coverage**: Full decode pipeline including header parsing, window processing, and instruction execution
- **Purpose**: Find crashes in the main API with malformed VCDIFF data

### FuzzReadVarint  
Tests varint parsing with random byte sequences.
- **Coverage**: Varint decoding edge cases, overflow protection, truncated data
- **Purpose**: Ensure varint parser never panics regardless of input

### FuzzParseDelta
Tests VCDIFF structure parsing with completely random data.
- **Coverage**: Header validation, window parsing, section parsing
- **Purpose**: Find parser crashes with malformed VCDIFF structures

### FuzzAddressCache
Tests address cache operations with random inputs.
- **Coverage**: Address decoding, cache management, mode validation
- **Purpose**: Ensure address cache operations are robust

### FuzzInstructionParsing
Tests instruction parsing with malformed instruction data.
- **Coverage**: Code table lookups, instruction validation, data section handling  
- **Purpose**: Find crashes in instruction processing

## Running Fuzz Tests

### Individual Tests
```bash
# Run specific fuzz test for 30 seconds
go test -fuzz=FuzzDecode -fuzztime=30s

# Run with shorter duration for quick testing
go test -fuzz=FuzzReadVarint -fuzztime=5s

# Run until first failure is found
go test -fuzz=FuzzParseDelta -fuzztime=1s
```

### Running Multiple Tests
```bash
# Note: Go fuzz testing requires running one test at a time
# To run all tests, run each individually:

go test -fuzz=FuzzDecode -fuzztime=10s
go test -fuzz=FuzzReadVarint -fuzztime=10s  
go test -fuzz=FuzzParseDelta -fuzztime=10s
go test -fuzz=FuzzAddressCache -fuzztime=10s
go test -fuzz=FuzzInstructionParsing -fuzztime=10s
```

## Understanding Fuzz Output

```
fuzz: elapsed: 3s, execs: 128144 (42714/sec), new interesting: 8 (total: 66)
```

- **elapsed**: Time spent fuzzing
- **execs**: Total test cases executed  
- **execs/sec**: Execution rate (higher is better)
- **new interesting**: New code paths discovered this period
- **total**: Total unique code paths covered

## Crash Investigation

If a fuzz test finds a crash:

1. **Reproducing**: The fuzzer automatically saves crashing inputs to `testdata/fuzz/`
2. **Debugging**: Use `go test -fuzz=<TestName> -fuzztime=1s` to reproduce quickly
3. **Fix**: Analyze the crashing input and add appropriate validation/error handling

## Seed Data Strategy

Each fuzz test includes carefully chosen seed data:
- **Valid inputs**: Known good VCDIFF data to ensure baseline coverage
- **Invalid inputs**: Known bad data that should be rejected gracefully
- **Edge cases**: Boundary conditions like empty inputs, maximum values

## Performance Notes

- **FuzzReadVarint**: Fastest (~300k-1M execs/sec) - simple varint parsing
- **FuzzDecode**: Slowest (~40k-100k execs/sec) - full decode pipeline
- **Memory usage**: Fuzzing is memory-intensive; monitor system resources for long runs

## Best Practices

1. **Regular fuzzing**: Run fuzz tests regularly during development
2. **CI integration**: Consider adding fuzz tests to continuous integration
3. **Duration**: Use longer durations (5-10+ minutes) for thorough testing
4. **Coverage**: Monitor "new interesting" count - should plateau after sufficient time
5. **Regression**: Keep crash-triggering inputs as regression tests

## Integration with Regular Tests

The existing `TestFuzz` function provides guidance on running fuzz tests:

```bash
go test -v -run TestFuzz
```

This shows available fuzz test commands when no external fuzz data is found.

## Quick Start

```bash
# See fuzz test usage instructions
go test -v -run TestFuzz

# Run a specific fuzz test for 30 seconds
go test -fuzz=FuzzDecode -fuzztime=30s

# Run a quick test for 5 seconds
go test -fuzz=FuzzReadVarint -fuzztime=5s
```