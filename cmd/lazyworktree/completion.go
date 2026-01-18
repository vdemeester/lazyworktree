package main

import (
	"fmt"
	"os"

	_ "embed"

	urfavecli "github.com/urfave/cli/v2"
)

//go:embed templates/zsh_completion.zsh
var zshCompletion []byte

//go:embed templates/bash_completion.bash
var bashCompletion []byte

// completionCommand returns the completion subcommand definition.
func completionCommand() *urfavecli.Command {
	return &urfavecli.Command{
		Name:      "completion",
		Usage:     "Generate shell completion scripts",
		ArgsUsage: "<bash|zsh|fish>",
		Flags: []urfavecli.Flag{
			&urfavecli.BoolFlag{
				Name:  "code",
				Usage: "Output completion code instead of installation instructions",
			},
		},
		Action: handleCompletion,
	}
}

// handleCompletion handles the completion subcommand.
func handleCompletion(c *urfavecli.Context) error {
	if c.NArg() == 0 {
		return fmt.Errorf("usage: lazyworktree completion <bash|zsh|fish> [--code]")
	}

	shell := c.Args().First()
	switch shell {
	case "bash":
		_, _ = os.Stdout.WriteString(string(bashCompletion))
	case "zsh":
		_, _ = os.Stdout.WriteString(string(zshCompletion))
	// case "fish":
	// 	fmt.Print(fishCompletionScript)
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shell)
	}
	return nil
}
