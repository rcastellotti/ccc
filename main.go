// in 2024 we do chatgpt-based development :)
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	// Ensure there are command-line arguments (excluding the program name)
	if len(os.Args) < 2 {
		log.Fatal("Usage: exec_command <command> [args...]")
	}

	cmdName,cmdArgs := os.Args[1], os.Args[2:]

	cmd := exec.Command(cmdName, cmdArgs...)

	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error executing command: %v\n", err)
	}

	fmt.Printf("Executed command: %s %v\n", cmdName, cmdArgs)
}
