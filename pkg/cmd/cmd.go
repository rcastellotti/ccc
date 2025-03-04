package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"unicode"

	"github.com/rs/zerolog/log"
)

type Command struct {
	CurlKeyword string         `json:"curlKeyword"`
	Method      string         `json:"method"`
	URL         string         `json:"url"`
	Headers     []string       `json:"headers"` // Does not contain cookies
	Cookies     []*http.Cookie `json:"cookies"` // Storing separately simplifies manipulation
	Flags       []string       `json:"flags"`
}

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

func New() *Command {
	return &Command{
		CurlKeyword: "curl",            // Default to "curl" as the keyword
		Method:      "GET",             // Default HTTP method
		URL:         "",                // URL must be set explicitly
		Headers:     make([]string, 0), // Empty headers map
		Cookies:     nil,               // Empty cookies map
		Flags:       make([]string, 0), // Empty flags slice
	}
}

// Following convention is adopted: if you have a better idea, please let me know :)
// flag: --compressed (no value)
// option: -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:133.0) Gecko/20100101 Firefox/133.0' (value)'
//         Key: -H
//         Value: 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:133.0) Gecko/20100101 Firefox/133.0' (value)'

// I will update these as stuff breaks, I have absolutely no intention to catch them all now :)
func isOption(s string) bool {
	options := []string{
		"-H", // header
		"-X", // method
		"-b", // cookies
		"-d",
		"-raw",
		"--data-ascii",
		"--data-raw",
		"--output",
	}
	return slices.Contains(options, s)
}
func isFlag(s string) bool {
	flags := []string{"--compressed", "--insecure", "-G", "-i", "--data", "-s"}
	return slices.Contains(flags, s)
}

// can either be
// -b 'colruytlanguage=fr'
// -H 'Cookie: prov=c68c0bf8-c954-41a6-92a9-d7ac3b5d8663'
func mustParseCookies(cookieStr string) []*http.Cookie {
	cookies, err := http.ParseCookie(cookieStr)
	if err != nil {
		panic(err)
	}
	return cookies
}

var ErrUnexpectedToken = errors.New("unexpected token in curl command")

// Parse function gets input from os.Args, ccc is called as ccc <CCC_FLAGS> curl <CURL_FLAGS/OPTIONS>,
// This means that options (see definition above) can be parsed by examining `cTok` (current token) and `nTok` (next token)
// Whenever an option is detected, we skip the next iteration.
func Parse(rawCmd []string) (*Command, error) {
	cmd := New()
	i := 0
	for {
		if i == len(rawCmd) {
			break
		}
		cTok := strings.Trim(rawCmd[i], "'")
		// Ignore trailing space after command itself
		if strings.TrimSpace(cTok) == "" || strings.HasPrefix(cTok, "curl") {
			i++
		} else if strings.HasPrefix(cTok, "http") {
			cmd.URL = cTok
			i++
		} else if isFlag(cTok) {
			cmd.Flags = append(cmd.Flags, cTok)
			i++
		} else if isOption(cTok) {
			nTok := rawCmd[i+1]
			if cTok == "-b" {
				log.Debug().Str("flag", nTok).Msg("cookie string to parse")
				cmd.Cookies = mustParseCookies(nTok)
			} else if cTok == "-H" {
				cmd.Headers = append(cmd.Headers, nTok)
			}
			i += 2
		} else {
			return nil, fmt.Errorf("unexpected token: %w: %q", ErrUnexpectedToken, cTok)
		}
	}
	return cmd, nil
}

// String method to format the Command struct into a curl command
func (c Command) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("%s \\", c.CurlKeyword))

	if c.Method != "" && c.Method != "GET" {
		parts = append(parts, fmt.Sprintf("\n  -X %s \\", c.Method))
	}

	// Add headers, each on a new line with -H
	if len(c.Headers) > 0 {
		for _, header := range c.Headers {
			parts = append(parts, fmt.Sprintf("\n  -H %q \\", header))
		}
	}

	// Add cookies, each on a new line with -b
	if len(c.Cookies) > 0 {
		for _, cookie := range c.Cookies {
			parts = append(parts, fmt.Sprintf("\n  -b %q \\", fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)))
		}
	}

	// Add flags
	for _, flag := range c.Flags {
		parts = append(parts, fmt.Sprintf("\n%s", flag))
	}

	// Add URL
	parts = append(parts, fmt.Sprintf("\n%q", c.URL))

	return strings.Join(parts, " ")
}
