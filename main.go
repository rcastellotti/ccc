package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"slices"

	"github.com/rcastellotti/ccc/pkg/cmd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var interactive = flag.Bool("i", false, "")
var magic = flag.Bool("m", false, "")

var usage = `Usage: ccc [options...] <curl commmand>	

Options:
  -i: Interactive mode ~> open $EDITOR buffer to edit curl command
  -m: Magic mode ~> automatically yank off some analytics and predefined cookies
`

func OpenCurlInVim(command string) (string, error) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "ccc-*.txt")
	if err != nil {
		return "", err
	}
	fmt.Println(tmpFile.Name())
	defer tmpFile.Close()

	// Write formatted curl command to the temp file
	_, err = tmpFile.WriteString(command + "\n")
	if err != nil {
		return "", err
	}

	// Open Vim with the temp file
	cmd := exec.Command("vim", tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run Vim and wait for it to finish
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	// After Vim is closed, read the file contents
	content, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		return "", err
	}

	// Return the content of the file as a string
	return string(content), nil
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if _, debug := os.LookupEnv("DEBUG"); debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	flag.Parse()
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}

	// We don't need to have a separator like `--` between our flags and curl flags,
	// we can just get the index for 'curl' in `os.Args`, and fail fast.
	curlIndex := slices.Index(os.Args, "curl")
	if curlIndex == -1 || flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	command, err := cmd.Parse(os.Args[curlIndex+1:])
	if err != nil {
		log.Error().Err(err).Msg("Parse curl from os.Args:")
	}
	editedcommand, err := OpenCurlInVim(command.String())
	if err != nil {
		panic(err)
	}
	fmt.Println(editedcommand)
}
