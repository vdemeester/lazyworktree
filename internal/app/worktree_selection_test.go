package app

import (
	"testing"

	"github.com/charmbracelet/bubbles/table"
	"github.com/chmouel/lazyworktree/internal/config"
	"github.com/chmouel/lazyworktree/internal/models"
)

func TestDetermineCurrentWorktreePrefersSelection(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")

	main := &models.WorktreeInfo{Path: "/tmp/main", Branch: "main", IsMain: true}
	feature := &models.WorktreeInfo{Path: "/tmp/feature", Branch: "feature"}
	m.data.worktrees = []*models.WorktreeInfo{main, feature}
	m.data.filteredWts = m.data.worktrees

	rows := []table.Row{
		{"main"},
		{"feature"},
	}
	m.ui.worktreeTable.SetRows(rows)
	m.ui.worktreeTable.SetCursor(1)

	got := m.determineCurrentWorktree()
	if got != feature {
		t.Fatalf("expected selected worktree, got %v", got)
	}
}
