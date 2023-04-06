// The Monkey Language vm definition
package vm

import (
	"fmt"

	"github.com/freddiehaddad/monkey.compiler/pkg/code"
	"github.com/freddiehaddad/monkey.compiler/pkg/compiler"
	"github.com/freddiehaddad/monkey.interpreter/pkg/object"
)

const (
	StackSize  = 2048
	GlobalSize = 65536
)

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}

var Null = &object.Null{}

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	global []object.Object

	stack []object.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]
}

func NewWithState(bytecode *compiler.Bytecode, global []object.Object) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		global: global,

		stack: make([]object.Object, StackSize),
		sp:    0,
	}
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		global: make([]object.Object, GlobalSize),

		stack: make([]object.Object, StackSize),
		sp:    0,
	}
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpNull:
			if err := vm.push(Null); err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			if err := vm.push(vm.constants[constIndex]); err != nil {
				return err
			}
		case code.OpTrue:
			if err := vm.push(True); err != nil {
				return err
			}
		case code.OpFalse:
			if err := vm.push(False); err != nil {
				return err
			}
		case code.OpArray:
			elements := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2

			array := &object.Array{Elements: make([]object.Object, elements)}

			// Last array element is at the top of the stack
			for i := elements; i > 0; i-- {
				array.Elements[i-1] = vm.pop()
			}

			if err := vm.push(array); err != nil {
				return err
			}
		case code.OpHash:
			pairs := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2

			hash := &object.Hash{
				Pairs: make(map[object.HashKey]object.HashPair),
			}

			// value, pair is the order on the stack
			for i := 0; i < pairs; i++ {
				value := vm.pop()
				key := vm.pop()

				pair := object.HashPair{Key: key, Value: value}
				hashKey, ok := key.(object.Hashable)
				if !ok {
					return fmt.Errorf("unusable as hash key: %s", key.Type())
				}

				hash.Pairs[hashKey.HashKey()] = pair
			}

			if err := vm.push(hash); err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			vm.executeBinaryOperation(op)
		case code.OpEqual, code.OpNotEqual, code.OpLessThan, code.OpGreaterThan:
			vm.executeComparison(op)
		case code.OpBang:
			if err := vm.executeBangOperator(); err != nil {
				return err
			}
		case code.OpMinus:
			if err := vm.executeMinusOperator(); err != nil {
				return err
			}
		case code.OpJump:
			address := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip = address - 1
		case code.OpJumpNotTruthy:
			address := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				ip = address - 1
			}
		case code.OpGetGlobal:
			globalIndex := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2
			value := vm.global[globalIndex]
			if err := vm.push(value); err != nil {
				return err
			}
		case code.OpSetGlobal:
			globalIndex := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2
			value := vm.pop()
			vm.global[globalIndex] = value
		}
	}
	return nil
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	value := operand.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -value})
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(left == right))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(left != right))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, leftType, rightType)
	}
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return vm.executeBinaryIntegerOberation(op, left, right)
	}

	if leftType == object.STRING_OBJ && rightType == object.STRING_OBJ {
		return vm.executeBinaryStringOperation(op, left, right)
	}

	return fmt.Errorf("unsupported types for binary operation: %s %s",
		leftType, rightType)
}

func (vm *VM) executeIntegerComparison(
	op code.Opcode, left, right object.Object) error {

	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue == rightValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue != rightValue))
	case code.OpLessThan:
		return vm.push(nativeBoolToBooleanObject(leftValue < rightValue))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func nativeBoolToBooleanObject(result bool) object.Object {
	if result {
		return True
	} else {
		return False
	}
}

func (vm *VM) executeBinaryIntegerOberation(
	op code.Opcode, left, right object.Object) error {

	lValue := left.(*object.Integer).Value
	rValue := right.(*object.Integer).Value

	var result int64

	switch op {
	case code.OpAdd:
		result = lValue + rValue
	case code.OpSub:
		result = lValue - rValue
	case code.OpMul:
		result = lValue * rValue
	case code.OpDiv:
		result = lValue / rValue
	default:
		return fmt.Errorf("unknown interger operator: %d", op)
	}

	return vm.push(&object.Integer{Value: result})
}

func (vm *VM) executeBinaryStringOperation(
	op code.Opcode, left, right object.Object) error {

	lValue := left.(*object.String).Value
	rValue := right.(*object.String).Value

	var result string

	switch op {
	case code.OpAdd:
		result = lValue + rValue
	default:
		return fmt.Errorf("unknown interger operator: %d", op)
	}

	return vm.push(&object.String{Value: result})
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
