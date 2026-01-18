// Package main provides CLI flag definitions for lazyworktree.
package main

import (
	"fmt"
	"strings"

	"github.com/chmouel/lazyworktree/internal/theme"
	urfavecli "github.com/urfave/cli/v2"
)

// globalFlags returns all global flags for the application.
// Note: --version is provided automatically by urfave/cli via App.Version
func globalFlags() []urfavecli.Flag {
	return []urfavecli.Flag{
		&urfavecli.StringFlag{
			Name:    "worktree-dir",
			Aliases: []string{"w"},
			Usage:   "Override the default worktree root directory",
		},
		&urfavecli.StringFlag{
			Name:  "debug-log",
			Usage: "Path to debug log file",
		},
		&urfavecli.StringFlag{
			Name:  "output-selection",
			Usage: "Write selected worktree path to a file",
		},
		&urfavecli.StringFlag{
			Name:    "theme",
			Aliases: []string{"t"},
			Usage:   "Override the UI theme",
		},
		&urfavecli.BoolFlag{
			Name:  "search-auto-select",
			Usage: "Start with filter focused",
		},
		&urfavecli.BoolFlag{
			Name:  "show-syntax-themes",
			Usage: "List available delta syntax themes",
		},
		&urfavecli.StringFlag{
			Name:  "config-file",
			Usage: "Path to configuration file",
		},
		&urfavecli.StringSliceFlag{
			Name:    "config",
			Aliases: []string{"C"},
			Usage:   "Override config values (repeatable): --config=lw.key=value",
		},
	}
}

// completeGlobalFlags provides basic completion for global flags.
// Note: urfave/cli/v2 has limited flag completion support compared to kong.
func completeGlobalFlags(c *urfavecli.Context) {
	// Complete subcommands if no args yet
	if c.NArg() == 0 {
		for _, cmd := range c.App.Commands {
			fmt.Println(cmd.Name)
		}
	}
}

// Note: The following completion helper functions are preserved for potential
// future use with custom completion scripts, but are not currently used by
// urfave/cli/v2's basic completion system.

// suggestConfigKeys returns config key suggestions matching the prefix.
// Returns suggestions in the format "lw.key=" for completion.
//
//nolint:unused // Preserved for potential future completion enhancements
func suggestConfigKeys(prefix string) []string {
	allKeys := []string{
		"theme", "worktree_dir", "sort_mode", "auto_fetch_prs", "auto_refresh",
		"refresh_interval", "search_auto_select", "fuzzy_finder_input", "show_icons",
		"max_untracked_diffs", "max_diff_chars", "max_name_length", "git_pager",
		"git_pager_args", "git_pager_interactive", "pager", "editor", "trust_mode",
		"debug_log", "init_commands", "terminate_commands", "merge_method",
		"issue_branch_name_template", "pr_branch_name_template", "branch_name_script",
		"session_prefix", "palette_mru", "palette_mru_limit",
	}

	var matches []string
	for _, key := range allKeys {
		if prefix == "" || strings.HasPrefix(key, prefix) {
			matches = append(matches, "lw."+key+"=")
		}
	}
	return matches
}

// suggestConfigValues returns value suggestions for a given config key.
//
//nolint:unused // Preserved for potential future completion enhancements
func suggestConfigValues(key string) []string {
	switch key {
	case "theme":
		return theme.AvailableThemes()
	case "sort_mode":
		return []string{"switched", "active", "path"}
	case "merge_method":
		return []string{"rebase", "merge"}
	case "trust_mode":
		return []string{"tofu", "never", "always"}
	case "auto_fetch_prs", "auto_refresh", "search_auto_select", "fuzzy_finder_input",
		"show_icons", "git_pager_interactive", "palette_mru":
		return []string{"true", "false"}
	default:
		return nil
	}
}
