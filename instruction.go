package vcdiff

// InstructionType represents the type of VCDIFF instruction
type InstructionType byte

const (
	NoOp InstructionType = 0
	Add  InstructionType = 1
	Run  InstructionType = 2
	Copy InstructionType = 3
)

// String returns string representation of instruction type
func (it InstructionType) String() string {
	switch it {
	case NoOp:
		return "NOOP"
	case Add:
		return "ADD"
	case Run:
		return "RUN"
	case Copy:
		return "COPY"
	default:
		return "UNKNOWN"
	}
}

// Instruction represents a single VCDIFF instruction from the code table
type Instruction struct {
	Type InstructionType
	Size byte
	Mode byte
}

// RuntimeInstruction represents an instruction with resolved size during decoding
type RuntimeInstruction struct {
	Type InstructionType
	Size uint32
	Mode byte
	Addr uint32
	Data []byte
}

// NewInstruction creates a new instruction
func NewInstruction(instrType InstructionType, size byte, mode byte) Instruction {
	return Instruction{
		Type: instrType,
		Size: size,
		Mode: mode,
	}
}
