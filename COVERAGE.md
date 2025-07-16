# VCDIFF Code Coverage

This document explains how to generate and analyze code coverage for the VCDIFF implementation.

## Quick Start

```bash
# Generate coverage report
./coverage.sh

# Or run manually
go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Coverage Commands

### Basic Coverage
```bash
# Run tests with coverage percentage
go test -cover

# Output: PASS coverage: 79.8% of statements
```

### Detailed Coverage
```bash
# Generate coverage profile
go test -coverprofile=coverage.out

# View function-level coverage
go tool cover -func=coverage.out

# Generate HTML report  
go tool cover -html=coverage.out -o coverage.html
```

### Advanced Analysis
```bash
# Show only functions with less than 100% coverage
go tool cover -func=coverage.out | grep -v '100.0%'

# Show only functions with less than 80% coverage  
go tool cover -func=coverage.out | awk '$3 < "80.0%" {print}'

# Count uncovered functions
go tool cover -func=coverage.out | grep -v '100.0%' | wc -l
```

## Coverage by Mode

### Atomic Mode (Recommended)
```bash
# Best for concurrent tests and fuzz tests
go test -coverprofile=coverage.out -covermode=atomic

# Atomic mode provides accurate coverage for concurrent execution
```

### Set Mode (Default)  
```bash
# Good for sequential tests
go test -coverprofile=coverage.out -covermode=set

# Set mode tracks whether each statement was executed at least once
```

### Count Mode
```bash  
# Detailed execution counts
go test -coverprofile=coverage.out -covermode=count

# Count mode tracks how many times each statement was executed
```

## Current Coverage Analysis

### Overall Coverage: **79.8%**

### Functions by Coverage Level:

#### ✅ **100% Coverage (11 functions)**
- `NewAddressCache`, `Reset`, `Update` - Address cache management
- `Get`, `BuildDefaultCodeTable` - Code table operations  
- `NewInstruction` - Instruction creation
- `errUnexpectedEOF`, `errDataOverrun`, `errInvalidValue` - Error helpers
- `NewDecoder`, `Decode` (standalone function) - Public API

#### 🟨 **90-99% Coverage (3 functions)**  
- `ReadVarint` (92.9%) - Varint decoding
- `ParseDelta` (95.5%) - Delta parsing
- `Decode` method (90.0%) - Decoder implementation

#### 🟧 **70-89% Coverage (4 functions)**
- `DecodeAddress` (77.4%) - Address decoding logic
- `decodeWindow` (77.3%) - Window processing  
- `parseWindow` (78.8%) - Window parsing
- `parseInstructions` (94.1%) - Instruction parsing

#### 🟥 **Below 70% Coverage (3 functions)**
- `parseHeader` (70.0%) - Header parsing
- `String` (66.7%) - Instruction type formatting
- `ComputeChecksum` (0.0%) - Adler32 checksum (unused)
- `errOutOfBounds` (0.0%) - Error helper (unused)

## Coverage Improvement Strategies

### 1. **Add Error Path Tests**
Many uncovered lines are error handling paths:
```bash
# Identify error paths with 0% coverage
go tool cover -func=coverage.out | grep '0.0%'
```

### 2. **Test Edge Cases**
Focus on functions with 70-90% coverage:
- Invalid header formats
- Malformed window structures  
- Address cache boundary conditions

### 3. **Add Feature Tests**
Unused features like Adler32 checksums need dedicated tests:
```go
func TestAdler32Checksum(t *testing.T) {
    // Test checksum computation and validation
}
```

### 4. **Fuzz Testing Coverage**
Fuzz tests can discover uncovered paths:
```bash
# Run fuzz tests and check coverage impact
go test -fuzz=FuzzDecode -fuzztime=30s
go test -coverprofile=fuzz_coverage.out
```

## Integration with CI/CD

### Coverage Thresholds
```bash
# Fail if coverage drops below 80%
go test -coverprofile=coverage.out
COVERAGE=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
if (( $(echo "$COVERAGE < 80" | bc -l) )); then
    echo "Coverage $COVERAGE% is below 80% threshold"
    exit 1
fi
```

### Coverage Reporting
```bash
# Generate coverage badge data
echo "coverage: ${COVERAGE}%" > coverage.txt

# Compare coverage with previous runs
diff coverage_baseline.txt coverage.txt
```

## Best Practices

1. **Target 85%+ overall coverage** - Current: 79.8%
2. **Critical paths should be 95%+** - Decode functions  
3. **Use atomic mode for accuracy** - Handles concurrent tests
4. **Focus on error paths** - Often missed in happy path testing
5. **Regular coverage monitoring** - Track trends over time
6. **Combine with fuzz testing** - Discovers edge cases

## Files to Focus On

Based on current analysis, prioritize improving coverage in:

1. **`vcdiff.go`** - Core decode logic (82.2% overall)
2. **`addresscache.go`** - Address handling (77.4% in DecodeAddress)  
3. **`varint.go`** - Varint parsing error paths (92.9%)

The goal should be to achieve **85%+ overall coverage** with **90%+ coverage on critical decode paths**.