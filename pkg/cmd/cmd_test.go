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
			input: "curl -s  -X GET 'https://www.mediawiki.org/w/api.php?action=unblock&id=105&format=json'",
			want:  []string{"curl", "-s", "-X", "GET", "https://www.mediawiki.org/w/api.php?action=unblock&id=105&format=json"},
		},
		{
			name:  "simple curl command url in single quotes",
			input: `curl "https://rcastellotti.dev"`,
			want:  []string{"curl", "https://rcastellotti.dev"},
		},
		{
			name:  "simple curl command url in double quotes",
			input: `curl "https://rcastellotti.dev"`,
			want:  []string{"curl", "https://rcastellotti.dev"},
		},
		{
			name:  "simple curl command url in quotes flag and option",
			input: `curl -X GET --compressed "https://rcastellotti.dev"`,
			want:  []string{"curl", "-X", "GET", "--compressed", "https://rcastellotti.dev"},
		},
		{
			name: "simple curl command, url in quotes flag and option and newline",
			input: `curl \
					-X GET "https://rcastellotti.dev" \
					--compressed`,
			want: []string{"curl", "-X", "GET", "https://rcastellotti.dev", "--compressed"},
		},
		{
			name:  "simple curl command no quotes random spaces",
			input: ` curl      https://rcastellotti.dev    `,
			want:  []string{"curl", "https://rcastellotti.dev"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := SplitCurlCommand(tc.input)

			if len(got) != len(tc.want) {
				t.Errorf("got %+q, want %+q", got, tc.want)
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
			input: []string{"curl", "-s", "-X", "GET", "https://www.mediawiki.org/w/api.php?action=unblock&id=105&format=json"},
			want: &Command{
				CurlKeyword: "curl",
				Method:      "GET",
				URL:         "https://www.mediawiki.org/w/api.php?action=unblock&id=105&format=json",
				Flags:       []string{"-s"},
				Headers:     []string{},
			},
		},
		{
			name: "real command parameters with b option",
			input: []string{"curl",
				"https://www.colruyt.be/en/produits/13890",
				"-H", "cache-control: max-age=0",
				"-b", "colruytlanguage=fr; TS0138f243=016303f955ca6a9699369bbcd652907f6863d51c141583066034d3b24041417b00030a8bce0d264aa685fc9d6ba1b88b20c2682d44",
			},
			want: &Command{
				CurlKeyword: "curl",
				Method:      "GET",
				URL:         "https://www.colruyt.be/en/produits/13890",
				Headers:     []string{"cache-control: max-age=0"},
				Cookies: []*http.Cookie{
					{
						Name:  "colruytlanguage",
						Value: "fr",
					},
					{
						Name:  "TS0138f243",
						Value: "016303f955ca6a9699369bbcd652907f6863d51c141583066034d3b24041417b00030a8bce0d264aa685fc9d6ba1b88b20c2682d44",
					},
				}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			got, err := Parse(tc.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check CurlKeyword
			if got.CurlKeyword != tc.want.CurlKeyword {
				t.Errorf("CurlKeyword: got %q, want %q", got.CurlKeyword, tc.want.CurlKeyword)
			}

			// Check Method
			if got.Method != tc.want.Method {
				t.Errorf("Method: got %q, want %q", got.Method, tc.want.Method)
			}

			// Check URL
			if got.URL != tc.want.URL {
				t.Errorf("URL: got %q, want %q", got.URL, tc.want.URL)
			}

			// Check Cookies
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

			// Check Headers
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

			// Check Flags
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
		})
	}
}
