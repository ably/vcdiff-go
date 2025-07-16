package vcdiff

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMetadata represents the metadata.json structure for test cases
type TestMetadata struct {
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	Category           string   `json:"category"`
	ExpectedBehavior   string   `json:"expected_behavior"`
	TestObjectives     []string `json:"test_objectives"`
	ExpectedErrorType  string   `json:"expected_error_type"`
	ExpectedProperties struct {
		SourceSize         int    `json:"source_size"`
		TargetSize         int    `json:"target_size"`
		HasChecksum        bool   `json:"has_checksum"`
		InstructionCount   int    `json:"instruction_count"`
		WindowCount        int    `json:"window_count"`
		PrimaryInstruction string `json:"primary_instruction"`
		ShouldFailFast     bool   `json:"should_fail_fast"`
		ErrorLocation      string `json:"error_location"`
	} `json:"expected_properties"`
}

// TestCase represents a single VCDIFF test case
type TestCase struct {
	Name          string
	TestDir       string
	SourceFile    string
	TargetFile    string
	DeltaFile     string
	MetadataFile  string
	Metadata      *TestMetadata
	ShouldSucceed bool
}

// discoverTestCases finds all test cases in a given category directory
func discoverTestCases(categoryDir string, shouldSucceed bool) ([]TestCase, error) {
	var testCases []TestCase

	err := filepath.WalkDir(categoryDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Look for directories containing delta.vcdiff
		if d.IsDir() {
			deltaFile := filepath.Join(path, "delta.vcdiff")
			if _, err := os.Stat(deltaFile); err == nil {
				// This is a test directory
				sourceFile := filepath.Join(path, "source")
				targetFile := filepath.Join(path, "target")
				metadataFile := filepath.Join(path, "metadata.json")

				// Check required files exist
				if _, err := os.Stat(sourceFile); err != nil {
					return fmt.Errorf("missing source file in %s", path)
				}
				if _, err := os.Stat(targetFile); err != nil {
					return fmt.Errorf("missing target file in %s", path)
				}

				// Load metadata if available
				var metadata *TestMetadata
				if metadataData, err := os.ReadFile(metadataFile); err == nil {
					metadata = &TestMetadata{}
					if err := json.Unmarshal(metadataData, metadata); err != nil {
						return fmt.Errorf("invalid metadata.json in %s: %v", path, err)
					}
				}

				// Generate test name from directory structure
				relPath, _ := filepath.Rel(categoryDir, path)
				testName := strings.ReplaceAll(relPath, string(filepath.Separator), "/")
				if metadata != nil && metadata.Name != "" {
					testName = fmt.Sprintf("%s (%s)", testName, metadata.Name)
				}

				testCases = append(testCases, TestCase{
					Name:          testName,
					TestDir:       path,
					SourceFile:    sourceFile,
					TargetFile:    targetFile,
					DeltaFile:     deltaFile,
					MetadataFile:  metadataFile,
					Metadata:      metadata,
					ShouldSucceed: shouldSucceed,
				})
			}
		}

		return nil
	})

	return testCases, err
}

// loadTestFile loads a test file and returns its contents
func loadTestFile(filepath string) ([]byte, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %v", filepath, err)
	}
	return data, nil
}

// TestTargetedPositive tests cases that should succeed in decoding
func TestTargetedPositive(t *testing.T) {
	testDir := "submodules/vcdiff-tests/targeted-positive"

	// Check if test directory exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Skipf("Test directory %s not found, skipping targeted positive tests", testDir)
		return
	}

	testCases, err := discoverTestCases(testDir, true)
	if err != nil {
		t.Fatalf("Failed to discover test cases: %v", err)
	}

	if len(testCases) == 0 {
		t.Skip("No targeted positive test cases found")
		return
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Load test files
			source, err := loadTestFile(tc.SourceFile)
			if err != nil {
				t.Fatalf("Failed to load source: %v", err)
			}

			target, err := loadTestFile(tc.TargetFile)
			if err != nil {
				t.Fatalf("Failed to load target: %v", err)
			}

			delta, err := loadTestFile(tc.DeltaFile)
			if err != nil {
				t.Fatalf("Failed to load delta: %v", err)
			}

			// Test decoding
			result, err := Decode(source, delta)
			if err != nil {
				t.Fatalf("Expected successful decode but got error: %v", err)
			}

			// Compare result with expected target
			if len(result) != len(target) {
				t.Fatalf("Result length mismatch: got %d bytes, expected %d bytes", len(result), len(target))
			}

			for i := range result {
				if result[i] != target[i] {
					t.Fatalf("Result differs from target at byte %d: got 0x%02x, expected 0x%02x", i, result[i], target[i])
				}
			}

			// Validate metadata expectations if available
			if tc.Metadata != nil && tc.Metadata.ExpectedProperties.TargetSize > 0 {
				if len(result) != tc.Metadata.ExpectedProperties.TargetSize {
					t.Errorf("Target size mismatch: got %d, expected %d", len(result), tc.Metadata.ExpectedProperties.TargetSize)
				}
			}
		})
	}
}

// TestTargetedNegative tests cases that should fail during decoding
func TestTargetedNegative(t *testing.T) {
	testDir := "submodules/vcdiff-tests/targeted-negative"

	// Check if test directory exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Skipf("Test directory %s not found, skipping targeted negative tests", testDir)
		return
	}

	testCases, err := discoverTestCases(testDir, false)
	if err != nil {
		t.Fatalf("Failed to discover test cases: %v", err)
	}

	if len(testCases) == 0 {
		t.Skip("No targeted negative test cases found")
		return
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Load test files
			source, err := loadTestFile(tc.SourceFile)
			if err != nil {
				t.Fatalf("Failed to load source: %v", err)
			}

			delta, err := loadTestFile(tc.DeltaFile)
			if err != nil {
				t.Fatalf("Failed to load delta: %v", err)
			}

			// Test decoding - should fail
			result, err := Decode(source, delta)
			if err == nil {
				t.Fatalf("Expected decode to fail but it succeeded, got result of %d bytes", len(result))
			}

			// Validate error type if specified in metadata
			if tc.Metadata != nil && tc.Metadata.ExpectedErrorType != "" {
				// This could be enhanced to check specific error types
				t.Logf("Got expected error: %v (type: %s)", err, tc.Metadata.ExpectedErrorType)
			}
		})
	}
}

// TestGeneralPositive tests general positive test cases with various content
func TestGeneralPositive(t *testing.T) {
	testDir := "submodules/vcdiff-tests/general-positive"

	// Check if test directory exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Skipf("Test directory %s not found, skipping general positive tests", testDir)
		return
	}

	testCases, err := discoverTestCases(testDir, true)
	if err != nil {
		t.Fatalf("Failed to discover test cases: %v", err)
	}

	if len(testCases) == 0 {
		t.Skip("No general positive test cases found")
		return
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Load test files
			source, err := loadTestFile(tc.SourceFile)
			if err != nil {
				t.Fatalf("Failed to load source: %v", err)
			}

			target, err := loadTestFile(tc.TargetFile)
			if err != nil {
				t.Fatalf("Failed to load target: %v", err)
			}

			delta, err := loadTestFile(tc.DeltaFile)
			if err != nil {
				t.Fatalf("Failed to load delta: %v", err)
			}

			// Test decoding
			result, err := Decode(source, delta)
			if err != nil {
				t.Fatalf("Expected successful decode but got error: %v", err)
			}

			// Compare result with expected target
			if len(result) != len(target) {
				t.Fatalf("Result length mismatch: got %d bytes, expected %d bytes", len(result), len(target))
			}

			for i := range result {
				if result[i] != target[i] {
					t.Fatalf("Result differs from target at byte %d: got 0x%02x, expected 0x%02x", i, result[i], target[i])
				}
			}
		})
	}
}

// TestFuzz tests fuzz test cases (corrupted inputs)
func TestFuzz(t *testing.T) {
	testDir := "submodules/vcdiff-tests/fuzz"

	// Check if test directory exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Log("No external fuzz test cases found")
		t.Log("To run built-in fuzz tests, use:")
		t.Log("  go test -fuzz=FuzzDecode -fuzztime=30s")
		t.Log("  go test -fuzz=FuzzReadVarint -fuzztime=30s")
		t.Log("  go test -fuzz=FuzzParseDelta -fuzztime=30s")
		t.Log("  go test -fuzz=FuzzAddressCache -fuzztime=30s")
		t.Log("  go test -fuzz=FuzzInstructionParsing -fuzztime=30s")
		t.Skip("Use built-in fuzz tests instead")
		return
	}

	testCases, err := discoverTestCases(testDir, false)
	if err != nil {
		t.Fatalf("Failed to discover test cases: %v", err)
	}

	if len(testCases) == 0 {
		t.Log("No external fuzz test cases found")
		t.Log("To run built-in fuzz tests, use:")
		t.Log("  go test -fuzz=FuzzDecode -fuzztime=30s")
		t.Log("  go test -fuzz=FuzzReadVarint -fuzztime=30s")
		t.Log("  go test -fuzz=FuzzParseDelta -fuzztime=30s")
		t.Log("  go test -fuzz=FuzzAddressCache -fuzztime=30s")
		t.Log("  go test -fuzz=FuzzInstructionParsing -fuzztime=30s")
		t.Skip("Use built-in fuzz tests instead")
		return
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Load test files
			source, err := loadTestFile(tc.SourceFile)
			if err != nil {
				t.Fatalf("Failed to load source: %v", err)
			}

			delta, err := loadTestFile(tc.DeltaFile)
			if err != nil {
				t.Fatalf("Failed to load delta: %v", err)
			}

			// Test decoding - should not crash (may succeed or fail)
			// The key requirement for fuzz tests is that the decoder doesn't panic
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("Decoder panicked on fuzz input: %v", r)
					}
				}()

				result, err := Decode(source, delta)
				if err != nil {
					t.Logf("Fuzz test failed as expected: %v", err)
				} else {
					t.Logf("Fuzz test unexpectedly succeeded, got %d bytes", len(result))
				}
			}()
		})
	}
}

// Legacy tests for basic functionality
func TestNewDecoder(t *testing.T) {
	source := []byte("hello world")
	decoder := NewDecoder(source)

	if decoder == nil {
		t.Fatal("NewDecoder returned nil")
	}
}

func TestDecode(t *testing.T) {
	source := []byte("hello world")
	// Use a valid empty-to-empty VCDIFF delta
	delta := []byte{0xd6, 0xc3, 0xc4, 0x00, 0x00, 0x04, 0x09, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}

	decoder := NewDecoder(source)
	result, err := decoder.Decode(delta)

	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Empty-to-empty should produce an empty result (not nil, but zero length)
	if result == nil {
		t.Fatal("Decode returned nil result")
	}

	if len(result) != 0 {
		t.Fatalf("Expected empty result, got %d bytes", len(result))
	}
}

func TestDecodeFunction(t *testing.T) {
	source := []byte("hello world")
	// Use a valid empty-to-empty VCDIFF delta
	delta := []byte{0xd6, 0xc3, 0xc4, 0x00, 0x00, 0x04, 0x09, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}

	result, err := Decode(source, delta)

	if err != nil {
		t.Fatalf("Decode function failed: %v", err)
	}

	// Empty-to-empty should produce an empty result (not nil, but zero length)
	if result == nil {
		t.Fatal("Decode function returned nil result")
	}

	if len(result) != 0 {
		t.Fatalf("Expected empty result, got %d bytes", len(result))
	}
}
