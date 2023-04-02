// The Monkey Language vm definition
package vm

import (
	"fmt"

	"github.com/freddiehaddad/monkey.compiler/pkg/code"
	"github.com/freddiehaddad/monkey.compiler/pkg/compiler"
	"github.com/freddiehaddad/monkey.interpreter/pkg/object"
)

const StackSize = 2048

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack []object.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		stack: make([]object.Object, StackSize),
		sp:    0,
	}
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			if err := vm.push(vm.constants[constIndex]); err != nil {
				return err
			}
		case code.OpAdd:
			rValue := vm.pop()
			lValue := vm.pop()
			rInteger := rValue.(*object.Integer).Value
			lInteger := lValue.(*object.Integer).Value
			sum := rInteger + lInteger
			vm.push(&object.Integer{Value: sum})
		case code.OpPop:
			vm.pop()
		}
	}
	return nil
}

func (vm *VM) LastPoppedStackElement() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}
