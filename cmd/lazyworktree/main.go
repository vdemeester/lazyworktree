// Package main is the entry point for the lazyworktree application.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/chmouel/lazyworktree/internal/app"
	"github.com/chmouel/lazyworktree/internal/config"
	"github.com/chmouel/lazyworktree/internal/log"
	"github.com/chmouel/lazyworktree/internal/theme"
	"github.com/chmouel/lazyworktree/internal/utils"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	// Handle special flags that exit early before Kong parsing
	// This is needed because Kong requires a subcommand when subcommands are defined
	args := os.Args[1:]
	for _, arg := range args {
		if arg == "--version" || arg == "-v" {
			printVersion()
			return
		}
		if arg == "--show-syntax-themes" {
			printSyntaxThemes()
			return
		}
	}

	parseResult, err := ParseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	cli := parseResult.CLI
	ctx := parseResult.Context
	cmd := parseResult.Command
	initialFilter := parseResult.InitialFilter

	// If completion command was selected, run it immediately and exit
	if cmd == "completion" || cmd == "completion <shell>" {
		if ctx != nil {
			if err := ctx.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
		// ctx.Run() should have called ctx.Exit(0), but ensure we exit
		os.Exit(0)
	}

	// Handle special flags that exit early (in case they weren't caught above)
	if cli.Version {
		printVersion()
		return
	}
	if cli.ShowSyntaxThemes {
		printSyntaxThemes()
		return
	}

	// Handle subcommands
	if cmd != "" {
		switch cmd {
		case "wt-create":
			handleWtCreate(cli.WtCreate, cli.WorktreeDir, cli.ConfigFile, cli.Config)
			return
		case "wt-delete":
			handleWtDelete(cli.WtDelete, cli.WorktreeDir, cli.ConfigFile, cli.Config)
			return
		case "completion":
			// This should have been handled above, but just in case
			if ctx != nil {
				if err := ctx.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
			}
			return
		}
	}

	// Set up debug logging before loading config, so debug output is captured
	if cli.DebugLog != "" {
		expanded, err := utils.ExpandPath(cli.DebugLog)
		if err == nil {
			if err := log.SetFile(expanded); err != nil {
				fmt.Fprintf(os.Stderr, "Error opening debug log file %q: %v\n", expanded, err)
			}
		} else {
			if err := log.SetFile(cli.DebugLog); err != nil {
				fmt.Fprintf(os.Stderr, "Error opening debug log file %q: %v\n", cli.DebugLog, err)
			}
		}
	}

	cfg, err := config.LoadConfig(cli.ConfigFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	// If debug log wasn't set via flag, check if it's in the config
	// If it is, enable logging. If not, disable logging and discard buffer.
	if cli.DebugLog == "" {
		if cfg.DebugLog != "" {
			expanded, err := utils.ExpandPath(cfg.DebugLog)
			path := cfg.DebugLog
			if err == nil {
				path = expanded
			}
			if err := log.SetFile(path); err != nil {
				fmt.Fprintf(os.Stderr, "Error opening debug log file from config %q: %v\n", path, err)
			}
		} else {
			// No debug log configured, discard any buffered logs
			_ = log.SetFile("")
		}
	}

	if err := applyThemeConfig(cfg, cli.Theme); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		_ = log.Close()
		os.Exit(1)
	}
	if cli.SearchAutoSelect {
		cfg.SearchAutoSelect = true
	}

	if err := applyWorktreeDirConfig(cfg, cli.WorktreeDir); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		_ = log.Close()
		os.Exit(1)
	}

	if cli.DebugLog != "" {
		expanded, err := utils.ExpandPath(cli.DebugLog)
		if err == nil {
			cfg.DebugLog = expanded
		} else {
			cfg.DebugLog = cli.DebugLog
		}
	}

	// Apply CLI config overrides (highest precedence)
	if len(cli.Config) > 0 {
		if err := cfg.ApplyCLIOverrides(cli.Config); err != nil {
			fmt.Fprintf(os.Stderr, "Error applying config overrides: %v\n", err)
			_ = log.Close()
			os.Exit(1)
		}
	}

	model := app.NewModel(cfg, initialFilter)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	_, err = p.Run()
	model.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		_ = log.Close()
		os.Exit(1)
	}

	selectedPath := model.GetSelectedPath()
	if cli.OutputSelection != "" {
		expanded, err := utils.ExpandPath(cli.OutputSelection)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error expanding output-selection: %v\n", err)
			_ = log.Close()
			os.Exit(1)
		}
		const defaultDirPerms = 0o750
		if err := os.MkdirAll(filepath.Dir(expanded), defaultDirPerms); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output-selection dir: %v\n", err)
			_ = log.Close()
			os.Exit(1)
		}
		data := ""
		if selectedPath != "" {
			data = selectedPath + "\n"
		}
		const defaultFilePerms = 0o600
		if err := os.WriteFile(expanded, []byte(data), defaultFilePerms); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output-selection: %v\n", err)
			_ = log.Close()
			os.Exit(1)
		}
		return
	}
	if selectedPath != "" {
		fmt.Println(selectedPath)
	}
	if err := log.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Error closing debug log: %v\n", err)
	}
}

// applyWorktreeDirConfig applies the worktree directory configuration.
// This ensures the same path expansion logic is used in both TUI and CLI modes.
func applyWorktreeDirConfig(cfg *config.AppConfig, worktreeDirFlag string) error {
	switch {
	case worktreeDirFlag != "":
		expanded, err := utils.ExpandPath(worktreeDirFlag)
		if err != nil {
			return fmt.Errorf("error expanding worktree-dir: %w", err)
		}
		cfg.WorktreeDir = expanded
	case cfg.WorktreeDir != "":
		expanded, err := utils.ExpandPath(cfg.WorktreeDir)
		if err == nil {
			cfg.WorktreeDir = expanded
		}
	default:
		home, _ := os.UserHomeDir()
		cfg.WorktreeDir = filepath.Join(home, ".local", "share", "worktrees")
	}
	return nil
}

func printSyntaxThemes() {
	names := theme.AvailableThemes()
	sort.Strings(names)
	fmt.Println("Available syntax themes (delta --syntax-theme defaults):")
	for _, name := range names {
		fmt.Printf("  %-16s -> %s\n", name, config.SyntaxThemeForUITheme(name))
	}
}

// printVersion prints version information.
func printVersion() {
	v := version
	c := commit
	d := date
	b := builtBy

	if c == "none" || b == "unknown" {
		if info, ok := debug.ReadBuildInfo(); ok {
			if c == "none" {
				for _, setting := range info.Settings {
					if setting.Key == "vcs.revision" {
						c = setting.Value
					}
				}
			}
			if b == "unknown" {
				b = info.GoVersion
			}
		}
	}

	fmt.Printf("lazyworktree version %s\ncommit: %s\nbuilt at: %s\nbuilt by: %s\n", v, c, d, b)
}

// applyThemeConfig applies theme configuration from command line flag.
func applyThemeConfig(cfg *config.AppConfig, themeName string) error {
	if themeName == "" {
		return nil
	}

	normalized := config.NormalizeThemeName(themeName)
	if normalized == "" {
		return fmt.Errorf("unknown theme %q", themeName)
	}

	cfg.Theme = normalized
	if !cfg.GitPagerArgsSet && filepath.Base(cfg.GitPager) == "delta" {
		cfg.GitPagerArgs = config.DefaultDeltaArgsForTheme(normalized)
	}

	return nil
}
