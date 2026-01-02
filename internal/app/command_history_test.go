package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/chmouel/lazyworktree/internal/config"
	"github.com/chmouel/lazyworktree/internal/models"
)

const (
	testRepoKey     = "test-repo"
	testMakeCommand = "make test"
)

func TestAddToCommandHistory(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.repoKey = testRepoKey

	// Add first command
	m.addToCommandHistory(testMakeCommand)
	if len(m.commandHistory) != 1 {
		t.Fatalf("expected 1 command in history, got %d", len(m.commandHistory))
	}
	if m.commandHistory[0] != testMakeCommand {
		t.Fatalf("expected %q, got %q", testMakeCommand, m.commandHistory[0])
	}

	// Add second command
	m.addToCommandHistory("npm install")
	if len(m.commandHistory) != 2 {
		t.Fatalf("expected 2 commands in history, got %d", len(m.commandHistory))
	}
	if m.commandHistory[0] != "npm install" {
		t.Fatalf("expected 'npm install' at front, got %q", m.commandHistory[0])
	}

	// Add duplicate - should move to front
	m.addToCommandHistory(testMakeCommand)
	if len(m.commandHistory) != 2 {
		t.Fatalf("expected 2 commands (no duplicate), got %d", len(m.commandHistory))
	}
	if m.commandHistory[0] != testMakeCommand {
		t.Fatalf("expected %q at front after duplicate, got %q", testMakeCommand, m.commandHistory[0])
	}

	// Add empty string - should be ignored
	m.addToCommandHistory("")
	if len(m.commandHistory) != 2 {
		t.Fatalf("expected empty command to be ignored, got %d commands", len(m.commandHistory))
	}
}

func TestLoadCommandHistoryEmptyFile(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.repoKey = testRepoKey

	// Load when no history file exists
	m.loadCommandHistory()
	if len(m.commandHistory) != 0 {
		t.Fatalf("expected empty history, got %d commands", len(m.commandHistory))
	}
}

func TestSaveAndLoadCommandHistory(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.repoKey = testRepoKey

	// Add commands
	m.addToCommandHistory("make build")
	m.addToCommandHistory("go test")
	m.addToCommandHistory("make lint")

	// Create a new model and load history
	m2 := NewModel(cfg, "")
	m2.repoKey = testRepoKey
	m2.loadCommandHistory()

	if len(m2.commandHistory) != 3 {
		t.Fatalf("expected 3 commands after load, got %d", len(m2.commandHistory))
	}
	if m2.commandHistory[0] != "make lint" {
		t.Fatalf("expected 'make lint' at front, got %q", m2.commandHistory[0])
	}
	if m2.commandHistory[1] != "go test" {
		t.Fatalf("expected 'go test' at position 1, got %q", m2.commandHistory[1])
	}
	if m2.commandHistory[2] != "make build" {
		t.Fatalf("expected 'make build' at position 2, got %q", m2.commandHistory[2])
	}
}

func TestCommandHistoryLimit(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.repoKey = testRepoKey

	// Add 150 unique commands (more than the 100 limit)
	for i := 0; i < 150; i++ {
		m.addToCommandHistory("command-" + string(rune(i)))
	}

	if len(m.commandHistory) != 100 {
		t.Fatalf("expected history to be limited to 100, got %d", len(m.commandHistory))
	}
}

func TestLoadCommandHistoryInvalidJSON(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.repoKey = testRepoKey

	// Create invalid JSON file
	historyPath := filepath.Join(cfg.WorktreeDir, testRepoKey, models.CommandHistoryFilename)
	if err := os.MkdirAll(filepath.Dir(historyPath), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(historyPath, []byte("invalid json"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Should not crash and should return empty history
	m.loadCommandHistory()
	if len(m.commandHistory) != 0 {
		t.Fatalf("expected empty history on invalid JSON, got %d commands", len(m.commandHistory))
	}
}

func TestCommandHistoryFilePath(t *testing.T) {
	cfg := &config.AppConfig{
		WorktreeDir: t.TempDir(),
	}
	m := NewModel(cfg, "")
	m.repoKey = testRepoKey

	m.addToCommandHistory("test command")

	// Verify file was created at correct path
	expectedPath := filepath.Join(cfg.WorktreeDir, testRepoKey, models.CommandHistoryFilename)
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("expected history file to exist at %s", expectedPath)
	}

	// Verify file contains correct JSON
	// #nosec G304 -- expectedPath is constructed from vetted test directory
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatal(err)
	}

	var payload struct {
		Commands []string `json:"commands"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("failed to parse history file: %v", err)
	}

	if len(payload.Commands) != 1 {
		t.Fatalf("expected 1 command in file, got %d", len(payload.Commands))
	}
	if payload.Commands[0] != "test command" {
		t.Fatalf("expected 'test command', got %q", payload.Commands[0])
	}
}
