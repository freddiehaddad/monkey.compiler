// The Monkey Language CLI
package main

import (
	"fmt"
	"log"
	"os"
	"os/user"

	"github.com/freddiehaddad/monkey.compiler/pkg/repl"
)

func main() {
	user, err := user.Current()
	if err != nil {
		log.Fatalf("Failed to get username: %s", err)
	}
	fmt.Printf("Hello %s! Welcome to the Monkey Language.\n", user.Username)
	fmt.Println("Press Ctrl+D to exit")

	repl.Start(os.Stdin, os.Stdout)

	fmt.Println("Goodbye", user.Username)
}
