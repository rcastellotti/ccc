package main

import "testing"

func TestSplit(t *testing.T) {
	tests := []struct {
		input string
		want  []string
		name  string
	}{
		{
			input: `curl https://rcastellotti.dev`,
			want:  []string{"curl", "https://rcastellotti.dev"},
			name:  "simple curl command, no quotes",
		},
		{
			input: `curl "https://rcastellotti.dev"`,
			want:  []string{"curl", "https://rcastellotti.dev"},
			name:  "simple curl command, url in single quotes",
		},
		{
			input: `curl "https://rcastellotti.dev"`,
			want:  []string{"curl", "https://rcastellotti.dev"},
			name:  "simple curl command, url in double quotes",
		},
		{
			input: `curl -X GET --compressed "https://rcastellotti.dev"`,
			want:  []string{"curl", "-X", "GET", "--compressed", "https://rcastellotti.dev"},
			name:  "simple curl command, url in quotes, flag and option",
		},
		{
			input: `curl \
					-X GET "https://rcastellotti.dev" \
					--compressed`,
			want: []string{"curl", "-X", "GET", "https://rcastellotti.dev", "--compressed"},
			name: "simple curl command, url in quotes, flag and option and newline",
		},
		{
			input: ` curl      https://rcastellotti.dev    `,
			want:  []string{"curl", "https://rcastellotti.dev"},
			name:  "simple curl command, no quotes, random spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitCurlCommand(tt.input)

			if len(result) != len(tt.want) {
				t.Errorf("got %+q, want %+q", result, tt.want)
			}

			for i := range result {
				if result[i] != tt.want[i] {
					t.Errorf("pos: %d: got %+q, want %+q", i, result[i], tt.want[i])
				}
			}
		})
	}
}
