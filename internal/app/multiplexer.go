package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/chmouel/lazyworktree/internal/config"
	"github.com/chmouel/lazyworktree/internal/models"
	"github.com/chmouel/lazyworktree/internal/multiplexer"
)

const (
	zellijSessionLabel = "zellij session"
	tmuxSessionLabel   = "tmux session"
)

type (
	zellijSessionReadyMsg struct {
		sessionName  string
		attach       bool
		insideZellij bool
	}
	tmuxSessionReadyMsg struct {
		sessionName string
		attach      bool
		insideTmux  bool
	}
)

func cleanupZellijLayouts(paths []string) {
	multiplexer.CleanupZellijLayouts(paths)
}

func buildZellijInfoMessage(sessionName string) string {
	quoted := multiplexer.ShellQuote(sessionName)
	return fmt.Sprintf("zellij session ready.\n\nAttach with:\n\n  zellij attach %s", quoted)
}

func (m *Model) attachZellijSessionCmd(sessionName string) tea.Cmd {
	// #nosec G204 -- zellij session name comes from user configuration.
	c := m.commandRunner(m.ctx, "zellij", multiplexer.OnExistsAttach, sessionName)
	return m.execProcess(c, func(err error) tea.Msg {
		if err != nil {
			return errMsg{err: err}
		}
		return refreshCompleteMsg{}
	})
}

// getZellijActiveSessions queries zellij for all sessions starting with the configured session prefix
// Returns session names with the prefix stripped, or empty slice if zellij is unavailable.
func (m *Model) getZellijActiveSessions() []string {
	// Check if zellij is available
	if _, err := exec.LookPath("zellij"); err != nil {
		return nil
	}

	// Query zellij for session list
	// #nosec G204 -- static command with format string
	cmd := m.commandRunner(m.ctx, "zellij", "list-sessions", "--short")
	output, err := cmd.Output()
	if err != nil {
		// zellij not running or no sessions
		return nil
	}

	// Parse output and filter for worktree session prefix
	var sessions []string
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, m.config.SessionPrefix) {
			// Strip worktree prefix
			sessionName := strings.TrimPrefix(line, m.config.SessionPrefix)
			if sessionName != "" {
				sessions = append(sessions, sessionName)
			}
		}
	}

	// Sort alphabetically for consistent display
	sort.Strings(sessions)
	return sessions
}

func sanitizeZellijSessionName(name string) string {
	return multiplexer.SanitizeZellijSessionName(name)
}

func buildTmuxInfoMessage(sessionName string, insideTmux bool) string {
	quoted := multiplexer.ShellQuote(sessionName)
	if insideTmux {
		return fmt.Sprintf("tmux session ready.\n\nSwitch with:\n\n  tmux switch-client -t %s", quoted)
	}
	return fmt.Sprintf("tmux session ready.\n\nAttach with:\n\n  tmux attach-session -t %s", quoted)
}

func (m *Model) attachTmuxSessionCmd(sessionName string, insideTmux bool) tea.Cmd {
	args := []string{"attach-session", "-t", sessionName}
	if insideTmux {
		args = []string{"switch-client", "-t", sessionName}
	}
	// #nosec G204 -- tmux session name comes from user configuration.
	c := m.commandRunner(m.ctx, "tmux", args...)
	return m.execProcess(c, func(err error) tea.Msg {
		if err != nil {
			return errMsg{err: err}
		}
		return refreshCompleteMsg{}
	})
}

func readTmuxSessionFile(path, fallback string) string {
	return multiplexer.ReadSessionFile(path, fallback)
}

func buildTmuxScript(sessionName string, tmuxCfg *config.TmuxCommand, windows []multiplexer.ResolvedWindow, env map[string]string) string {
	return multiplexer.BuildTmuxScript(sessionName, tmuxCfg, windows, env)
}

func buildZellijScript(sessionName string, zellijCfg *config.TmuxCommand, layoutPaths []string) string {
	return multiplexer.BuildZellijScript(sessionName, zellijCfg, layoutPaths)
}

func writeZellijLayouts(windows []multiplexer.ResolvedWindow) ([]string, error) {
	return multiplexer.WriteZellijLayouts(windows)
}

func sanitizeTmuxSessionName(name string) string {
	return multiplexer.SanitizeTmuxSessionName(name)
}

// getTmuxActiveSessions queries tmux for all sessions starting with the configured session prefix
// Returns session names with the prefix stripped, or empty slice if tmux is unavailable.
func (m *Model) getTmuxActiveSessions() []string {
	// Check if tmux is available
	if _, err := exec.LookPath("tmux"); err != nil {
		return nil
	}

	// Query tmux for session list
	// #nosec G204 -- static command with format string
	cmd := m.commandRunner(m.ctx, "tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		// tmux not running or no sessions
		return nil
	}

	// Parse output and filter for worktree session prefix
	var sessions []string
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, m.config.SessionPrefix) {
			// Strip worktree prefix
			sessionName := strings.TrimPrefix(line, m.config.SessionPrefix)
			if sessionName != "" {
				sessions = append(sessions, sessionName)
			}
		}
	}

	// Sort alphabetically for consistent display
	sort.Strings(sessions)
	return sessions
}

func resolveTmuxWindows(windows []config.TmuxWindow, env map[string]string, defaultCwd string) ([]multiplexer.ResolvedWindow, bool) {
	return multiplexer.ResolveTmuxWindows(windows, env, defaultCwd)
}

func buildTmuxWindowCommand(command string, env map[string]string) string {
	return multiplexer.BuildTmuxWindowCommand(command, env)
}

func (m *Model) openTmuxSession(tmuxCfg *config.TmuxCommand, wt *models.WorktreeInfo) tea.Cmd {
	if tmuxCfg == nil {
		return nil
	}

	env := m.buildCommandEnv(wt.Branch, wt.Path)
	insideTmux := os.Getenv("TMUX") != ""
	sessionName := expandWithEnv(tmuxCfg.SessionName, env)
	if strings.TrimSpace(sessionName) == "" {
		sessionName = fmt.Sprintf("%s%s", m.config.SessionPrefix, filepath.Base(wt.Path))
	}
	sessionName = sanitizeTmuxSessionName(sessionName)

	resolved, ok := resolveTmuxWindows(tmuxCfg.Windows, env, wt.Path)
	if !ok {
		return func() tea.Msg {
			return errMsg{err: fmt.Errorf("failed to resolve tmux windows")}
		}
	}

	sessionFile, err := os.CreateTemp("", "lazyworktree-tmux-")
	if err != nil {
		return func() tea.Msg {
			return errMsg{err: err}
		}
	}
	sessionPath := sessionFile.Name()
	if closeErr := sessionFile.Close(); closeErr != nil {
		return func() tea.Msg {
			return errMsg{err: closeErr}
		}
	}

	scriptCfg := *tmuxCfg
	scriptCfg.Attach = false
	env["LW_TMUX_SESSION_FILE"] = sessionPath
	script := buildTmuxScript(sessionName, &scriptCfg, resolved, env)
	// #nosec G204 -- command is built from user-configured tmux session settings.
	c := m.commandRunner(m.ctx, "bash", "-lc", script)
	c.Dir = wt.Path
	c.Env = append(os.Environ(), envMapToList(env)...)

	return m.execProcess(c, func(err error) tea.Msg {
		defer func() {
			_ = os.Remove(sessionPath)
		}()
		if err != nil {
			return errMsg{err: err}
		}
		finalSession := readTmuxSessionFile(sessionPath, sessionName)
		return tmuxSessionReadyMsg{
			sessionName: finalSession,
			attach:      tmuxCfg.Attach,
			insideTmux:  insideTmux,
		}
	})
}

func (m *Model) openZellijSession(zellijCfg *config.TmuxCommand, wt *models.WorktreeInfo) tea.Cmd {
	if zellijCfg == nil {
		return nil
	}

	env := m.buildCommandEnv(wt.Branch, wt.Path)
	insideZellij := os.Getenv("ZELLIJ") != "" || os.Getenv("ZELLIJ_SESSION_NAME") != ""
	sessionName := strings.TrimSpace(expandWithEnv(zellijCfg.SessionName, env))
	if sessionName == "" {
		sessionName = fmt.Sprintf("%s%s", m.config.SessionPrefix, filepath.Base(wt.Path))
	}
	sessionName = sanitizeZellijSessionName(sessionName)

	resolved, ok := resolveTmuxWindows(zellijCfg.Windows, env, wt.Path)
	if !ok {
		return func() tea.Msg {
			return errMsg{err: fmt.Errorf("failed to resolve zellij windows")}
		}
	}

	layoutPaths, err := writeZellijLayouts(resolved)
	if err != nil {
		return func() tea.Msg {
			return errMsg{err: err}
		}
	}

	sessionFile, err := os.CreateTemp("", "lazyworktree-zellij-")
	if err != nil {
		cleanupZellijLayouts(layoutPaths)
		return func() tea.Msg {
			return errMsg{err: err}
		}
	}
	sessionPath := sessionFile.Name()
	if closeErr := sessionFile.Close(); closeErr != nil {
		cleanupZellijLayouts(layoutPaths)
		return func() tea.Msg {
			return errMsg{err: closeErr}
		}
	}

	scriptCfg := *zellijCfg
	scriptCfg.Attach = false
	env["LW_ZELLIJ_SESSION_FILE"] = sessionPath
	script := buildZellijScript(sessionName, &scriptCfg, layoutPaths)
	// #nosec G204 -- command is built from user-configured zellij session settings.
	c := m.commandRunner(m.ctx, "bash", "-lc", script)
	c.Dir = wt.Path
	c.Env = append(os.Environ(), envMapToList(env)...)

	return m.execProcess(c, func(err error) tea.Msg {
		defer func() {
			_ = os.Remove(sessionPath)
			cleanupZellijLayouts(layoutPaths)
		}()
		if err != nil {
			return errMsg{err: err}
		}
		finalSession := readTmuxSessionFile(sessionPath, sessionName)
		return zellijSessionReadyMsg{
			sessionName:  finalSession,
			attach:       zellijCfg.Attach,
			insideZellij: insideZellij,
		}
	})
}
