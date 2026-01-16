package main

import (
	"strings"
	"testing"
)

func TestBuildBooleanFlagsMap(t *testing.T) {
	_, parser, err := NewParser()
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	booleanFlags := buildBooleanFlagsMap(parser)

	// Check that boolean flags are detected
	if !booleanFlags["--search-auto-select"] {
		t.Error("expected --search-auto-select to be detected as boolean")
	}
	if !booleanFlags["--version"] {
		t.Error("expected --version to be detected as boolean")
	}
	if !booleanFlags["-v"] {
		t.Error("expected -v to be detected as boolean")
	}

	// Check that string flags are not in the map
	if booleanFlags["--worktree-dir"] {
		t.Error("expected --worktree-dir to not be in boolean flags map")
	}
	if booleanFlags["--theme"] {
		t.Error("expected --theme to not be in boolean flags map")
	}
}

func TestDetectSubcommand(t *testing.T) {
	_, parser, err := NewParser()
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	booleanFlags := buildBooleanFlagsMap(parser)
	isBooleanFlag := func(flagArg string) bool {
		flagName := strings.SplitN(flagArg, "=", 2)[0]
		return booleanFlags[flagName]
	}

	tests := []struct {
		name         string
		args         []string
		expectSubcmd bool
	}{
		{
			name:         "no subcommand",
			args:         []string{"--theme", "nord"},
			expectSubcmd: false,
		},
		{
			name:         "subcommand present",
			args:         []string{"wt-create"},
			expectSubcmd: true,
		},
		{
			name:         "subcommand after flags",
			args:         []string{"--theme", "nord", "wt-create"},
			expectSubcmd: true,
		},
		{
			name:         "subcommand as flag value",
			args:         []string{"--worktree-dir", "wt-delete"},
			expectSubcmd: false,
		},
		{
			name:         "subcommand after boolean flag",
			args:         []string{"--search-auto-select", "wt-create"},
			expectSubcmd: true,
		},
		{
			name:         "completion subcommand",
			args:         []string{"completion", "bash"},
			expectSubcmd: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectSubcommand(tt.args, isBooleanFlag)
			if result != tt.expectSubcmd {
				t.Errorf("expected hasSubcommand=%v, got %v", tt.expectSubcmd, result)
			}
		})
	}
}

func TestExtractFilterArgs(t *testing.T) {
	_, parser, err := NewParser()
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	booleanFlags := buildBooleanFlagsMap(parser)
	isBooleanFlag := func(flagArg string) bool {
		flagName := strings.SplitN(flagArg, "=", 2)[0]
		return booleanFlags[flagName]
	}

	tests := []struct {
		name           string
		args           []string
		expectedFilter string
	}{
		{
			name:           "no filter",
			args:           []string{"--theme", "nord"},
			expectedFilter: "",
		},
		{
			name:           "simple filter",
			args:           []string{"myfilter"},
			expectedFilter: "myfilter",
		},
		{
			name:           "filter after boolean flag",
			args:           []string{"--search-auto-select", "myfilter"},
			expectedFilter: "myfilter",
		},
		{
			name:           "filter after string flag",
			args:           []string{"--theme", "nord", "myfilter"},
			expectedFilter: "myfilter",
		},
		{
			name:           "filter with boolean flag in between",
			args:           []string{"filter1", "--search-auto-select", "filter2"},
			expectedFilter: "filter1 filter2",
		},
		{
			name:           "filter stops at subcommand",
			args:           []string{"filter1", "wt-create"},
			expectedFilter: "filter1",
		},
		{
			name:           "flag value not treated as filter",
			args:           []string{"--worktree-dir", "somepath", "myfilter"},
			expectedFilter: "myfilter",
		},
		{
			name:           "multiple filter words",
			args:           []string{"filter", "word1", "word2"},
			expectedFilter: "filter word1 word2",
		},
		{
			name:           "flag with equals sign",
			args:           []string{"--theme=nord", "myfilter"},
			expectedFilter: "myfilter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFilterArgs(tt.args, isBooleanFlag)
			if result != tt.expectedFilter {
				t.Errorf("expected filter %q, got %q", tt.expectedFilter, result)
			}
		})
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectError   bool
		expectCommand string
		expectFilter  string
		expectSubcmd  bool
	}{
		{
			name:          "no args - TUI mode",
			args:          []string{},
			expectError:   false,
			expectCommand: "",
			expectFilter:  "",
			expectSubcmd:  false,
		},
		{
			name:          "with filter",
			args:          []string{"myfilter"},
			expectError:   false,
			expectCommand: "",
			expectFilter:  "myfilter",
			expectSubcmd:  false,
		},
		{
			name:          "wt-create subcommand",
			args:          []string{"wt-create", "--from-branch", "main"},
			expectError:   false,
			expectCommand: "wt-create",
			expectSubcmd:  true,
		},
		{
			name:          "wt-delete subcommand",
			args:          []string{"wt-delete", "/path/to/worktree"},
			expectError:   false,
			expectCommand: "wt-delete <worktree-path>",
			expectSubcmd:  true,
		},
		{
			name:          "completion subcommand",
			args:          []string{"completion", "bash"},
			expectError:   false,
			expectCommand: "completion <shell>",
			expectSubcmd:  true,
		},
		{
			name:          "filter with boolean flag",
			args:          []string{"--search-auto-select", "myfilter"},
			expectError:   false,
			expectCommand: "",
			expectFilter:  "myfilter",
			expectSubcmd:  false,
		},
		{
			name:          "flag value not treated as subcommand",
			args:          []string{"--worktree-dir", "wt-delete"},
			expectError:   false,
			expectCommand: "",
			expectSubcmd:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseArgs(tt.args)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.expectError {
				if result.Command != tt.expectCommand {
					t.Errorf("expected command %q, got %q", tt.expectCommand, result.Command)
				}
				if result.InitialFilter != tt.expectFilter {
					t.Errorf("expected filter %q, got %q", tt.expectFilter, result.InitialFilter)
				}
				if result.HasSubcommand != tt.expectSubcmd {
					t.Errorf("expected hasSubcommand=%v, got %v", tt.expectSubcmd, result.HasSubcommand)
				}
				if result.CLI == nil {
					t.Error("expected CLI to be non-nil")
				}
				if result.Parser == nil {
					t.Error("expected Parser to be non-nil")
				}
			}
		})
	}
}

func TestSuggestConfigKeys(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		expected []string
	}{
		{
			name:     "empty prefix",
			prefix:   "",
			expected: []string{"lw.theme=", "lw.worktree_dir=", "lw.sort_mode="},
		},
		{
			name:     "prefix match",
			prefix:   "theme",
			expected: []string{"lw.theme="},
		},
		{
			name:     "prefix match multiple",
			prefix:   "git_",
			expected: []string{"lw.git_pager=", "lw.git_pager_args=", "lw.git_pager_interactive="},
		},
		{
			name:     "no match",
			prefix:   "nonexistent",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := suggestConfigKeys(tt.prefix)
			if len(result) < len(tt.expected) {
				t.Errorf("expected at least %d results, got %d", len(tt.expected), len(result))
			}
			// Check that all expected items are in result
			for _, exp := range tt.expected {
				found := false
				for _, r := range result {
					if r == exp {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected %q in results, but not found. Got: %v", exp, result)
				}
			}
		})
	}
}

func TestSuggestConfigValues(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected []string
	}{
		{
			name:     "theme key",
			key:      "theme",
			expected: []string{"dracula", "nord", "monokai"},
		},
		{
			name:     "sort_mode key",
			key:      "sort_mode",
			expected: []string{"switched", "active", "path"},
		},
		{
			name:     "merge_method key",
			key:      "merge_method",
			expected: []string{"rebase", "merge"},
		},
		{
			name:     "trust_mode key",
			key:      "trust_mode",
			expected: []string{"tofu", "never", "always"},
		},
		{
			name:     "boolean key",
			key:      "auto_fetch_prs",
			expected: []string{"true", "false"},
		},
		{
			name:     "unknown key",
			key:      "unknown_key",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := suggestConfigValues(tt.key)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil for unknown key, got %v", result)
				}
				return
			}
			if len(result) < len(tt.expected) {
				t.Errorf("expected at least %d results, got %d", len(tt.expected), len(result))
			}
			// Check that all expected items are in result
			for _, exp := range tt.expected {
				found := false
				for _, r := range result {
					if r == exp {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected %q in results, but not found. Got: %v", exp, result)
				}
			}
		})
	}
}
