package cmd

import (
	"flag"
	"net/http"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestMain(m *testing.M) {
	flag.Parse()
	verbose := flag.Lookup("test.v") != nil && flag.Lookup("test.v").Value.String() == "true"

	if verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		log.Debug().Msg("Debug logging enabled")
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	os.Exit(m.Run())
}

func TestSplit(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "simple curl command no quotes",
			input: `curl https://rcastellotti.dev`,
			want:  []string{"curl", "https://rcastellotti.dev"},
		},
		{
			name:  "simple command with parameters in url",
			input: "curl 'https://rcastellotti/dev/api/v0/getBear?species=grizzly&name=yogi'",
			want:  []string{"curl", `'https://rcastellotti/dev/api/v0/getBear?species=grizzly&name=yogi'`},
		},
		{
			name:  "simple curl command url in quotes",
			input: `curl "https://rcastellotti.dev"`,
			want:  []string{"curl", `"https://rcastellotti.dev"`},
		},
		{
			name:  "simple curl command url in quotes flag and option",
			input: `curl -X GET --compressed "https://rcastellotti.dev"`,
			want:  []string{"curl", "-X", "GET", "--compressed", `"https://rcastellotti.dev"`},
		},
		{
			name: "simple curl command, url in quotes flag and option and newlines",
			input: `curl \
					-X GET "https://rcastellotti.dev" \
					--compressed`,
			want: []string{"curl", "-X", "GET", `"https://rcastellotti.dev"`, "--compressed"},
		},
		{
			name:  "simple curl command random spaces",
			input: ` curl      https://rcastellotti.dev    `,
			want:  []string{"curl", "https://rcastellotti.dev"},
		},
		{
			name:  "curl command with header with quotes",
			input: `curl https://rcastellotti.dev -H 'sec-ch-ua: "Chromium";v="134", "Not:A-Brand";v="24", "Google Chrome";v="134"'`,
			want:  []string{"curl", "https://rcastellotti.dev", "-H", `'sec-ch-ua: "Chromium";v="134", "Not:A-Brand";v="24", "Google Chrome";v="134"'`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := SplitCurlCommand(tc.input)

			if len(got) != len(tc.want) {
				t.Fatalf("got %d tokens, want %d", len(got), len(tc.want))
			}

			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("pos: %d: got %+q, want %+q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  *Command
	}{
		{
			name:  "real command parameters in url",
			input: []string{"curl", "-s", "-X", "GET", "https://rcastellotti/dev/api/v0/getBear?species=grizzly&name=yogi"},
			want: &Command{
				CurlKeyword: "curl",
				URL:         "https://rcastellotti/dev/api/v0/getBear?species=grizzly&name=yogi",
				Flags:       []string{"-s"},
				Headers:     []string{},
				Options:     map[string]string{"-X": "GET"},
			},
		},
		{
			name:  "real command parameters with b option for cookies",
			input: []string{"curl", "https://rcastellotti.dev", "-b", "kitten=siberian; bear=grizzly"},
			want: &Command{
				CurlKeyword: "curl",
				URL:         "https://rcastellotti.dev",
				Cookies: []*http.Cookie{
					{Name: "kitten", Value: "siberian"},
					{Name: "bear", Value: "grizzly"},
				},
			},
		},
		{
			name:  "real command parameters with -H for cookies",
			input: []string{"curl", "https://rcastellotti.dev", "-H", `'cookie: kitten=siberian; bear=grizzly;'`},
			want: &Command{
				CurlKeyword: "curl",
				URL:         "https://rcastellotti.dev",
				Cookies: []*http.Cookie{
					{Name: "kitten", Value: "siberian"},
					{Name: "bear", Value: "grizzly"},
				},
			},
		},
		{
			name: "real command parameters with -H  and -b for cookies",
			input: []string{
				"curl",
				"https://rcastellotti.dev",
				"-H", `'cookie: kitten=siberian; bear=grizzly;'`,
				"-b", `fish=puffer`,
			},
			want: &Command{
				CurlKeyword: "curl",
				URL:         "https://rcastellotti.dev",
				Cookies: []*http.Cookie{
					{Name: "kitten", Value: "siberian"},
					{Name: "bear", Value: "grizzly"},
					{Name: "fish", Value: "puffer"},
				},
			},
		},
		{
			name:  "real command data raw",
			input: []string{"curl", "https://rcastellotti.dev", "--data-raw", `'{"bearName":"yogi"}'`},
			want: &Command{
				CurlKeyword: "curl",
				URL:         "https://rcastellotti.dev",
				Options: map[string]string{
					"--data-raw": `'{"bearName":"yogi"}'`,
				},
			},
		},
		{
			name:  "real command with flag",
			input: []string{"curl", "https://rcastellotti.dev", "--compressed"},
			want: &Command{
				CurlKeyword: "curl",
				URL:         "https://rcastellotti.dev",
				Flags:       []string{"--compressed"},
			},
		},
		{
			name:  "real command with header (non cookie)",
			input: []string{"curl", "https://rcastellotti.dev", "-H", "sec-ch-ua-platform: \"macOS\""},
			want: &Command{
				CurlKeyword: "curl",
				URL:         "https://rcastellotti.dev",
				Headers:     []string{"sec-ch-ua-platform: \"macOS\""},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			log.Debug().Strs("tc.name", tc.input).Msg("command")

			got, err := Parse(tc.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if got.CurlKeyword != tc.want.CurlKeyword {
				t.Errorf("CurlKeyword: got %q, want %q", got.CurlKeyword, tc.want.CurlKeyword)
			}

			if got.URL != tc.want.URL {
				t.Errorf("URL: got %q, want %q", got.URL, tc.want.URL)
			}

			if tc.want.Cookies == nil && len(got.Cookies) > 0 {
				t.Errorf("no cookies expected, got: %v", got.Cookies)
			} else if tc.want.Cookies != nil {
				if len(got.Cookies) != len(tc.want.Cookies) {
					t.Fatalf("cookie count mismatch: got %d, want %d", len(got.Cookies), len(tc.want.Cookies))
				}

				for k := range tc.want.Cookies {
					if k >= len(got.Cookies) || got.Cookies[k].Name != tc.want.Cookies[k].Name || got.Cookies[k].Value != tc.want.Cookies[k].Value {
						t.Fatalf("cookie %d: got {Name: %q, Value: %q}, want {Name: %q, Value: %q}",
							k, got.Cookies[k].Name, got.Cookies[k].Value, tc.want.Cookies[k].Name, tc.want.Cookies[k].Value)
					}
				}
			}

			if tc.want.Headers == nil && len(got.Headers) > 0 {
				t.Errorf("no headers expected, got: %v", got.Headers)
			} else if tc.want.Headers != nil {
				if len(got.Headers) != len(tc.want.Headers) {
					t.Fatalf("header count mismatch: got %d, want %d", len(got.Headers), len(tc.want.Headers))
				}

				for i, h := range tc.want.Headers {
					if got.Headers[i] != h {
						t.Fatalf("header %d: got %q, want %q", i, got.Headers[i], h)
					}
				}
			}

			if tc.want.Flags == nil && len(got.Flags) > 0 {
				t.Errorf("no flags expected, got: %v", got.Flags)
			} else if tc.want.Flags != nil {
				if len(got.Flags) != len(tc.want.Flags) {
					t.Fatalf("flag count mismatch: got %d, want %d", len(got.Flags), len(tc.want.Flags))
				}

				for i, f := range tc.want.Flags {
					if got.Flags[i] != f {
						t.Fatalf("flag %d: got %q, want %q", i, got.Flags[i], f)
					}
				}
			}

			if tc.want.Options == nil && len(got.Options) > 0 {
				t.Errorf("no options expected, got: %v", got.Options)
			} else if tc.want.Options != nil {
				if len(got.Options) != len(tc.want.Options) {
					t.Fatalf("option count mismatch: got %d, want %d", len(got.Options), len(tc.want.Options))
				}

				for ok, ov := range tc.want.Options {
					if got.Options[ok] != ov {
						t.Fatalf("option %s: got %q, want %q", ov, got.Options[ov], ok)
					}
				}
			}
		})
	}
}

func TestCmdString(t *testing.T) {
	tests := []struct {
		name  string
		input *Command
		want  string
	}{
		{
			name: "simple curl commanand with flag",
			input: &Command{
				CurlKeyword: "curl",
				URL:         "https://rcastellotti.dev",
				Headers:     []string{},
				Cookies:     []*http.Cookie{},
				Options:     map[string]string{},
				Flags:       []string{},
			},
			want: `curl \
	"https://rcastellotti.dev"`,
		},
		{
			name: "simple curl commanand with flag",
			input: &Command{
				CurlKeyword: "curl",
				URL:         "https://rcastellotti.dev",
				Headers:     []string{},
				Cookies:     []*http.Cookie{},
				Options:     map[string]string{},
				Flags:       []string{"--compressed"},
			},
			want: `curl \
	--compressed \
	"https://rcastellotti.dev"`,
		},
		{
			name: "simple curl commanand with cookie",
			input: &Command{
				CurlKeyword: "curl",
				URL:         "https://rcastellotti.dev",
				Headers:     []string{},
				Cookies: []*http.Cookie{
					{Name: "bear", Value: "grizzly"},
				},
				Options: map[string]string{},
				Flags:   []string{"--compressed"},
			},
			want: `curl \
	-b 'bear=grizzly' \
	--compressed \
	"https://rcastellotti.dev"`,
		},
		{
			name: "simple curl commanand with header",
			input: &Command{
				CurlKeyword: "curl",
				URL:         "https://rcastellotti.dev",
				Headers:     []string{"sec-ch-ua: \"Chromium\";v=\"134\", \"Not:A-Brand\";v=\"24\", \"Google Chrome\";v=\"134\""},
				Cookies: []*http.Cookie{
					{Name: "bear", Value: "grizzly"},
				},
				Options: map[string]string{},
				Flags:   []string{"--compressed"},
			},
			want: `curl \
	-H 'sec-ch-ua: "Chromium";v="134", "Not:A-Brand";v="24", "Google Chrome";v="134"' \
	-b 'bear=grizzly' \
	--compressed \
	"https://rcastellotti.dev"`,
		},
		{
			name: "simple curl commanand with option",
			input: &Command{
				CurlKeyword: "curl",
				URL:         "https://rcastellotti.dev",
				Options: map[string]string{
					"--data-raw": "{\"bearName\":\"yogi\"}",
				},
			},
			want: `curl \
	--data-raw '{"bearName":"yogi"}' \
	"https://rcastellotti.dev"`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			log.Debug().Str("tc.name", tc.name).Msg("command")

			got := tc.input.String()
			if got != tc.want {
				t.Errorf("error printing command: \ngot:\n%s, \nwant:\n%s", got, tc.want)
			}
		})
	}
}
