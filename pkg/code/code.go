// The Monkey Language bytecode definition
package code

import (
	"encoding/binary"
	"fmt"
)

type Instructions []byte
type Opcode byte

const (
	OpConstant Opcode = iota
)

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}},
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
