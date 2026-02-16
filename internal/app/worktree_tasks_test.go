package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	appscreen "github.com/chmouel/lazyworktree/internal/app/screen"
	"github.com/chmouel/lazyworktree/internal/config"
	"github.com/chmouel/lazyworktree/internal/models"
)

func TestParseMarkdownTaskLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantOK   bool
		wantDone bool
		wantText string
	}{
		{name: "unchecked", line: "- [ ] Write docs", wantOK: true, wantDone: false, wantText: "Write docs"},
		{name: "checked upper", line: "* [X] Ship it", wantOK: true, wantDone: true, wantText: "Ship it"},
		{name: "checked lower", line: "+ [x] Merge PR", wantOK: true, wantDone: true, wantText: "Merge PR"},
		{name: "not task", line: "- TODO: plain text tag", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done, text, ok := parseMarkdownTaskLine(tt.line)
			if ok != tt.wantOK {
				t.Fatalf("ok=%v want=%v", ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if done != tt.wantDone {
				t.Fatalf("checked=%v want=%v", done, tt.wantDone)
			}
			if text != tt.wantText {
				t.Fatalf("text=%q want=%q", text, tt.wantText)
			}
		})
	}
}

func TestToggleMarkdownTaskLinePreservesFormatting(t *testing.T) {
	note := "## Notes\n  - [ ]   Keep spacing exactly\n- [x] done"
	updated, ok := toggleMarkdownTaskLine(note, 1)
	if !ok {
		t.Fatal("expected toggle to succeed")
	}
	lines := strings.Split(updated, "\n")
	if len(lines) != 3 {
		t.Fatalf("unexpected line count: %d", len(lines))
	}
	if lines[0] != "## Notes" {
		t.Fatalf("expected heading unchanged, got %q", lines[0])
	}
	if lines[1] != "  - [x]   Keep spacing exactly" {
		t.Fatalf("expected only checkbox marker to flip, got %q", lines[1])
	}
	if lines[2] != "- [x] done" {
		t.Fatalf("expected unrelated line unchanged, got %q", lines[2])
	}
}

func TestShowTaskboardNoTasksShowsInfo(t *testing.T) {
	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")
	m.state.data.worktrees = []*models.WorktreeInfo{
		{Path: "/tmp/wt-a", Branch: "feat-a"},
	}
	m.setWorktreeNote("/tmp/wt-a", "Just prose, no checkboxes.")

	cmd := m.showTaskboard()
	if cmd != nil {
		t.Fatal("expected nil command for no-task flow")
	}
	if !m.state.ui.screenManager.IsActive() || m.state.ui.screenManager.Type() != appscreen.TypeInfo {
		t.Fatalf("expected info screen, got active=%v type=%v", m.state.ui.screenManager.IsActive(), m.state.ui.screenManager.Type())
	}
}

func TestShowTaskboardAndToggleTask(t *testing.T) {
	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")
	wtPath := "/tmp/wt-a"
	m.state.data.worktrees = []*models.WorktreeInfo{
		{Path: wtPath, Branch: "feat-a"},
	}
	m.state.data.filteredWts = m.state.data.worktrees
	m.state.data.selectedIndex = 0
	m.state.view.FocusedPane = 0
	m.setWorktreeNote(wtPath, "- [ ] Write tests\n- [x] done already")

	cmd := m.showTaskboard()
	if cmd != nil {
		t.Fatalf("expected nil command, got %v", cmd)
	}
	if !m.state.ui.screenManager.IsActive() || m.state.ui.screenManager.Type() != appscreen.TypeTaskboard {
		t.Fatalf("expected taskboard screen, got active=%v type=%v", m.state.ui.screenManager.IsActive(), m.state.ui.screenManager.Type())
	}

	updated, _ := m.handleScreenKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = updated.(*Model)

	note, ok := m.getWorktreeNote(wtPath)
	if !ok {
		t.Fatal("expected note to exist")
	}
	if !strings.Contains(note.Note, "- [x] Write tests") {
		t.Fatalf("expected first task to be toggled, got note:\n%s", note.Note)
	}
}

func TestHandleBuiltInKeyTaskboardShortcut(t *testing.T) {
	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")
	wtPath := "/tmp/wt-a"
	m.state.data.worktrees = []*models.WorktreeInfo{{Path: wtPath, Branch: "feat-a"}}
	m.setWorktreeNote(wtPath, "- [ ] Open taskboard")

	_, _ = m.handleBuiltInKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'T'}})
	if !m.state.ui.screenManager.IsActive() || m.state.ui.screenManager.Type() != appscreen.TypeTaskboard {
		t.Fatalf("expected taskboard screen on T shortcut, got active=%v type=%v", m.state.ui.screenManager.IsActive(), m.state.ui.screenManager.Type())
	}
}
