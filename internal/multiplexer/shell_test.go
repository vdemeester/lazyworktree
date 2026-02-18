package multiplexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShellQuote(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "''",
		},
		{
			name:  "simple string",
			input: "hello",
			want:  "'hello'",
		},
		{
			name:  "string with single quote",
			input: "it's",
			want:  "'it'\"'\"'s'",
		},
		{
			name:  "string with multiple single quotes",
			input: "it's Bob's",
			want:  "'it'\"'\"'s Bob'\"'\"'s'",
		},
		{
			name:  "string with spaces",
			input: "hello world",
			want:  "'hello world'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShellQuote(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExportEnvCommand(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{
			name: "empty env",
			env:  map[string]string{},
			want: "",
		},
		{
			name: "single env var",
			env:  map[string]string{"FOO": "bar"},
			want: "export FOO='bar';",
		},
		{
			name: "multiple env vars sorted",
			env: map[string]string{
				"ZZZ": "last",
				"AAA": "first",
				"MMM": "middle",
			},
			want: "export AAA='first'; export MMM='middle'; export ZZZ='last';",
		},
		{
			name: "env var with special chars",
			env:  map[string]string{"PATH": "/usr/bin:/bin"},
			want: "export PATH='/usr/bin:/bin';",
		},
		{
			name: "env var with quote",
			env:  map[string]string{"MSG": "it's fine"},
			want: "export MSG='it'\"'\"'s fine';",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExportEnvCommand(tt.env)
			assert.Equal(t, tt.want, got)
		})
	}
}
