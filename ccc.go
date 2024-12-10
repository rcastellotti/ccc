package ccc

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"slices"
	"strings"
)

type Command struct {
	CurlCmd string        `json:"cmd"`
	Method  string        `json:"method"`
	URL     string        `json:"url"`
	Headers []http.Header `json:"headers"` // Does not contain cookies
	// TODO: is http.Cookie needed for our usecase?
	Cookies []*http.Cookie `json:"cookies"` // Storing separately somplifies manipulation
	Flags   []string       `json:"flags"`
}

func (c Command) String() string {
	var result strings.Builder

	curlCmd := `
curl -X {{.Method}} '{{ .URL }}' \
{{range .Flags -}}
{{"  "}}{{- . }} \
{{end -}}
{{range .Headers -}}
{{"  "}}-H '{{range  $k, $v := . }}{{$k}}: {{ range $v }}{{.}}{{end}}{{end}}' \
{{ end }}
{{- range .Cookies -}}
{{"  "}}--cookie '{{.Name }}={{.Value}}' \
{{end}}`

	templ := template.Must(template.New("myname").Parse(curlCmd))
	templ.Execute(&result, c)
	return result.String()
}

// Following convention is adopted: if you have a better idea, please let me know :)
// flag: --compressed (no value)
// option: -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:133.0) Gecko/20100101 Firefox/133.0' (value)'
//         Key: -H
//         Value: 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:133.0) Gecko/20100101 Firefox/133.0' (value)'

func isOption(s string) bool {
	// I will update these as stuff breaks, I have absolutely no intention to catch them all now :)
	options := []string{
		"-H", "-X",
		"-d", "-raw", "--data-ascii",
	}
	return slices.Contains(options, s)
}
func isFlag(s string) bool {
	// I will update these as stuff breaks, I have absolutely no intention to catch them all now :)
	flags := []string{"--compressed", "--insecure", "-G", "-i", "--data"}
	return slices.Contains(flags, s)
}

func parseHeader(rhs string) *http.Header {
	rawHString := strings.Trim(rhs, "'")
	values := strings.SplitN(rawHString, ":", 2)
	headerKey, headerValue := values[0], strings.TrimSpace(values[1])
	h := http.Header{headerKey: []string{headerValue}}
	return &h
}

func parseCookies(rhs string) []*http.Cookie {
	rawHString := strings.Trim(rhs, "'")
	values := strings.SplitN(rawHString, ":", 2)
	cookies, err := http.ParseCookie(values[1])
	if err != nil {
		panic(err)
	}
	return cookies
}

func Parse(rawCmd []string) (*Command, error) {
	var cmd Command
	cmd.Method = "GET"
	cmd.CurlCmd = "curl"
	i := 0
	for {
		if i >= len(rawCmd) {
			break
		}
		ctok := strings.Trim(rawCmd[i], "'")
		if ctok == "curl" {
			cmd.CurlCmd = ctok
			i++
		} else if strings.HasPrefix(ctok, "http") {
			slog.Debug("URL:", slog.String("url", ctok))
			cmd.URL = ctok
			i++
		} else if isFlag(ctok) {
			slog.Debug("parse", slog.String("flag", ctok))
			cmd.Flags = append(cmd.Flags, ctok)
			i++
		} else if isOption(ctok) {
			slog.Debug("parse", slog.String("option", ctok))
			if ctok == "-H" {
				if strings.HasPrefix(strings.Trim(strings.ToLower(rawCmd[i+1]), "'"), "cookie") {
					cmd.Cookies = parseCookies(rawCmd[i+1])
				} else {
					h := parseHeader(rawCmd[i+1])
					cmd.Headers = append(cmd.Headers, *h)
				}
			} else if ctok == "-X" {
				cmd.Method = rawCmd[i+1]
			} else {
				return nil, fmt.Errorf("unexpected option %s:%s", ctok, rawCmd[i+1])
			}
			i += 2
		} else {
			return nil, fmt.Errorf("unexpected token: %v", ctok)
			// i++
		}
	}
	return &cmd, nil
}

// TokenizeCurlCommand allows to split a curl command split on multiple lines
// just like a shell would do, it is mainly used when reading a curl command from file,
// as they are saved with newline breaks. While the function name might be very fancy,
// this is effectively a glorified `strings.Split(input, " ")`.
func TokenizeCurlCommand(input string) []string {
	var toks []string
	var lpos int // lpos keeps track of the last position where a token started
	// -H options have values specified inside quotes, since spaces are part
	// of the header value, we do not split on spaces inside.
	var isInsideQuotes bool
	for i := 0; i < len(input); i++ {
		cTok := input[i]

		if cTok == '\'' {
			isInsideQuotes = !isInsideQuotes
		}
		//  split on spaces and `\` (bash line terminator)
		if (cTok == ' ' && !isInsideQuotes) || cTok == '\\' {
			// ignore spaces/tabs
			if strings.TrimSpace(input[lpos:i]) != "" {
				tokenToInsert := input[lpos:i]
				slog.Debug(tokenToInsert)
				// ignore bash command newline separators
				if tokenToInsert != "\\\n" {
					toks = append(toks, strings.TrimSpace(tokenToInsert))
				}
			}
			lpos = i
		}
	}
	// Add the last token (command might not have trailing space)
	if strings.TrimSpace(input[lpos:]) != "" {
		toks = append(toks, strings.TrimSpace(input[lpos:]))
	}
	return toks
}
