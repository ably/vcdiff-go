package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	vcdiff "github.com/ably/vcdiff-go"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vcdiff",
	Short: "VCDIFF CLI Tool",
	Long: `A command-line tool for working with VCDIFF (RFC 3284) delta files.

VCDIFF is a format for expressing one data stream as a variant of another data stream,
commonly used for binary differencing, compression, and patch applications.`,
	Version: "1.0.0",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(parseCmd)
	rootCmd.AddCommand(analyzeCmd)
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a VCDIFF delta to a base document",
	Long: `Apply a VCDIFF delta to a base document to produce the target document.

The base document is the original file, and the delta contains the changes
needed to transform it into the target document.`,
	Example: `  vcdiff apply -base old.txt -delta patch.vcdiff -output new.txt
  vcdiff apply -base old.txt -delta patch.vcdiff  # Output to stdout`,
	RunE: runApply,
}

var (
	applyBaseFile   string
	applyDeltaFile  string
	applyOutputFile string
)

func init() {
	applyCmd.Flags().StringVarP(&applyBaseFile, "base", "b", "", "Path to base document file")
	applyCmd.Flags().StringVarP(&applyDeltaFile, "delta", "d", "", "Path to VCDIFF delta file")
	applyCmd.Flags().StringVarP(&applyOutputFile, "output", "o", "", "Path to output file (default: stdout)")

	// Mark required flags
	applyCmd.MarkFlagRequired("base")
	applyCmd.MarkFlagRequired("delta")
}

func runApply(cmd *cobra.Command, args []string) error {
	baseData, err := os.ReadFile(applyBaseFile)
	if err != nil {
		return fmt.Errorf("error reading base file: %w", err)
	}

	deltaData, err := os.ReadFile(applyDeltaFile)
	if err != nil {
		return fmt.Errorf("error reading delta file: %w", err)
	}

	result, err := vcdiff.Decode(baseData, deltaData)
	if err != nil {
		return fmt.Errorf("error applying delta: %w", err)
	}

	var output io.Writer = os.Stdout
	if applyOutputFile != "" {
		file, err := os.Create(applyOutputFile)
		if err != nil {
			return fmt.Errorf("error creating output file: %w", err)
		}
		defer file.Close()
		output = file
	}

	if _, err := output.Write(result); err != nil {
		return fmt.Errorf("error writing output: %w", err)
	}

	return nil
}

var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "Parse a VCDIFF delta and show human-readable representation",
	Long: `Parse a VCDIFF delta file and display its contents in a human-readable format.

This command shows the VCDIFF header information, window details, and
instruction sequences contained in the delta file.`,
	Example: `  vcdiff parse -delta patch.vcdiff
  vcdiff parse -d patch.vcdiff  # Short form`,
	RunE: runParse,
}

var parseDeltaFile string

func init() {
	parseCmd.Flags().StringVarP(&parseDeltaFile, "delta", "d", "", "Path to VCDIFF delta file")
	parseCmd.MarkFlagRequired("delta")
}

func runParse(cmd *cobra.Command, args []string) error {
	deltaData, err := os.ReadFile(parseDeltaFile)
	if err != nil {
		return fmt.Errorf("error reading delta file: %w", err)
	}

	parsed, err := vcdiff.ParseDelta(deltaData)
	if err != nil {
		return fmt.Errorf("error parsing delta: %w", err)
	}

	printDelta(parsed)
	fmt.Println()

	if err := printInstructions(parsed, os.Stdout); err != nil {
		return fmt.Errorf("error printing instructions: %w", err)
	}

	return nil
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze a VCDIFF delta with base document context",
	Long: `Analyze a VCDIFF delta file with access to the base document to provide
detailed information about the instructions and referenced data.

This command shows the same information as 'parse' but also includes
hexdump-style output of the actual data chunks referenced by COPY instructions.`,
	Example: `  vcdiff analyze -base old.txt -delta patch.vcdiff
  vcdiff analyze -b old.txt -d patch.vcdiff  # Short form`,
	RunE: runAnalyze,
}

var (
	analyzeBaseFile  string
	analyzeDeltaFile string
)

func init() {
	analyzeCmd.Flags().StringVarP(&analyzeBaseFile, "base", "b", "", "Path to base document file")
	analyzeCmd.Flags().StringVarP(&analyzeDeltaFile, "delta", "d", "", "Path to VCDIFF delta file")

	// Mark required flags
	analyzeCmd.MarkFlagRequired("base")
	analyzeCmd.MarkFlagRequired("delta")
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	baseData, err := os.ReadFile(analyzeBaseFile)
	if err != nil {
		return fmt.Errorf("error reading base file: %w", err)
	}

	deltaData, err := os.ReadFile(analyzeDeltaFile)
	if err != nil {
		return fmt.Errorf("error reading delta file: %w", err)
	}

	parsed, err := vcdiff.ParseDelta(deltaData)
	if err != nil {
		return fmt.Errorf("error parsing delta: %w", err)
	}

	printDelta(parsed)
	fmt.Println()

	if err := printDetailedInstructions(parsed, baseData, os.Stdout); err != nil {
		return fmt.Errorf("error printing detailed instructions: %w", err)
	}

	return nil
}

func printDelta(parsed *vcdiff.ParsedDelta) {
	printHeader(&parsed.Header)
	fmt.Printf("  Windows:   %d\n", len(parsed.Windows))

	for i, window := range parsed.Windows {
		fmt.Printf("  Window %d:\n", i)
		printWindow(&window)
	}
}

func printHeader(header *vcdiff.Header) {
	fmt.Printf("VCDIFF Header:\n")
	fmt.Printf("  Magic:     0x%02x 0x%02x 0x%02x\n",
		header.Magic[0], header.Magic[1], header.Magic[2])
	fmt.Printf("  Version:   0x%02x\n", header.Version)
	fmt.Printf("  Indicator: 0x%02x", header.Indicator)
	if header.Indicator != 0 {
		fmt.Printf(" (")
		var flags []string
		if header.Indicator&vcdiff.VCDDecompress != 0 {
			flags = append(flags, "VCD_DECOMPRESS")
		}
		if header.Indicator&vcdiff.VCDCodetable != 0 {
			flags = append(flags, "VCD_CODETABLE")
		}
		if header.Indicator&vcdiff.VCDAppHeader != 0 {
			flags = append(flags, "VCD_APPHEADER")
		}
		for i, flag := range flags {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", flag)
		}
		fmt.Printf(")")
	}
	fmt.Printf("\n")
}

func printWindow(window *vcdiff.Window) {
	fmt.Printf("    WinIndicator:   0x%02x", window.WinIndicator)
	if window.WinIndicator != 0 {
		fmt.Printf(" (")
		var flags []string
		if window.WinIndicator&vcdiff.VCDSource != 0 {
			flags = append(flags, "VCD_SOURCE")
		}
		if window.WinIndicator&vcdiff.VCDTarget != 0 {
			flags = append(flags, "VCD_TARGET")
		}
		if window.WinIndicator&vcdiff.VCDAdler32 != 0 {
			flags = append(flags, "VCD_ADLER32")
		}
		for j, flag := range flags {
			if j > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", flag)
		}
		fmt.Printf(")")
	}
	fmt.Printf("\n")
	fmt.Printf("    SourceSegmentSize:  0x%x (%d)\n", window.SourceSegmentSize, window.SourceSegmentSize)
	fmt.Printf("    SourceSegmentPosition:   0x%x (%d)\n", window.SourceSegmentPosition, window.SourceSegmentPosition)
	fmt.Printf("    TargetWindowLength:  0x%x (%d)\n", window.TargetWindowLength, window.TargetWindowLength)
	fmt.Printf("    DeltaEncodingLength: 0x%x (%d)\n", window.DeltaEncodingLength, window.DeltaEncodingLength)
	fmt.Printf("    DeltaIndicator: 0x%02x\n", window.DeltaIndicator)
	fmt.Printf("    DataSectionLength: 0x%x (%d)\n", window.DataSectionLength, window.DataSectionLength)
	fmt.Printf("    InstructionSectionLength: 0x%x (%d)\n", window.InstructionSectionLength, window.InstructionSectionLength)
	fmt.Printf("    AddressSectionLength: 0x%x (%d)\n", window.AddressSectionLength, window.AddressSectionLength)
	if window.HasChecksum {
		fmt.Printf("    Adler32:     0x%08x\n", window.Checksum)
	}
}

func printDetailedInstructions(parsed *vcdiff.ParsedDelta, baseData []byte, w io.Writer) error {
	fmt.Fprintf(w, "Instructions with Data Context:\n")
	fmt.Fprintf(w, "===============================\n\n")

	for i, instruction := range parsed.Instructions {
		fmt.Fprintf(w, "Instruction %d:\n", i+1)

		var instType string
		switch instruction.Type {
		case vcdiff.Add:
			instType = "ADD"
		case vcdiff.Copy:
			instType = "COPY"
		case vcdiff.Run:
			instType = "RUN"
		case vcdiff.NoOp:
			instType = "NOOP"
		default:
			instType = fmt.Sprintf("UNK(%02x)", instruction.Type)
		}

		fmt.Fprintf(w, "  Type: %s\n", instType)
		fmt.Fprintf(w, "  Mode: 0x%02x\n", instruction.Mode)
		fmt.Fprintf(w, "  Size: 0x%x (%d bytes)\n", instruction.Size, instruction.Size)

		if instruction.Type == vcdiff.Copy {
			fmt.Fprintf(w, "  Addr: 0x%x (%d)\n", instruction.Addr, instruction.Addr)

			if instruction.Addr < uint32(len(baseData)) {
				endAddr := instruction.Addr + instruction.Size
				if endAddr > uint32(len(baseData)) {
					endAddr = uint32(len(baseData))
				}

				fmt.Fprintf(w, "  Data from base [0x%x:0x%x]:\n", instruction.Addr, endAddr)
				printHexDump(baseData[instruction.Addr:endAddr], w, int(instruction.Addr))
			} else {
				fmt.Fprintf(w, "  Data: <address out of bounds>\n")
			}
		} else if len(instruction.Data) > 0 {
			fmt.Fprintf(w, "  Data:\n")
			printHexDump(instruction.Data, w, 0)
		}

		fmt.Fprintf(w, "\n")
	}

	return nil
}

func printHexDump(data []byte, w io.Writer, baseOffset int) {
	const bytesPerLine = 16

	for i := 0; i < len(data); i += bytesPerLine {
		end := i + bytesPerLine
		if end > len(data) {
			end = len(data)
		}

		line := data[i:end]

		fmt.Fprintf(w, "    %08x  ", baseOffset+i)

		for j := 0; j < bytesPerLine; j++ {
			if j < len(line) {
				fmt.Fprintf(w, "%02x ", line[j])
			} else {
				fmt.Fprintf(w, "   ")
			}

			if j == 7 {
				fmt.Fprintf(w, " ")
			}
		}

		fmt.Fprintf(w, " |")
		for j := 0; j < len(line); j++ {
			if line[j] >= 32 && line[j] <= 126 {
				fmt.Fprintf(w, "%c", line[j])
			} else {
				fmt.Fprintf(w, ".")
			}
		}

		fmt.Fprintf(w, "|\n")
	}
}

func printInstructions(parsed *vcdiff.ParsedDelta, w io.Writer) error {
	fmt.Fprintf(w, "  Offset Code Type1 Size1  @Addr1 + Type2 Size2 @Addr2\n")

	for _, window := range parsed.Windows {
		err := printWindowInstructions(&window, w)
		if err != nil {
			return err
		}
	}

	return nil
}

func printWindowInstructions(window *vcdiff.Window, w io.Writer) error {
	instructionStream := bytes.NewReader(window.InstructionSection)
	addressStream := bytes.NewReader(window.AddressSection)

	offset := 0

	for {
		code, err := instructionStream.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Look up instructions from code table
		inst1 := vcdiff.DefaultCodeTable.Get(code, 0)
		inst2 := vcdiff.DefaultCodeTable.Get(code, 1)

		fmt.Fprintf(w, "  %06x %03d  ", offset, code)

		// Print first instruction
		if inst1.Type != vcdiff.NoOp {
			err := printSingleInstruction(inst1, instructionStream, addressStream, w)
			if err != nil {
				return err
			}
		}

		// Print second instruction if it exists
		if inst2.Type != vcdiff.NoOp {
			fmt.Fprintf(w, " + ")
			err := printSingleInstruction(inst2, instructionStream, addressStream, w)
			if err != nil {
				return err
			}
		}

		fmt.Fprintf(w, "\n")
		offset++
	}

	return nil
}

func printSingleInstruction(inst vcdiff.Instruction, instructionStream *bytes.Reader, addressStream *bytes.Reader, w io.Writer) error {
	// Get instruction type string
	var typeStr string
	switch inst.Type {
	case vcdiff.Add:
		typeStr = "ADD"
	case vcdiff.Copy:
		typeStr = fmt.Sprintf("CPY_%d", inst.Mode)
	case vcdiff.Run:
		typeStr = "RUN"
	case vcdiff.NoOp:
		typeStr = "NOOP"
	default:
		typeStr = fmt.Sprintf("UNK_%02x", inst.Type)
	}

	// Get size
	size := uint32(inst.Size)
	if size == 0 && inst.Type != vcdiff.NoOp {
		var err error
		size, err = vcdiff.ReadVarint(instructionStream)
		if err != nil {
			return err
		}
	}

	// Get address for COPY instructions
	var addrStr string
	if inst.Type == vcdiff.Copy {
		switch inst.Mode {
		case 0: // SELF mode
			addr, err := vcdiff.ReadVarint(addressStream)
			if err != nil {
				return err
			}
			addrStr = fmt.Sprintf("S@%d", addr)
		case 1: // HERE mode
			offset, err := vcdiff.ReadVarint(addressStream)
			if err != nil {
				return err
			}
			addrStr = fmt.Sprintf("H@%d", offset)
		default:
			// Near/Same cache modes
			if inst.Mode < 6 {
				offset, err := vcdiff.ReadVarint(addressStream)
				if err != nil {
					return err
				}
				addrStr = fmt.Sprintf("N%d@%d", inst.Mode-2, offset)
			} else {
				b, err := addressStream.ReadByte()
				if err != nil {
					return err
				}
				addrStr = fmt.Sprintf("S%d@%d", inst.Mode-6, b)
			}
		}
	}

	if inst.Type == vcdiff.Copy {
		fmt.Fprintf(w, "%s %6d %s", typeStr, size, addrStr)
	} else {
		fmt.Fprintf(w, "%s %6d", typeStr, size)
	}

	return nil
}
