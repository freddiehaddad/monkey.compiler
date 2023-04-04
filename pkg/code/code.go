// The Monkey Language bytecode definition
package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte
type Opcode byte

type Definition struct {
	Name          string
	OperandWidths []int
}

const (
	OpNull Opcode = iota

	OpConstant

	OpPop

	OpJump
	OpJumpNotTruthy

	OpTrue
	OpFalse

	OpBang
	OpMinus

	OpAdd
	OpSub
	OpMul
	OpDiv

	OpEqual
	OpNotEqual
	OpLessThan
	OpGreaterThan
)

var definitions = map[Opcode]*Definition{
	OpNull: {"OpNull", []int{}},

	OpConstant: {"OpConstant", []int{2}},

	OpPop: {"OpPop", []int{}},

	OpJump:          {"OpJump", []int{2}},
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}},

	OpBang:  {"OpBang", []int{}},
	OpMinus: {"OpMinus", []int{}},

	OpTrue:  {"OpTrue", []int{}},
	OpFalse: {"OpFalse", []int{}},

	OpAdd: {"OpAdd", []int{}},
	OpSub: {"OpSub", []int{}},
	OpMul: {"OpMul", []int{}},
	OpDiv: {"OpDiv", []int{}},

	OpEqual:       {"OpEqual", []int{}},
	OpNotEqual:    {"OpNotEqual", []int{}},
	OpLessThan:    {"OpLessThan", []int{}},
	OpGreaterThan: {"OpGreaterThan", []int{}},
}

func Lookup(op byte) (*Definition, error) {
	if def, ok := definitions[Opcode(op)]; ok {
		return def, nil
	}
	return nil, fmt.Errorf("opcode %d undefined", op)
}

func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionsLen := 1
	for _, w := range def.OperandWidths {
		instructionsLen += w
	}

	instruction := make([]byte, instructionsLen)
	instruction[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		}
		offset += width
	}

	return instruction
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		}

		offset += width
	}

	return operands, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])

		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))

		i += 1 + read
	}

	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n",
			len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return fmt.Sprintf("%s", def.Name)
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}
