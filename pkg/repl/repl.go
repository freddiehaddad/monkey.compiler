// The Monkey Language REPL
package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/freddiehaddad/monkey.compiler/pkg/compiler"
	"github.com/freddiehaddad/monkey.compiler/pkg/vm"
	"github.com/freddiehaddad/monkey.interpreter/pkg/lexer"
	"github.com/freddiehaddad/monkey.interpreter/pkg/parser"
)

const PROMPT = "> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Fprintf(out, PROMPT)

		if scanned := scanner.Scan(); !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		compiler := compiler.New()
		if err := compiler.Compile(program); err != nil {
			fmt.Fprintf(out, "Woops! Compilation failed:\n %s\n", err)
			continue
		}

		machine := vm.New(compiler.Bytecode())
		if err := machine.Run(); err != nil {
			fmt.Fprintf(out, "Woops! Executing bytecode failed:\n %s\n", err)
			continue
		}

		stackTop := machine.LastPoppedStackElement()
		io.WriteString(out, stackTop.Inspect())
		io.WriteString(out, "\n")
	}
}

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, "Woops! Parser errors detected...\n")
	io.WriteString(out, "  Errors:\n")
	for _, error := range errors {
		io.WriteString(out, "\t"+error+"\n")
	}
}
