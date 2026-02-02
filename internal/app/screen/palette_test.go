package screen

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/chmouel/lazyworktree/internal/theme"
)

func TestCommandPaletteFilterToggle(t *testing.T) {
	items := []PaletteItem{
		{ID: "alpha", Label: "Alpha"},
		{ID: "beta", Label: "Beta"},
	}

	scr := NewCommandPaletteScreen(items, 80, 24, theme.Dracula())
	if scr.FilterActive {
		t.Fatal("expected filter to be inactive by default")
	}

	next, _ := scr.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	nextScr, ok := next.(*CommandPaletteScreen)
	if !ok || nextScr == nil {
		t.Fatal("expected Update to return command palette screen after f")
	}
	scr = nextScr
	if !scr.FilterActive {
		t.Fatal("expected filter to be active after f")
	}

	next, _ = scr.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	nextScr, ok = next.(*CommandPaletteScreen)
	if !ok || nextScr == nil {
		t.Fatal("expected Update to return command palette screen after typing")
	}
	scr = nextScr
	if len(scr.Filtered) != 1 || scr.Filtered[0].ID != "beta" {
		t.Fatalf("expected filtered results to include only 'beta', got %v", scr.Filtered)
	}

	next, _ = scr.Update(tea.KeyMsg{Type: tea.KeyEsc})
	nextScr, ok = next.(*CommandPaletteScreen)
	if !ok || nextScr == nil {
		t.Fatal("expected Update to return command palette screen after Esc")
	}
	scr = nextScr
	if scr.FilterActive {
		t.Fatal("expected filter to be inactive after Esc")
	}
	if len(scr.Filtered) != 1 || scr.Filtered[0].ID != "beta" {
		t.Fatalf("expected filter to remain applied after Esc, got %v", scr.Filtered)
	}
}
