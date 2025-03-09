package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"slices"

	"github.com/rcastellotti/ccc/cmd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func OpenCurlInVim(command string) (string, error) {
	tmpFile, err := os.CreateTemp("", "ccc-*.txt")
	if err != nil {
		return "", err
	}

	tmpFileName := tmpFile.Name()
	log.Debug().Str("tmpfile", tmpFileName).Msg("cookie string to parse")
	defer tmpFile.Close()

	_, err = tmpFile.WriteString(command + "\n")
	if err != nil {
		return "", err
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		// vi || gtfo, dont' ask yourself whether ccc can have good fallbacks,
		// if you are so picky, ask yourself why you have no `$EDITOR` set
		editor = "vim"
	}

	cmd := exec.Command(editor, tmpFileName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(tmpFileName)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func main() {
	verbose := flag.Bool("v", false, "verbose mode")
	flag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if *verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// We don't need to have a separator like `--` between our flags and curl flags,
	// we can just get the index for 'curl' in `os.Args`, and fail fast.
	curlIndex := slices.Index(os.Args, "curl")
	if curlIndex == -1 || flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	command, err := cmd.Parse(os.Args[curlIndex+1:])
	log.Debug().Str("command", command.String()).Msg("parsed command:")

	if err != nil {
		log.Error().Err(err).Msg("Parse curl from os.Args:")
	}

	editedcommand, err := OpenCurlInVim(command.String())
	if err != nil {
		panic(err)
	}

	fmt.Println(editedcommand)
}
