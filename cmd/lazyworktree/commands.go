// Package main provides CLI command definitions for lazyworktree.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chmouel/lazyworktree/internal/cli"
	"github.com/chmouel/lazyworktree/internal/log"
	urfavecli "github.com/urfave/cli/v2"
)

// wtCreateCommand returns the wt-create subcommand definition.
func wtCreateCommand() *urfavecli.Command {
	return &urfavecli.Command{
		Name:   "wt-create",
		Usage:  "Create a new worktree",
		Before: validateWtCreateFlags,
		Action: handleWtCreateAction,
		Flags: []urfavecli.Flag{
			&urfavecli.StringFlag{
				Name:  "from-branch",
				Usage: "Create worktree from branch (defaults to current branch)",
			},
			&urfavecli.IntFlag{
				Name:  "from-pr",
				Usage: "Create worktree from PR number",
			},
			&urfavecli.StringFlag{
				Name:  "name",
				Usage: "Name for the new worktree/branch (defaults to sanitised source branch name)",
			},
			&urfavecli.BoolFlag{
				Name:  "with-change",
				Usage: "Carry over uncommitted changes to the new worktree",
			},
			&urfavecli.BoolFlag{
				Name:  "silent",
				Usage: "Suppress progress messages",
			},
		},
	}
}

// wtDeleteCommand returns the wt-delete subcommand definition.
func wtDeleteCommand() *urfavecli.Command {
	return &urfavecli.Command{
		Name:      "wt-delete",
		Usage:     "Delete a worktree",
		ArgsUsage: "[worktree-path]",
		Action:    handleWtDeleteAction,
		Flags: []urfavecli.Flag{
			&urfavecli.BoolFlag{
				Name:  "no-branch",
				Usage: "Skip branch deletion",
			},
			&urfavecli.BoolFlag{
				Name:  "silent",
				Usage: "Suppress progress messages",
			},
		},
	}
}

// validateWtCreateFlags validates mutual exclusivity rules for wt-create flags.
func validateWtCreateFlags(c *urfavecli.Context) error {
	fromBranch := c.String("from-branch")
	fromPR := c.Int("from-pr")
	name := c.String("name")
	withChange := c.Bool("with-change")

	// Mutual exclusivity: from-branch and from-pr
	if fromBranch != "" && fromPR > 0 {
		return fmt.Errorf("--from-branch and --from-pr are mutually exclusive")
	}

	// Name cannot be used with from-pr
	if name != "" && fromPR > 0 {
		return fmt.Errorf("--name cannot be used with --from-pr")
	}

	// with-change cannot be used with from-pr
	if withChange && fromPR > 0 {
		return fmt.Errorf("--with-change cannot be used with --from-pr")
	}

	return nil
}

// handleWtCreateAction handles the wt-create subcommand action.
func handleWtCreateAction(c *urfavecli.Context) error {
	ctx := context.Background()

	// Load config with global flags
	cfg, err := loadCLIConfig(
		c.String("config-file"),
		c.String("worktree-dir"),
		c.StringSlice("config"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}

	gitSvc := newCLIGitService(cfg)

	// Extract command-specific flags
	fromPR := c.Int("from-pr")
	fromBranch := c.String("from-branch")
	name := c.String("name")
	withChange := c.Bool("with-change")
	silent := c.Bool("silent")

	// Execute appropriate operation
	var opErr error
	if fromPR > 0 {
		// Create from PR
		opErr = cli.CreateFromPR(ctx, gitSvc, cfg, fromPR, silent)
	} else {
		// Create from branch (either specified or current)
		sourceBranch := fromBranch

		// If no branch specified, use current branch
		if sourceBranch == "" {
			currentBranch, err := gitSvc.GetCurrentBranch(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				fmt.Fprintf(os.Stderr, "Hint: Specify a branch explicitly with --from-branch\n")
				_ = log.Close()
				return err
			}
			sourceBranch = currentBranch

			if !silent {
				fmt.Fprintf(os.Stderr, "Creating worktree from current branch: %s\n", sourceBranch)
			}
		}

		opErr = cli.CreateFromBranch(ctx, gitSvc, cfg, sourceBranch, name, withChange, silent)
	}

	if opErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", opErr)
		_ = log.Close()
		return opErr
	}

	_ = log.Close()
	return nil
}

// handleWtDeleteAction handles the wt-delete subcommand action.
func handleWtDeleteAction(c *urfavecli.Context) error {
	ctx := context.Background()

	// Load config with global flags
	cfg, err := loadCLIConfig(
		c.String("config-file"),
		c.String("worktree-dir"),
		c.StringSlice("config"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}

	gitSvc := newCLIGitService(cfg)

	// Get worktree path from args
	worktreePath := ""
	if c.NArg() > 0 {
		worktreePath = c.Args().Get(0)
	}

	// Extract command-specific flags
	noBranch := c.Bool("no-branch")
	silent := c.Bool("silent")

	// Execute delete operation
	deleteBranch := !noBranch
	if err := cli.DeleteWorktree(ctx, gitSvc, cfg, worktreePath, deleteBranch, silent); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		_ = log.Close()
		return err
	}

	_ = log.Close()
	return nil
}
