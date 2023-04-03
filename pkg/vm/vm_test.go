// The Monkey Language vm definition unit tests
package vm

import (
	"fmt"
	"testing"

	"github.com/freddiehaddad/monkey.compiler/pkg/compiler"
	"github.com/freddiehaddad/monkey.interpreter/pkg/ast"
	"github.com/freddiehaddad/monkey.interpreter/pkg/lexer"
	"github.com/freddiehaddad/monkey.interpreter/pkg/object"
	"github.com/freddiehaddad/monkey.interpreter/pkg/parser"
)

type vmTestCase struct {
	input    string
	expected interface{}
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
	}

	return nil
}

func testBooleanObject(expected bool, actual object.Object) error {
	result, ok := actual.(*object.Boolean)
	if !ok {
		return fmt.Errorf("object is not Boolean. got=%T (%+v)",
			actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%t, want=%t",
			result.Value, expected)
	}

	return nil
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)

		comp := compiler.New()
		if err := comp.Compile(program); err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(comp.Bytecode())
		if err := vm.Run(); err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElem := vm.LastPoppedStackElement()

		testExpectedObject(t, tt.expected, stackElem)
	}
}

func testExpectedObject(t *testing.T, expected interface{}, actual object.Object) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		if err := testIntegerObject(int64(expected), actual); err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case bool:
		if err := testBooleanObject(expected, actual); err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
		}
	default:
		t.Errorf("unsupported type encountered: %T (%+v)", expected, expected)
	}
}

func TestComparisonOperators(t *testing.T) {
	tests := []vmTestCase{
		{"10 == 10", true},
		{"10 != 10", false},
		{"true == true", true},
		{"true == false", false},
		{"true != false", true},
		{"1 < 2", true},
		{"2 < 1", false},
		{"2 > 1", true},
		{"1 > 2", false},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}

	runVmTests(t, tests)
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
		{"3 - 2", 1},
		{"2 * 2", 4},
		{"10 / 2", 5},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"5 + 2 * 10", 25},
		{"5 * (2 + 10)", 60},
		{"-5", -5},
		{"-15", -15},
		{"-50 + 100 + -50", 0},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	runVmTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"true", true},
		{"false", false},
		{"!true", false},
		{"!false", true},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	runVmTests(t, tests)
}
