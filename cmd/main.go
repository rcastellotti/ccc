package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/rcastellotti/ccc"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Usage: exec_command <command> [args...]")
	}

	cmdName, cmdArgs := os.Args[1], os.Args[2:]
	_ = cmdName

	// 	cmd := exec.Command(cmdName, cmdArgs...)
	// 	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	// 	err := cmd.Run()
	// 	if err != nil {
	// 		log.Fatalf("Error executing command: %v\n", err)
	// 	}
	// 	fmt.Printf("Executed command: %s %v\n", cmdName, cmdArgs)

	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelInfo)
	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: lvl})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	if _, debug := os.LookupEnv("DEBUG"); debug {
		lvl.Set(slog.LevelDebug)
	}

	cmd, err := ccc.Parse(cmdArgs)
	if err != nil {
		panic(err)
	}

	fmt.Println(cmd.String())
}
