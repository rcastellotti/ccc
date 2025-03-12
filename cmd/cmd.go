package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strings"
	"unicode"

	"github.com/rs/zerolog/log"
)

type Command struct {
	CurlKeyword string
	URL         string
	Headers     []string          // Does not contain cookies
	Cookies     []*http.Cookie    // Storing separately simplifies manipulation
	Options     map[string]string // Every other option (non flag, non header) simplifies manipulation
	Flags       []string          // every
}

func SplitCurlCommand(input string) []string {
	input = strings.TrimSpace(input)

	insideSingleQuotes := false
	insideDoubleQuotes := false
	f := func(r rune) bool {
		if r == '\'' {
			// if the quote is the first one we see we are definitely inside quotes,
			// conversely, if we saw a quote earlier, this quote must be the closing one
			insideSingleQuotes = !insideSingleQuotes
		}

		if r == '"' {
			insideDoubleQuotes = !insideDoubleQuotes
		}

		return (unicode.IsSpace(r) || r == '\\') && !insideSingleQuotes && !insideDoubleQuotes
	}
	fields := strings.FieldsFunc(input, f)
	log.Debug().Strs("fields", fields).Msg("splitted fields")

	return fields
}

func New() *Command {
	return &Command{
		CurlKeyword: "curl",
		URL:         "",
		Headers:     make([]string, 0),
		Cookies:     nil,
		Flags:       make([]string, 0),
		Options:     make(map[string]string),
	}
}

// Following convention is adopted: if you have a better idea, please let me know :)
// flag: --compressed (no value)
// option: -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:133.0) Gecko/20100101 Firefox/133.0' (value)'
//   Key: -H
//   Value: 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:133.0) Gecko/20100101 Firefox/133.0' (value)'

// I will update these as stuff breaks, I have absolutely no intention to catch them all now :).
func isOption(s string) bool {
	options := []string{
		"-H", // header
		"-X", // method
		"-b", // cookies
		"-d", "-raw", "--data-ascii", "--data-raw",
		"--output",
	}

	return slices.Contains(options, s)
}

func isFlag(s string) bool {
	flags := []string{"--compressed", "--insecure", "-G", "-i", "-s"}

	return slices.Contains(flags, s)
}

// can either be -b 'bear=grizzly' or -H 'Cookie: kitten=siberian'.
func mustParseCookies(cookieStr string) []*http.Cookie {
	cookieStr = strings.ReplaceAll(cookieStr, "\"", "")
	cookieStr = strings.ReplaceAll(cookieStr, "'", "")
	cookieStr = strings.ReplaceAll(cookieStr, "cookie:", "")
	cookieStr = strings.TrimSpace(cookieStr)
	// cookie string cannot end with `;`
	cookieStr = strings.TrimSuffix(cookieStr, ";")
	log.Debug().Str("cookieStr", cookieStr).Msg("mustParseCookies:")

	cookies, err := http.ParseCookie(cookieStr)
	if err != nil {
		panic(err)
	}

	return cookies
}

var ErrUnrecognizedToken = errors.New("unrecognized token in curl command")

// Parse function gets input from os.Args, ccc is called as ccc <CCC_FLAGS> curl <CURL_FLAGS/OPTIONS>,
// This means that options (see definition above) can be parsed by examining `cTok` (current token) and `nTok` (next token)
// Whenever an option is detected, we skip the next iteration.
func Parse(rawCmd []string) (*Command, error) {
	cmd := New()
	counter := 0

	for {
		if counter == len(rawCmd) {
			break
		}

		cTok := strings.Trim(rawCmd[counter], "'")
		log.Debug().Str("cTok", cTok).Msg("current token")

		switch {
		case strings.HasPrefix(cTok, "curl"):
			counter++

		case strings.HasPrefix(cTok, "http"):
			cmd.URL = cTok
			counter++

		case isFlag(cTok):
			cmd.Flags = append(cmd.Flags, cTok)
			counter++

		case isOption(cTok):
			nTok := rawCmd[counter+1]
			log.Debug().Str("nTok", nTok).Msg("next token")

			switch cTok {
			case "-b":
				cmd.Cookies = append(cmd.Cookies, mustParseCookies(nTok)...)
			case "-H":
				if !strings.Contains(strings.ToLower(nTok), "cookie") {
					cmd.Headers = append(cmd.Headers, nTok)
				} else {
					nTok = strings.ReplaceAll(nTok, "\"", "")
					cmd.Cookies = append(cmd.Cookies, mustParseCookies(nTok)...)
				}
			default:
				log.Debug().Str("option key", cTok).Str("option val", nTok).Msg("flag to append")
				cmd.Options[cTok] = nTok
			}

			counter += 2

		default:
			log.Error().Str("cTok", cTok).Msg("unrecognized token:")

			return nil, fmt.Errorf("cmd.Parse: %w: %s", ErrUnrecognizedToken, cTok)
		}
	}

	return cmd, nil
}

func (c Command) String() string {
	parts := make([]string, 0, 1)
	parts = append(parts, fmt.Sprintf("%s \\", c.CurlKeyword))

	if len(c.Headers) > 0 {
		for _, header := range c.Headers {
			parts = append(parts, fmt.Sprintf("\n\t-H '%s' \\", header))
		}
	}

	if len(c.Cookies) > 0 {
		for _, cookie := range c.Cookies {
			parts = append(parts, fmt.Sprintf("\n\t-b '%s' \\", fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)))
		}
	}

	for _, flag := range c.Flags {
		parts = append(parts, fmt.Sprintf("\n\t%s \\", flag))
	}

	for ok, ov := range c.Options {
		parts = append(parts, fmt.Sprintf("\n\t%s '%s' \\", ok, ov))
	}

	parts = append(parts, fmt.Sprintf("\n\t%q", c.URL))

	return strings.Join(parts, "")
}

// temporarily returns a string, as we do not want fail if the user decides to
// edit a valid command into a broken one, this will probably change in the future.
func (c Command) OpenInEditor() (string, error) {
	tmpFile, err := os.CreateTemp("", "ccc-*.txt")
	if err != nil {
		return "", err
	}

	tmpFileName := tmpFile.Name()
	log.Debug().Str("tmpfile", tmpFileName).Msg("cookie string to parse")
	defer tmpFile.Close()

	_, err = tmpFile.WriteString(c.String() + "\n")
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
