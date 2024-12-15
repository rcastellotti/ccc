// Knowing I cannot make this code look too interesting, I will at least try to keep it short (and single file).
// I'm trying my best boss, come hang out in gh issues, come say it, I am dummy, I know, you would have done this better.
// I hope no-one will ever force you to use this, let alone read this source code.
// live, laugh, love bears.

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"unicode"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var interactive = flag.Bool("i", false, "")

var usage = `Usage: ccc [options...] <curl commmand>

Options:
  -i: Interactive mode (open $EDITOR buffer to edit curl command)
`

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

	// We don't need to have a separator like `--` between our flags
	// and curl flags, we can just get the index for 'curl' in `os.Args`,
	// this additionally gives an opportunity to fail faster.
	curlIndex := slices.Index(os.Args, "curl")
	if curlIndex == -1 || flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	cmd, err := Parse(os.Args[curlIndex+1:])
	if err != nil {
		log.Error().Err(err).Msg("Parse:")
	}

	if *interactive {
		newCmd, err := EditInEditor(cmd.String())
		if err != nil {
			log.Fatal().Err(err).Msgf("Cannot run command %s", newCmd)
		}
		cmdSlice := SplitCurlCommand(newCmd)
		curlCmd := exec.Command("curl", cmdSlice[1:]...)
		curlCmd.Stdout, curlCmd.Stderr = os.Stdout, os.Stderr
		if err := curlCmd.Run(); err != nil {
			log.Fatal().Err(err).Msgf("Cannot run command %s", curlCmd)
		}
	}

}

type KeyVal struct {
	Key string
	Val string
}
type Command struct {
	Method  string   `json:"method"`
	URL     string   `json:"url"`
	Headers []KeyVal `json:"headers"` // Does not contain cookies
	Cookies []KeyVal `json:"cookies"` // Storing separately somplifies manipulation
	Flags   []string `json:"flags"`
}

func (cmd Command) String() string {
	var result strings.Builder

	result.WriteString(`curl -X ` + cmd.Method + " ")
	result.WriteString("'" + cmd.URL + "' \\\n")
	for _, f := range cmd.Flags {
		result.WriteString("  " + f + " \\\n")
	}

	for i, h := range cmd.Headers {
		result.WriteString("  " + "-H '" + h.Key + ": " + h.Val + "'")
		if i != len(cmd.Headers) {
			result.WriteString(" \\\n")
		}
	}

	for i, c := range cmd.Cookies {
		result.WriteString("  " + "--cookie '" + c.Key + ": " + c.Val + "'")
		if i != len(cmd.Cookies)-1 {
			result.WriteString(" \\\n")
		}
	}

	return result.String()
}

// Following convention is adopted: if you have a better idea, please let me know :)
// flag: --compressed (no value)
// option: -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:133.0) Gecko/20100101 Firefox/133.0' (value)'
//         Key: -H
//         Value: 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:133.0) Gecko/20100101 Firefox/133.0' (value)'

// I will update these as stuff breaks, I have absolutely no intention to catch them all now :)
func isOption(s string) bool {
	options := []string{"-H", "-X", "-d", "-raw", "--data-ascii", "--output"}
	return slices.Contains(options, s)
}
func isFlag(s string) bool {
	flags := []string{"--compressed", "--insecure", "-G", "-i", "--data", "-s"}
	return slices.Contains(flags, s)
}

func parseHeader(rhs string) KeyVal {
	rawHString := strings.Trim(rhs, "'")
	values := strings.SplitN(rawHString, ":", 2)
	key, val := values[0], strings.TrimSpace(values[1])
	h := KeyVal{key, val}
	return h
}

func parseCookies(rhs string) []KeyVal {
	rawHString := strings.Trim(rhs, "'")
	values := strings.SplitN(rawHString, ":", 2)
	cookies := strings.Split(values[1], ";")

	var resalts []KeyVal
	for _, c := range cookies {
		ca := strings.Split(c, "=")
		h := KeyVal{strings.TrimSpace(ca[0]), strings.TrimSpace(ca[1])}
		resalts = append(resalts, h)
	}
	return resalts
}

// Parse function gets input from os.Args, ccc is called as ccc <CCC_FLAGS> curl <CURL_FLAGS/OPTIONS>,
// This means that options (see definition above) can only be parsed by examining `cTok` (current token) and `nTok` (next token)
// Thus, after parsing an option we need to increment the index by two.
var ErrUnexpectedToken = errors.New("unexpected token in curl command")

func Parse(rawCmd []string) (*Command, error) {
	var cmd Command
	cmd.Method = "GET" // solid default
	i := 0
	for {
		if i == len(rawCmd) {
			break
		}
		cTok := strings.Trim(rawCmd[i], "'")
		log.Debug().Str("token", cTok).Msg("parse")

		// Ignore trailing space after command itself
		if strings.TrimSpace(cTok) == "" {
			i++
		} else if strings.HasPrefix(cTok, "http") {
			cmd.URL = cTok
			i++
		} else if isFlag(cTok) {
			log.Debug().Str("flag", cTok).Msg("parsed flag")
			cmd.Flags = append(cmd.Flags, cTok)
			i++
		} else if isOption(cTok) {
			nTok := rawCmd[i+1]
			log.Debug().Str("key", cTok).Str("val", nTok).Msg("parsed option")
			if cTok == "-H" {
				// When copying as curl from chrome cookies are in a `-H` option such as
				//
				//	`-H 'cookie: _k1=v1; k2=v2; k3=v3`
				//
				// We prefer to use the `--cookie` option, so we save them in a separate field.
				if strings.HasPrefix(strings.Trim(strings.ToLower(nTok), "'"), "cookie") {
					cmd.Cookies = parseCookies(nTok)
				} else if cTok == "-X" {
					log.Debug().Str("method", cTok).Msg("parsed method")
					cmd.Method = nTok
				} else {
					cmd.Headers = append(cmd.Headers, parseHeader(nTok))
				}
			}
			i += 2
		} else {
			return nil, fmt.Errorf("unexpected token: %w: %q", ErrUnexpectedToken, cTok)
		}
	}
	return &cmd, nil
}

// TODO: handle string is empty
func EditInEditor(initialContent string) (string, error) {
	tempFile, err := os.CreateTemp("", "editor-*.txt")
	if err != nil {
		return "", err
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write([]byte(initialContent)); err != nil {
		tempFile.Close()
		return "", err
	}
	tempFile.Close()

	editor := os.Getenv("EDITOR")
	if editor == "" {
		// vi || gtfo, dont' ask yourself whether CCC can pick good fallbacks,
		// if you are so picky, ask yourself why you have no `$EDITOR` set
		editor = "vi"
	}
	cmd := exec.Command(editor, tempFile.Name())

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error running editor: %v", err)
	}

	editedContent, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(editedContent)), nil
}

// SplitCurlCommand splits a curl command string into a slice
// of strings other functions can use, additionally, quotes are stripped
// because `exec.Command` does not like them, aka
//
//	cmd:= exec.Command("curl", "'https://rcastellotti.dev'")
//
// fails, while
//
//	cmd:= exec.Command("curl", "https://rcastellotti.dev")
//
// just works :tm
func SplitCurlCommand(input string) []string {
	f := func(r rune) bool {
		return unicode.IsSpace(r) || r == '\\'
	}
	fields := strings.FieldsFunc(input, f)
	for i, field := range fields {
		if len(field) >= 2 {
			if (field[0] == '"' && field[len(field)-1] == '"') ||
				(field[0] == '\'' && field[len(field)-1] == '\'') {
				fields[i] = field[1 : len(field)-1]
			}
		}
	}

	return fields
}
