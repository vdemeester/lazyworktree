package app

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const commitListLimit = 25

type branchOption struct {
	name     string
	isRemote bool
}

type commitOption struct {
	fullHash  string
	shortHash string
	date      string
	subject   string
}

func (m *Model) showBaseSelection(defaultBase string) tea.Cmd {
	items := []selectionItem{
		{id: "branch-list", label: "Pick a base branch", description: "Local and remote branches"},
		{id: "commit-list", label: "Pick a base commit", description: "Choose a branch, then a commit"},
		{id: "freeform", label: "Enter base ref manually", description: "Type a branch or commit"},
	}
	title := "Select base for new worktree"

	m.listScreen = NewListSelectionScreen(items, title, "Filter options...", "No base options available.", m.windowWidth, m.windowHeight, "", m.theme)
	m.listSubmit = func(item selectionItem) tea.Cmd {
		switch item.id {
		case "branch-list":
			return m.showBranchSelection(
				"Select base branch",
				"Filter branches...",
				"No branches found.",
				defaultBase,
				func(branch string) tea.Cmd {
					return m.showBranchNameInput(branch, branch)
				},
			)
		case "commit-list":
			return m.showCommitSelection(defaultBase)
		case "freeform":
			return m.showFreeformBaseInput(defaultBase)
		default:
			return nil
		}
	}

	m.currentScreen = screenListSelect
	return textinput.Blink
}

func (m *Model) showFreeformBaseInput(defaultBase string) tea.Cmd {
	m.clearListSelection()
	m.inputScreen = NewInputScreen("Base ref", defaultBase, defaultBase, m.theme)
	m.inputSubmit = func(baseVal string) (tea.Cmd, bool) {
		baseRef := strings.TrimSpace(baseVal)
		if baseRef == "" {
			m.inputScreen.errorMsg = "Base ref cannot be empty."
			return nil, false
		}
		if !m.baseRefExists(baseRef) {
			m.inputScreen.errorMsg = "Base ref not found."
			return nil, false
		}
		m.inputScreen.errorMsg = ""
		return m.showBranchNameInput(baseRef, ""), false
	}
	m.currentScreen = screenInput
	return textinput.Blink
}

func (m *Model) showBranchSelection(title, placeholder, noResults, preferred string, onSelect func(string) tea.Cmd) tea.Cmd {
	items := m.branchSelectionItems(preferred)
	m.listScreen = NewListSelectionScreen(items, title, placeholder, noResults, m.windowWidth, m.windowHeight, preferred, m.theme)
	m.listSubmit = func(item selectionItem) tea.Cmd {
		return onSelect(item.id)
	}
	m.currentScreen = screenListSelect
	return textinput.Blink
}

func (m *Model) showCommitSelection(baseBranch string) tea.Cmd {
	raw := m.git.RunGit(
		m.ctx,
		[]string{
			"git", "log",
			fmt.Sprintf("--max-count=%d", commitListLimit),
			"--date=short",
			"--pretty=format:%H%x1f%h%x1f%ad%x1f%s",
			baseBranch,
		},
		"",
		[]int{0},
		true,
		false,
	)
	commits := parseCommitOptions(raw)
	items := buildCommitItems(commits)
	commitLookup := make(map[string]commitOption, len(commits))
	for _, commit := range commits {
		commitLookup[commit.fullHash] = commit
	}
	title := fmt.Sprintf("Select commit from %q", baseBranch)
	noResults := fmt.Sprintf("No commits found on %s.", baseBranch)

	m.listScreen = NewListSelectionScreen(items, title, "Filter commits...", noResults, m.windowWidth, m.windowHeight, "", m.theme)
	m.listSubmit = func(item selectionItem) tea.Cmd {
		m.clearListSelection()
		commit, ok := commitLookup[item.id]
		if !ok {
			commit = commitOption{fullHash: item.id}
		}

		var commitMessage string
		if strings.TrimSpace(m.config.BranchNameScript) != "" {
			commitMessage = m.commitLog(item.id)
		}

		defaultName := ""
		scriptErr := ""
		if strings.TrimSpace(m.config.BranchNameScript) != "" && commitMessage != "" {
			if generatedName, err := runBranchNameScript(m.ctx, m.config.BranchNameScript, commitMessage); err != nil {
				scriptErr = fmt.Sprintf("Branch name script error: %v", err)
			} else if generatedName != "" {
				defaultName = generatedName
			}
		}

		if defaultName == "" {
			subject := strings.TrimSpace(commit.subject)
			if subject == "" && commitMessage != "" {
				subject = strings.TrimSpace(strings.Split(commitMessage, "\n")[0])
			}
			defaultName = sanitizeBranchNameFromTitle(subject, commit.shortHash)
		}

		if scriptErr != "" {
			m.showInfo(scriptErr, func() tea.Msg {
				cmd := m.showBranchNameInput(item.id, defaultName)
				if cmd != nil {
					return cmd()
				}
				return nil
			})
			return nil
		}
		return m.showBranchNameInput(item.id, defaultName)
	}
	m.currentScreen = screenListSelect
	return textinput.Blink
}

func (m *Model) showBranchNameInput(baseRef, defaultName string) tea.Cmd {
	m.clearListSelection()
	suggested := strings.TrimSpace(defaultName)
	if suggested != "" {
		suggested = m.suggestBranchName(suggested)
	}
	m.inputScreen = NewInputScreen("Create worktree: branch name", "feature/my-branch", suggested, m.theme)
	m.inputSubmit = func(value string) (tea.Cmd, bool) {
		newBranch := strings.TrimSpace(value)
		if newBranch == "" {
			m.inputScreen.errorMsg = errBranchEmpty
			return nil, false
		}

		// Prevent duplicates
		for _, wt := range m.worktrees {
			if wt.Branch == newBranch {
				m.inputScreen.errorMsg = fmt.Sprintf("Branch %q already exists.", newBranch)
				return nil, false
			}
		}

		targetPath := filepath.Join(m.getWorktreeDir(), newBranch)
		if _, err := os.Stat(targetPath); err == nil {
			m.inputScreen.errorMsg = fmt.Sprintf("Path already exists: %s", targetPath)
			return nil, false
		}

		return m.createWorktreeFromBase(newBranch, targetPath, baseRef), true
	}
	m.currentScreen = screenInput
	return textinput.Blink
}

func (m *Model) suggestBranchName(baseName string) string {
	existing := make(map[string]struct{})
	for _, wt := range m.worktrees {
		if wt.Branch == "" || wt.Branch == "(detached)" {
			continue
		}
		existing[wt.Branch] = struct{}{}
	}

	raw := m.git.RunGit(
		m.ctx,
		[]string{
			"git", "for-each-ref",
			"--format=%(refname:short)",
			"refs/heads",
		},
		"",
		[]int{0},
		true,
		false,
	)
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		branch := strings.TrimSpace(line)
		if branch == "" {
			continue
		}
		existing[branch] = struct{}{}
	}

	return suggestBranchNameWithExisting(baseName, existing)
}

func suggestBranchNameWithExisting(baseName string, existing map[string]struct{}) string {
	baseName = strings.TrimSpace(baseName)
	if baseName == "" {
		return ""
	}
	if _, ok := existing[baseName]; !ok {
		return baseName
	}
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s-%d", baseName, i)
		if _, ok := existing[candidate]; !ok {
			return candidate
		}
	}
}

func (m *Model) commitLog(hash string) string {
	return m.git.RunGit(
		m.ctx,
		[]string{"git", "show", "--quiet", "--pretty=format:%s%n%b", hash},
		"",
		[]int{0},
		true,
		false,
	)
}

func (m *Model) baseRefExists(ref string) bool {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return false
	}
	refQuery := fmt.Sprintf("%s^{commit}", ref)
	out := m.git.RunGit(
		m.ctx,
		[]string{"git", "rev-parse", "--verify", refQuery},
		"",
		[]int{0, 1},
		true,
		true,
	)
	return strings.TrimSpace(out) != ""
}

func sanitizeBranchNameFromTitle(title, fallback string) string {
	sanitized := strings.ToLower(strings.TrimSpace(title))
	sanitized = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(sanitized, "-")
	sanitized = strings.Trim(sanitized, "-")
	sanitized = regexp.MustCompile(`-+`).ReplaceAllString(sanitized, "-")
	if len(sanitized) > 50 {
		sanitized = strings.TrimRight(sanitized[:50], "-")
	}
	if sanitized == "" {
		sanitized = strings.TrimSpace(fallback)
	}
	if sanitized == "" {
		sanitized = "commit"
	}
	return sanitized
}

func (m *Model) createWorktreeFromBase(newBranch, targetPath, baseRef string) tea.Cmd {
	if err := os.MkdirAll(m.getWorktreeDir(), defaultDirPerms); err != nil {
		return func() tea.Msg { return errMsg{err: fmt.Errorf("failed to create worktree directory: %w", err)} }
	}

	ok := m.git.RunCommandChecked(
		m.ctx,
		[]string{"git", "worktree", "add", "-b", newBranch, targetPath, baseRef},
		"",
		fmt.Sprintf("Failed to create worktree %s", newBranch),
	)
	if !ok {
		return func() tea.Msg { return errMsg{err: fmt.Errorf("failed to create worktree %s", newBranch)} }
	}

	env := m.buildCommandEnv(newBranch, targetPath)
	initCmds := m.collectInitCommands()
	after := func() tea.Msg {
		worktrees, err := m.git.GetWorktrees(m.ctx)
		return worktreesLoadedMsg{
			worktrees: worktrees,
			err:       err,
		}
	}
	return m.runCommandsWithTrust(initCmds, targetPath, env, after)
}

func (m *Model) clearListSelection() {
	m.listScreen = nil
	m.listSubmit = nil
	if m.currentScreen == screenListSelect {
		m.currentScreen = screenNone
	}
}

func (m *Model) branchSelectionItems(preferred string) []selectionItem {
	raw := m.git.RunGit(
		m.ctx,
		[]string{
			"git", "for-each-ref",
			"--sort=refname",
			"--format=%(refname:short)\t%(refname)",
			"refs/heads",
			"refs/remotes",
		},
		"",
		[]int{0},
		true,
		false,
	)

	options := parseBranchOptions(raw)
	options = prioritizeBranchOptions(options, preferred)
	items := make([]selectionItem, 0, len(options))
	for _, opt := range options {
		desc := ""
		if opt.isRemote {
			desc = "remote"
		}
		items = append(items, selectionItem{
			id:          opt.name,
			label:       opt.name,
			description: desc,
		})
	}
	return items
}

func parseBranchOptions(raw string) []branchOption {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) == 1 && strings.TrimSpace(lines[0]) == "" {
		return nil
	}

	options := make([]branchOption, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		fullRef := strings.TrimSpace(parts[1])
		if name == "" || fullRef == "" {
			continue
		}
		if strings.HasSuffix(name, "/HEAD") {
			continue
		}
		options = append(options, branchOption{
			name:     name,
			isRemote: strings.HasPrefix(fullRef, "refs/remotes/"),
		})
	}
	return options
}

func prioritizeBranchOptions(options []branchOption, preferred string) []branchOption {
	if preferred == "" || len(options) == 0 {
		return options
	}
	idx := -1
	for i, opt := range options {
		if opt.name == preferred {
			idx = i
			break
		}
	}
	if idx <= 0 {
		return options
	}
	ordered := make([]branchOption, 0, len(options))
	ordered = append(ordered, options[idx])
	ordered = append(ordered, options[:idx]...)
	ordered = append(ordered, options[idx+1:]...)
	return ordered
}

func parseCommitOptions(raw string) []commitOption {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) == 1 && strings.TrimSpace(lines[0]) == "" {
		return nil
	}

	options := make([]commitOption, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\x1f", 4)
		if len(parts) < 4 {
			continue
		}
		full := strings.TrimSpace(parts[0])
		short := strings.TrimSpace(parts[1])
		date := strings.TrimSpace(parts[2])
		subject := strings.TrimSpace(parts[3])
		if full == "" {
			continue
		}
		options = append(options, commitOption{
			fullHash:  full,
			shortHash: short,
			date:      date,
			subject:   subject,
		})
	}
	return options
}

func buildCommitItems(options []commitOption) []selectionItem {
	items := make([]selectionItem, 0, len(options))
	for _, opt := range options {
		label := strings.TrimSpace(fmt.Sprintf("%s %s", opt.shortHash, opt.subject))
		items = append(items, selectionItem{
			id:          opt.fullHash,
			label:       label,
			description: opt.date,
		})
	}
	return items
}
