package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/chmouel/lazyworktree/internal/config"
	"github.com/chmouel/lazyworktree/internal/models"
)

type repoInfo struct {
	dir    string
	branch string
	commit commitOption
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, output)
	}
	return strings.TrimSpace(string(output))
}

func initTestRepo(t *testing.T) repoInfo {
	t.Helper()

	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")

	filePath := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(filePath, []byte("one\n"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runGit(t, dir, "add", "file.txt")
	runGit(t, dir, "commit", "-m", "Initial commit")

	if err := os.WriteFile(filePath, []byte("two\n"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runGit(t, dir, "commit", "-am", "Add new feature")

	branch := runGit(t, dir, "branch", "--show-current")
	if branch == "" {
		t.Fatal("expected branch name")
	}

	runGit(t, dir, "branch", featureBranch)
	runGit(t, dir, "update-ref", "refs/remotes/origin/"+branch, "HEAD")

	log := runGit(t, dir, "log", "-1", "--pretty=format:%H%x1f%h%x1f%s")
	parts := strings.SplitN(log, "\x1f", 3)
	if len(parts) < 3 {
		t.Fatalf("unexpected log output: %q", log)
	}

	return repoInfo{
		dir:    dir,
		branch: branch,
		commit: commitOption{fullHash: parts[0], shortHash: parts[1], subject: parts[2]},
	}
}

func withCwd(t *testing.T, dir string) {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
}

func TestShowBaseSelection(t *testing.T) {
	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")
	m.windowWidth = 120
	m.windowHeight = 40

	cmd := m.showBaseSelection(mainWorktreeName)
	if cmd == nil {
		t.Fatal("expected command to be returned")
	}
	if m.listScreen == nil || m.currentScreen != screenListSelect {
		t.Fatal("expected list screen to be active")
	}

	m.listSubmit(selectionItem{id: "freeform"})
	if m.inputScreen == nil || m.currentScreen != screenInput {
		t.Fatal("expected input screen to be active")
	}
	if m.inputScreen.prompt != "Base ref" {
		t.Fatalf("expected base ref prompt, got %q", m.inputScreen.prompt)
	}
}

func TestShowFreeformBaseInputValidation(t *testing.T) {
	repo := initTestRepo(t)
	withCwd(t, repo.dir)

	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")

	m.showFreeformBaseInput(repo.branch)
	if _, ok := m.inputSubmit(" "); ok {
		t.Fatal("expected empty base ref to be rejected")
	}
	if m.inputScreen.errorMsg != "Base ref cannot be empty." {
		t.Fatalf("unexpected error: %q", m.inputScreen.errorMsg)
	}

	if _, ok := m.inputSubmit("missing-ref"); ok {
		t.Fatal("expected invalid base ref to be rejected")
	}
	if m.inputScreen.errorMsg != "Base ref not found." {
		t.Fatalf("unexpected error: %q", m.inputScreen.errorMsg)
	}

	if _, ok := m.inputSubmit(repo.branch); ok {
		t.Fatal("expected base ref flow to keep screen open")
	}
	if m.inputScreen == nil || m.inputScreen.prompt != "Create worktree: branch name" {
		t.Fatal("expected branch name input to be shown")
	}
}

func TestShowBranchSelection(t *testing.T) {
	repo := initTestRepo(t)
	withCwd(t, repo.dir)

	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")
	m.windowWidth = 120
	m.windowHeight = 40

	selected := ""
	cmd := m.showBranchSelection("Pick", "Filter...", "None", repo.branch, func(branch string) tea.Cmd {
		selected = branch
		return nil
	})

	if cmd == nil {
		t.Fatal("expected command to be returned")
	}
	if m.listScreen == nil || len(m.listScreen.items) == 0 {
		t.Fatal("expected branch list to be populated")
	}
	if m.listScreen.items[0].id != repo.branch {
		t.Fatalf("expected preferred branch first, got %q", m.listScreen.items[0].id)
	}

	remoteFound := false
	for _, item := range m.listScreen.items {
		if item.description == "remote" {
			remoteFound = true
			break
		}
	}
	if !remoteFound {
		t.Fatal("expected a remote branch entry")
	}

	m.listSubmit(m.listScreen.items[0])
	if selected != repo.branch {
		t.Fatalf("expected %q to be selected, got %q", repo.branch, selected)
	}
}

func TestShowCommitSelection(t *testing.T) {
	repo := initTestRepo(t)
	withCwd(t, repo.dir)

	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")
	m.windowWidth = 120
	m.windowHeight = 40

	cmd := m.showCommitSelection(repo.branch)
	if cmd == nil {
		t.Fatal("expected command to be returned")
	}
	if m.listScreen == nil || len(m.listScreen.items) == 0 {
		t.Fatal("expected commit list to be populated")
	}

	item := m.listScreen.items[0]
	if item.description == "" {
		t.Fatal("expected commit item to include date")
	}

	m.listSubmit(item)
	if m.inputScreen == nil || m.inputScreen.prompt != "Create worktree: branch name" {
		t.Fatal("expected branch name input to be shown")
	}

	expected := sanitizeBranchNameFromTitle(repo.commit.subject, repo.commit.shortHash)
	if got := m.inputScreen.input.Value(); got != expected {
		t.Fatalf("expected branch name %q, got %q", expected, got)
	}
}

func TestShowCommitSelectionShowsInfoOnBranchNameScriptError(t *testing.T) {
	repo := initTestRepo(t)
	withCwd(t, repo.dir)

	cfg := &config.AppConfig{
		WorktreeDir:      t.TempDir(),
		BranchNameScript: "false",
	}
	m := NewModel(cfg, "")
	m.windowWidth = 120
	m.windowHeight = 40

	cmd := m.showCommitSelection(repo.branch)
	if cmd == nil {
		t.Fatal("expected command to be returned")
	}
	if m.listScreen == nil || len(m.listScreen.items) == 0 {
		t.Fatal("expected commit list to be populated")
	}

	item := m.listScreen.items[0]
	if cmd := m.listSubmit(item); cmd != nil {
		t.Fatal("expected nil command on script error")
	}
	if m.currentScreen != screenInfo {
		t.Fatalf("expected info screen, got %v", m.currentScreen)
	}
	if m.infoScreen == nil || !strings.Contains(m.infoScreen.message, "Branch name script error") {
		t.Fatalf("expected branch name script error modal, got %#v", m.infoScreen)
	}

	_, action := m.handleScreenKey(tea.KeyMsg{Type: tea.KeyEnter})
	if action != nil {
		_ = action()
	}

	if m.currentScreen != screenInput {
		t.Fatalf("expected input screen, got %v", m.currentScreen)
	}
	if m.inputScreen == nil {
		t.Fatal("expected input screen to be set")
	}
}

func TestShowBranchNameInputValidation(t *testing.T) {
	repo := initTestRepo(t)
	withCwd(t, repo.dir)

	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")
	m.worktrees = []*models.WorktreeInfo{{Branch: "demo"}}

	m.showBranchNameInput(repo.branch, "demo")
	if got := m.inputScreen.input.Value(); got != "demo-1" {
		t.Fatalf("expected suggested branch name, got %q", got)
	}

	if _, ok := m.inputSubmit("demo"); ok {
		t.Fatal("expected duplicate branch to be rejected")
	}
	if !strings.Contains(m.inputScreen.errorMsg, "already exists") {
		t.Fatalf("unexpected error: %q", m.inputScreen.errorMsg)
	}

	pathBranch := "path-branch"
	if err := os.MkdirAll(filepath.Join(m.getWorktreeDir(), pathBranch), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if _, ok := m.inputSubmit(pathBranch); ok {
		t.Fatal("expected existing path to be rejected")
	}
	if !strings.Contains(m.inputScreen.errorMsg, "Path already exists") {
		t.Fatalf("unexpected error: %q", m.inputScreen.errorMsg)
	}
}

func TestSuggestBranchName(t *testing.T) {
	repo := initTestRepo(t)
	withCwd(t, repo.dir)

	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")
	m.worktrees = []*models.WorktreeInfo{{Branch: "demo"}}

	if got := m.suggestBranchName("demo"); got != "demo-1" {
		t.Fatalf("expected demo-1, got %q", got)
	}
}

func TestCommitLog(t *testing.T) {
	repo := initTestRepo(t)
	withCwd(t, repo.dir)

	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")

	log := m.commitLog(repo.commit.fullHash)
	if !strings.Contains(log, repo.commit.subject) {
		t.Fatalf("expected commit log to include subject, got %q", log)
	}
}

func TestCreateWorktreeFromBase(t *testing.T) {
	repo := initTestRepo(t)
	withCwd(t, repo.dir)

	worktreeDir := t.TempDir()
	cfg := &config.AppConfig{WorktreeDir: worktreeDir}
	m := NewModel(cfg, "")

	newBranch := "new-worktree"
	targetPath := filepath.Join(worktreeDir, newBranch)

	cmd := m.createWorktreeFromBase(newBranch, targetPath, repo.branch)
	if cmd == nil {
		t.Fatal("expected command to be returned")
	}
	msg := cmd()
	loaded, ok := msg.(worktreesLoadedMsg)
	if !ok {
		t.Fatalf("expected worktreesLoadedMsg, got %T", msg)
	}
	if loaded.err != nil {
		t.Fatalf("unexpected load error: %v", loaded.err)
	}

	if _, err := os.Stat(targetPath); err != nil {
		t.Fatalf("expected worktree path to exist: %v", err)
	}

	found := false
	for _, wt := range loaded.worktrees {
		if wt.Branch == newBranch {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected worktree for branch %s", newBranch)
	}
}

func TestClearListSelection(t *testing.T) {
	cfg := &config.AppConfig{WorktreeDir: t.TempDir()}
	m := NewModel(cfg, "")
	m.listScreen = &ListSelectionScreen{}
	m.listSubmit = func(selectionItem) tea.Cmd { return nil }
	m.currentScreen = screenListSelect

	m.clearListSelection()
	if m.listScreen != nil || m.listSubmit != nil {
		t.Fatal("expected list selection to be cleared")
	}
	if m.currentScreen != screenNone {
		t.Fatalf("expected screenNone, got %v", m.currentScreen)
	}
}

func TestBuildCommitItems(t *testing.T) {
	items := buildCommitItems([]commitOption{{
		fullHash:  "full",
		shortHash: "short",
		date:      "2024-01-01",
		subject:   "subject",
	}})

	if len(items) != 1 {
		t.Fatalf("expected one item, got %d", len(items))
	}
	if items[0].label != "short subject" {
		t.Fatalf("unexpected label: %q", items[0].label)
	}
	if items[0].description != "2024-01-01" {
		t.Fatalf("unexpected description: %q", items[0].description)
	}
}
