package multiplexer

import (
	"fmt"
	"sort"
	"strings"
)

// ShellQuote quotes a string for use in a shell command.
// Returns an empty quoted string for empty input.
func ShellQuote(input string) string {
	if input == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(input, "'", "'\"'\"'") + "'"
}

// ExportEnvCommand builds a shell command string that exports environment variables.
// Returns empty string if env is empty.
func ExportEnvCommand(env map[string]string) string {
	if len(env) == 0 {
		return ""
	}
	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("export %s=%s;", key, ShellQuote(env[key])))
	}
	return strings.Join(parts, " ")
}
