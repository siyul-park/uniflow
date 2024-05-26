package code

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// Opcode represents an opcode.
type Opcode byte

const (
	// OpConstant is an opcode to push a constant value on to the stack.
	OpConstant Opcode = iota
	// OpPop is an opcode to pop the topmost element off the stack.
	OpPop
	// OpAdd is an opcode for addition (+).
	OpAdd
	// OpSub is an opcode for subtraction (-).
	OpSub
	// OpMul is an opcode for multiplication (*).
	OpMul
	// OpDiv is an opcode for division (/).
	OpDiv
	// OpTrue is an opcode to push `true` value on to the stack.
	OpTrue
	// OpFalse is an opcode to push `false` value on to the stack.
	OpFalse
	// OpEqual is an opcode to check the equality of the two topmost elements on the stack.
	OpEqual
	// OpNotEqual is an opcode to check the inequality of the two topmost elements on the stack.
	OpNotEqual
	// OpGreaterThan is an opcode to check the second topmost element is greater than the first.
	OpGreaterThan
	// OpGreaterThanOrEqual is an opcode to check the second topmost element is greater than or
	// equal to the first.
	OpGreaterThanOrEqual
	// OpAnd is an opcode for logical AND.
	OpAnd
	// OpOr is an opcode for logical OR.
	OpOr
	// OpMinus is an opcode to negate integers.
	OpMinus
	// OpBang is an opcode to negate booleans.
	OpBang
	// OpJumpNotTruthy is an opcode to jump if the condition is not truthy.
	OpJumpNotTruthy
	// OpJump is an opcode to jump.
	OpJump
	// OpNil is an opcode to push `nil` value on to the stack.
	OpNil
	// OpSetGlobal is an opcode to create a global binding.
	OpSetGlobal
	// OpGetGlobal is an opcode to retrieve a value of a global binding.
	OpGetGlobal
	// OpArray is an opcode to create an array.
	OpArray
	// OpHash is an opcode to create a hash map.
	OpHash
	// OpSetIndex is an opcode to set a value into the place at the index in an indexed data
	// structure.
	OpSetIndex
	// OpGetIndex is an opcode to get an element at the index from an indexed data structure.
	OpGetIndex
	// OpCall is an opcode to call compiled functions.
	OpCall
	// OpReturnValue is an opcode to return a value from a function.
	OpReturnValue
	// OpReturn is an opcode to return from a function without return value.
	OpReturn
	// OpSetLocal is an opcode to create a local binding.
	OpSetLocal
	// OpGetLocal is an opcode to retrieve a value of a local binding.
	OpGetLocal
	// OpGetBuiltin is an opcode to get a built-in function.
	OpGetBuiltin
	// OpClosure is an opcode to create a closure.
	OpClosure
	// OpGetFree is an opcode to retrieve a free variable on to the stack.
	OpGetFree
	// OpCurrentClosure is an opcode to self-reference the current closure.
	OpCurrentClosure
)

// Definition represents the definition of an opcode.
type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant:           {Name: "OpConstant", OperandWidths: []int{2}},
	OpPop:                {Name: "OpPop", OperandWidths: nil},
	OpAdd:                {Name: "OpAdd", OperandWidths: nil},
	OpSub:                {Name: "OpSub", OperandWidths: nil},
	OpMul:                {Name: "OpMul", OperandWidths: nil},
	OpDiv:                {Name: "OpDiv", OperandWidths: nil},
	OpTrue:               {Name: "OpTrue", OperandWidths: nil},
	OpFalse:              {Name: "OpFalse", OperandWidths: nil},
	OpEqual:              {Name: "OpEqual", OperandWidths: nil},
	OpNotEqual:           {Name: "OpNotEqual", OperandWidths: nil},
	OpGreaterThan:        {Name: "OpGreaterThan", OperandWidths: nil},
	OpGreaterThanOrEqual: {Name: "OpGreaterThanOrEqual", OperandWidths: nil},
	OpAnd:                {Name: "OpAnd", OperandWidths: nil},
	OpOr:                 {Name: "OpOr", OperandWidths: nil},
	OpMinus:              {Name: "OpMinus", OperandWidths: nil},
	OpBang:               {Name: "OpBang", OperandWidths: nil},
	OpJumpNotTruthy:      {Name: "OpJumpNotTruthy", OperandWidths: []int{2}},
	OpJump:               {Name: "OpJump", OperandWidths: []int{2}},
	OpNil:                {Name: "OpNil", OperandWidths: nil},
	OpSetGlobal:          {Name: "OpSetGlobal", OperandWidths: []int{2}},
	OpGetGlobal:          {Name: "OpGetGlobal", OperandWidths: []int{2}},
	OpArray:              {Name: "OpArray", OperandWidths: []int{2}},
	OpHash:               {Name: "OpHash", OperandWidths: []int{2}},
	OpSetIndex:           {Name: "OpSetIndex", OperandWidths: nil},
	OpGetIndex:           {Name: "OpGetIndex", OperandWidths: nil},
	OpCall:               {Name: "OpCall", OperandWidths: []int{1}},
	OpReturnValue:        {Name: "OpReturnValue", OperandWidths: nil},
	OpReturn:             {Name: "OpReturn", OperandWidths: nil},
	OpSetLocal:           {Name: "OpSetLocal", OperandWidths: []int{1}},
	OpGetLocal:           {Name: "OpGetLocal", OperandWidths: []int{1}},
	OpGetBuiltin:         {Name: "OpGetBuiltin", OperandWidths: []int{1}},
	OpClosure:            {Name: "OpClosure", OperandWidths: []int{2, 1}},
	OpGetFree:            {Name: "OpGetFree", OperandWidths: []int{1}},
	OpCurrentClosure:     {Name: "OpCurrentClosure", OperandWidths: nil},
}

// Lookup performs a lookup for `op` in the definitions of opcodes.
func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

// Instructions represents a sequence of instructions.
type Instructions []byte

func (insns Instructions) String() string {
	var out strings.Builder

	i := 0
	for i < len(insns) {
		def, err := Lookup(insns[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, insns[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, insns.formatInstruction(def, operands))

		i += 1 + read
	}

	return out.String()
}

func (insns Instructions) formatInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand length %d does not match defined %d",
			len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s 0x%X", def.Name, operands[0])
	case 2:
		return fmt.Sprintf("%s 0x%X 0x%X", def.Name, operands[0], operands[1])
	}

	return fmt.Sprintf("ERROR: unhandled operand width for %s: %d", def.Name, operandCount)
}

// Make makes a bytecode instruction sequence from an opcode and operands.
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return nil
	}

	insnLen := 1
	for _, w := range def.OperandWidths {
		insnLen += w
	}

	insn := make([]byte, insnLen)
	insn[0] = byte(op)
	offset := 1

	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 1: // 1 byte-width operand
			insn[offset] = byte(o)
		case 2: // 2 byte-width operand
			binary.BigEndian.PutUint16(insn[offset:], uint16(o))
		}
		offset += width
	}

	return insn
}

// ReadOperands reads operands from bytecode instructions based on the definition of an opcode.
// It returns operands read and the offset describing the starting position of next opcode.
func ReadOperands(def *Definition, insns Instructions) (operands []int, offset int) {
	operands = make([]int, len(def.OperandWidths))

	for i, width := range def.OperandWidths {
		switch width {
		case 1: // 1 byte-width operand
			operands[i] = int(ReadUint8(insns[offset:]))
		case 2: // 2 byte-width operand
			operands[i] = int(ReadUint16(insns[offset:]))
		}

		offset += width
	}

	return operands, offset
}

// ReadUint8 reads a single uint8 value from bytecode instruction sequence.
func ReadUint8(insns Instructions) uint8 {
	return uint8(insns[0])
}

// ReadUint16 reads a single uint16 value from bytecode instruction sequence.
func ReadUint16(insns Instructions) uint16 {
	return binary.BigEndian.Uint16(insns)
}
