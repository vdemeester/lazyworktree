package app

import (
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/chmouel/lazyworktree/internal/config"
)

func TestFilterPaletteItemsEmptyQueryReturnsAll(t *testing.T) {
	items := []paletteItem{
		{id: "create", label: "Create worktree", description: "Add a new worktree"},
		{id: "delete", label: "Delete worktree", description: "Remove worktree"},
		{id: "help", label: "Help", description: "Show help"},
	}

	got := filterPaletteItems(items, "")
	if !reflect.DeepEqual(got, items) {
		t.Fatalf("expected items to be unchanged, got %#v", got)
	}
}

func TestFilterPaletteItemsPrefersLabelMatches(t *testing.T) {
	items := []paletteItem{
		{id: "desc", label: "Delete worktree", description: "Create new worktree"},
		{id: "label", label: "Create worktree", description: "Remove worktree"},
		{id: "help", label: "Help", description: "Show help"},
	}

	got := filterPaletteItems(items, "create")
	if len(got) < 2 {
		t.Fatalf("expected at least two matches, got %d", len(got))
	}
	if got[0].id != "label" {
		t.Fatalf("expected label match first, got %q", got[0].id)
	}
	if got[1].id != "desc" {
		t.Fatalf("expected description match second, got %q", got[1].id)
	}
}

func TestFuzzyScoreLowerMissingChars(t *testing.T) {
	if _, ok := fuzzyScoreLower("zz", "create worktree"); ok {
		t.Fatalf("expected fuzzy match to fail")
	}
}

func TestHandleMouseDoesNotPanic(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: "/tmp/test",
	}
	m := NewModel(cfg, "")
	m.windowWidth = 120
	m.windowHeight = 40

	// Test mouse wheel events
	mouseMsg := tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonWheelUp,
		X:      10,
		Y:      5,
	}

	result, _ := m.handleMouse(mouseMsg)
	if result == nil {
		t.Fatal("handleMouse returned nil model")
	}

	// Test mouse click
	mouseMsg = tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      10,
		Y:      5,
	}

	result, _ = m.handleMouse(mouseMsg)
	if result == nil {
		t.Fatal("handleMouse returned nil model")
	}
}
