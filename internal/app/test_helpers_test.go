package app

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func hasGitRepo(t *testing.T) bool {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		return false
	}
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return false
	}
	return strings.TrimSpace(out.String()) == "true"
}

func requireGitRepo(t *testing.T) {
	t.Helper()
	if !hasGitRepo(t) {
		t.Skip("git not available or not in a git worktree")
	}
}
