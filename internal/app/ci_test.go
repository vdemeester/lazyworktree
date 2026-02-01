package app

import (
	"path/filepath"
	"testing"

	"github.com/chmouel/lazyworktree/internal/config"
	"github.com/chmouel/lazyworktree/internal/models"
)

func TestShouldRefreshCI_NoWorktreeSelected(t *testing.T) {
	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")

	// No worktrees, selectedIndex is -1
	m.state.data.selectedIndex = -1

	if m.shouldRefreshCI() {
		t.Error("expected false when no worktree is selected")
	}
}

func TestShouldRefreshCI_IndexOutOfBounds(t *testing.T) {
	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")

	// Index beyond available worktrees
	m.state.data.selectedIndex = 5
	m.state.data.filteredWts = []*models.WorktreeInfo{}

	if m.shouldRefreshCI() {
		t.Error("expected false when selectedIndex is out of bounds")
	}
}

func TestShouldRefreshCI_OpenPR(t *testing.T) {
	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")

	wtPath := filepath.Join(cfg.WorktreeDir, "wt1")
	m.state.data.filteredWts = []*models.WorktreeInfo{
		{
			Path:   wtPath,
			Branch: "feature-branch",
			PR: &models.PRInfo{
				Number: 123,
				State:  "OPEN",
			},
		},
	}
	m.state.data.selectedIndex = 0

	// Should always refresh for open PRs
	if !m.shouldRefreshCI() {
		t.Error("expected true for open PR (new jobs can start anytime)")
	}
}

func TestShouldRefreshCI_ClosedPR_AllJobsComplete(t *testing.T) {
	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")

	wtPath := filepath.Join(cfg.WorktreeDir, "wt1")
	branch := "feature-branch"
	m.state.data.filteredWts = []*models.WorktreeInfo{
		{
			Path:   wtPath,
			Branch: branch,
			PR: &models.PRInfo{
				Number: 123,
				State:  "CLOSED",
			},
		},
	}
	m.state.data.selectedIndex = 0

	// Add completed CI checks to cache
	m.cache.ciCache.Set(branch, []*models.CICheck{
		{Name: "build", Conclusion: "success"},
		{Name: "test", Conclusion: "failure"},
	})

	// Should not refresh for closed PR with all jobs complete
	if m.shouldRefreshCI() {
		t.Error("expected false for closed PR with all jobs complete")
	}
}

func TestShouldRefreshCI_ClosedPR_PendingJobs(t *testing.T) {
	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")

	wtPath := filepath.Join(cfg.WorktreeDir, "wt1")
	branch := "feature-branch"
	m.state.data.filteredWts = []*models.WorktreeInfo{
		{
			Path:   wtPath,
			Branch: branch,
			PR: &models.PRInfo{
				Number: 123,
				State:  "CLOSED",
			},
		},
	}
	m.state.data.selectedIndex = 0

	// Add CI checks with pending job to cache
	m.cache.ciCache.Set(branch, []*models.CICheck{
		{Name: "build", Conclusion: "success"},
		{Name: "test", Conclusion: "pending"},
	})

	// Should refresh when jobs are pending
	if !m.shouldRefreshCI() {
		t.Error("expected true for closed PR with pending jobs")
	}
}

func TestShouldRefreshCI_NoPR_NoCache(t *testing.T) {
	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")

	wtPath := filepath.Join(cfg.WorktreeDir, "wt1")
	m.state.data.filteredWts = []*models.WorktreeInfo{
		{
			Path:   wtPath,
			Branch: "feature-branch",
			PR:     nil, // No PR
		},
	}
	m.state.data.selectedIndex = 0

	// No cache entry, should fetch
	if !m.shouldRefreshCI() {
		t.Error("expected true when no cache exists (should fetch)")
	}
}

func TestShouldRefreshCI_NoPR_EmptyConclusion(t *testing.T) {
	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")

	wtPath := filepath.Join(cfg.WorktreeDir, "wt1")
	branch := "feature-branch"
	m.state.data.filteredWts = []*models.WorktreeInfo{
		{
			Path:   wtPath,
			Branch: branch,
			PR:     nil,
		},
	}
	m.state.data.selectedIndex = 0

	// Add CI check with empty conclusion (in-progress job)
	m.cache.ciCache.Set(branch, []*models.CICheck{
		{Name: "build", Conclusion: "success"},
		{Name: "test", Conclusion: ""},
	})

	// Should refresh when conclusion is empty (job in progress)
	if !m.shouldRefreshCI() {
		t.Error("expected true when a job has empty conclusion (in-progress)")
	}
}

func TestShouldRefreshCI_NoPR_AllComplete(t *testing.T) {
	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")

	wtPath := filepath.Join(cfg.WorktreeDir, "wt1")
	branch := "feature-branch"
	m.state.data.filteredWts = []*models.WorktreeInfo{
		{
			Path:   wtPath,
			Branch: branch,
			PR:     nil,
		},
	}
	m.state.data.selectedIndex = 0

	// Add all completed CI checks to cache
	m.cache.ciCache.Set(branch, []*models.CICheck{
		{Name: "build", Conclusion: "success"},
		{Name: "test", Conclusion: "skipped"},
		{Name: "deploy", Conclusion: "cancelled"},
	})

	// Should not refresh when all jobs are complete
	if m.shouldRefreshCI() {
		t.Error("expected false when all jobs are complete (success/skipped/cancelled)")
	}
}
