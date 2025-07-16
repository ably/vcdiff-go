package vcdiff

// CodeTable represents the VCDIFF instruction code table
type CodeTable struct {
	entries [256][2]Instruction
}

// Get returns the instruction at the given code and slot
func (ct *CodeTable) Get(code byte, slot int) Instruction {
	return ct.entries[code][slot]
}

// BuildDefaultCodeTable creates the default code table specified in RFC 3284
func BuildDefaultCodeTable() *CodeTable {
	ct := &CodeTable{}

	// Initialize all entries to NoOp
	for i := 0; i < 256; i++ {
		ct.entries[i][0] = NewInstruction(NoOp, 0, 0)
		ct.entries[i][1] = NewInstruction(NoOp, 0, 0)
	}

	// Entry 0: RUN with size 0
	ct.entries[0][0] = NewInstruction(Run, 0, 0)

	// Entries 1-18: ADD with sizes 0-17
	for i := byte(0); i < 18; i++ {
		ct.entries[i+1][0] = NewInstruction(Add, i, 0)
	}

	index := 19

	// Entries 19-162: COPY instructions with different modes and sizes
	for mode := byte(0); mode < 9; mode++ {
		// COPY with size 0 (size will be read from stream)
		ct.entries[index][0] = NewInstruction(Copy, 0, mode)
		index++

		// COPY with sizes 4-18
		for size := byte(4); size < 19; size++ {
			ct.entries[index][0] = NewInstruction(Copy, size, mode)
			index++
		}
	}

	// Entries 163-234: Combined ADD+COPY instructions
	for mode := byte(0); mode < 6; mode++ {
		for addSize := byte(1); addSize < 5; addSize++ {
			for copySize := byte(4); copySize < 7; copySize++ {
				ct.entries[index][0] = NewInstruction(Add, addSize, 0)
				ct.entries[index][1] = NewInstruction(Copy, copySize, mode)
				index++
			}
		}
	}

	// Entries 235-246: More combined ADD+COPY instructions
	for mode := byte(6); mode < 9; mode++ {
		for addSize := byte(1); addSize < 5; addSize++ {
			ct.entries[index][0] = NewInstruction(Add, addSize, 0)
			ct.entries[index][1] = NewInstruction(Copy, 4, mode)
			index++
		}
	}

	// Entries 247-255: COPY+ADD combinations
	for mode := byte(0); mode < 9; mode++ {
		ct.entries[index][0] = NewInstruction(Copy, 4, mode)
		ct.entries[index][1] = NewInstruction(Add, 1, 0)
		index++
	}

	return ct
}

// DefaultCodeTable is the default code table instance
var DefaultCodeTable = BuildDefaultCodeTable()
