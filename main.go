package main

import (
	"flag"
	"fmt"
	"os"
	"slices"

	"github.com/rcastellotti/ccc/cmd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	interactive := flag.Bool("i", false, "interactive mode, open command in $EDITOR")
	verbose := flag.Bool("v", false, "verbose mode")
	flag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// log.Logger = log.With().Caller().Logger()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if *verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	curlIndex := slices.Index(os.Args, "curl")
	if curlIndex == -1 || flag.NArg() < 1 {
		flag.Usage()
		log.Debug().Strs("command", os.Args).Msg("error:")
		os.Exit(1)
	}

	command, err := cmd.Parse(os.Args[curlIndex+1:])
	// log.Debug().Str("command", command.String()).Msg("parsed command:")
	if err != nil {
		log.Error().Err(err).Msg("main:")
		os.Exit(1)
	}

	if *interactive {
		editedcommand, err := command.OpenInEditor()
		if err != nil {
			panic(err)
		}

		fmt.Println(editedcommand)
	} else {
		fmt.Println(command.String())
	}
}
