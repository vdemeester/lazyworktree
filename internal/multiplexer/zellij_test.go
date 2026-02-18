package multiplexer

import (
	"os"
	"strings"
	"testing"

	"github.com/chmouel/lazyworktree/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeZellijSessionName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "simple name",
			input: "feature-branch",
			want:  "feature-branch",
		},
		{
			name:  "name with slash",
			input: "feature/branch",
			want:  "feature-branch",
		},
		{
			name:  "name with backslash",
			input: "feature\\branch",
			want:  "feature-branch",
		},
		{
			name:  "name with colon",
			input: "feature:branch",
			want:  "feature-branch",
		},
		{
			name:  "name with multiple special chars",
			input: "feature/foo:bar\\baz",
			want:  "feature-foo-bar-baz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeZellijSessionName(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestKdlQuote(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple string",
			input: "hello",
			want:  `"hello"`,
		},
		{
			name:  "string with quotes",
			input: `say "hello"`,
			want:  `"say \"hello\""`,
		},
		{
			name:  "string with backslash",
			input: `path\to\file`,
			want:  `"path\\to\\file"`,
		},
		{
			name:  "string with both",
			input: `path\"quote`,
			want:  `"path\\\"quote"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := KdlQuote(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildZellijTabLayout(t *testing.T) {
	t.Run("basic layout", func(t *testing.T) {
		window := ResolvedWindow{
			Name:    "main",
			Command: "vim",
			Cwd:     "/path",
		}
		layout := BuildZellijTabLayout(window)
		assert.Contains(t, layout, "layout {")
		assert.Contains(t, layout, `tab name="main"`)
		assert.Contains(t, layout, `command "bash"`)
		assert.Contains(t, layout, `cwd "/path"`)
		assert.Contains(t, layout, "vim")
	})

	t.Run("layout without cwd", func(t *testing.T) {
		window := ResolvedWindow{
			Name:    "main",
			Command: "vim",
			Cwd:     "",
		}
		layout := BuildZellijTabLayout(window)
		assert.NotContains(t, layout, "cwd")
	})

	t.Run("layout with special chars", func(t *testing.T) {
		window := ResolvedWindow{
			Name:    `test"name`,
			Command: `echo "hello"`,
			Cwd:     `/path\to\file`,
		}
		layout := BuildZellijTabLayout(window)
		assert.Contains(t, layout, `tab name="test\"name"`)
		assert.Contains(t, layout, `cwd "/path\\to\\file"`)
	})
}

func TestWriteZellijLayouts(t *testing.T) {
	t.Run("empty windows", func(t *testing.T) {
		paths, err := WriteZellijLayouts([]ResolvedWindow{})
		require.NoError(t, err)
		assert.Empty(t, paths)
	})

	t.Run("single window", func(t *testing.T) {
		windows := []ResolvedWindow{
			{Name: "main", Command: "vim", Cwd: "/path"},
		}
		paths, err := WriteZellijLayouts(windows)
		require.NoError(t, err)
		require.Len(t, paths, 1)

		// Verify file exists and has content
		content, err := os.ReadFile(paths[0])
		require.NoError(t, err)
		assert.Contains(t, string(content), "layout {")
		assert.Contains(t, string(content), "main")

		// Cleanup
		CleanupZellijLayouts(paths)

		// Verify cleanup
		_, err = os.Stat(paths[0])
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("multiple windows", func(t *testing.T) {
		windows := []ResolvedWindow{
			{Name: "first", Command: "vim", Cwd: "/path1"},
			{Name: "second", Command: "git status", Cwd: "/path2"},
		}
		paths, err := WriteZellijLayouts(windows)
		require.NoError(t, err)
		require.Len(t, paths, 2)

		// Verify both files exist
		for i, path := range paths {
			content, err := os.ReadFile(path) //#nosec G304 -- test file created by current process
			require.NoError(t, err)
			assert.Contains(t, string(content), windows[i].Name)
		}

		// Cleanup
		CleanupZellijLayouts(paths)
	})
}

func TestCleanupZellijLayouts(t *testing.T) {
	t.Run("cleanup multiple files", func(t *testing.T) {
		// Create temporary files
		tmpFile1, err := os.CreateTemp("", "zellij-layout-")
		require.NoError(t, err)
		_ = tmpFile1.Close()

		tmpFile2, err := os.CreateTemp("", "zellij-layout-")
		require.NoError(t, err)
		_ = tmpFile2.Close()

		paths := []string{tmpFile1.Name(), tmpFile2.Name()}

		// Cleanup
		CleanupZellijLayouts(paths)

		// Verify files are removed
		for _, path := range paths {
			_, err := os.Stat(path) //#nosec G703 -- test verification
			assert.True(t, os.IsNotExist(err))
		}
	})

	t.Run("cleanup non-existent files does not error", func(t *testing.T) {
		paths := []string{"/nonexistent/file1", "/nonexistent/file2"}
		CleanupZellijLayouts(paths) // Should not panic
	})
}

func TestBuildZellijScript(t *testing.T) {
	t.Run("basic script without layouts", func(t *testing.T) {
		cfg := &config.TmuxCommand{OnExists: "switch", Attach: false}
		script := BuildZellijScript("test-session", cfg, []string{})
		assert.Contains(t, script, "test-session")
		assert.Contains(t, script, "zellij attach --create-background")
		assert.Contains(t, script, "session_exists()")
	})

	t.Run("script with layout paths", func(t *testing.T) {
		cfg := &config.TmuxCommand{OnExists: "switch", Attach: false}
		layouts := []string{"/tmp/layout1.kdl", "/tmp/layout2.kdl"}
		script := BuildZellijScript("test-session", cfg, layouts)
		assert.Contains(t, script, "zellij action new-tab --layout")
		assert.Contains(t, script, "/tmp/layout1.kdl")
		assert.Contains(t, script, "/tmp/layout2.kdl")
		assert.Contains(t, script, "zellij action go-to-tab 1")
		assert.Contains(t, script, "zellij action close-tab")
	})

	t.Run("script with onExists kill", func(t *testing.T) {
		cfg := &config.TmuxCommand{OnExists: "kill", Attach: false}
		script := BuildZellijScript("test-session", cfg, []string{})
		assert.Contains(t, script, "zellij kill-session")
	})

	t.Run("script with onExists new", func(t *testing.T) {
		cfg := &config.TmuxCommand{OnExists: "new", Attach: false}
		script := BuildZellijScript("test-session", cfg, []string{})
		assert.Contains(t, script, "base_session")
		assert.Contains(t, script, "while session_exists")
	})

	t.Run("script with session file", func(t *testing.T) {
		cfg := &config.TmuxCommand{OnExists: "switch", Attach: false}
		script := BuildZellijScript("test-session", cfg, []string{})
		assert.Contains(t, script, "LW_ZELLIJ_SESSION_FILE")
	})

	t.Run("script with session timeout", func(t *testing.T) {
		cfg := &config.TmuxCommand{OnExists: "switch", Attach: false}
		script := BuildZellijScript("test-session", cfg, []string{})
		assert.Contains(t, script, "tries=0")
		assert.Contains(t, script, "if [ $tries -ge 50 ]")
		assert.Contains(t, script, "Timeout waiting for zellij session")
	})

	t.Run("script only creates tabs when session is new", func(t *testing.T) {
		cfg := &config.TmuxCommand{OnExists: "switch", Attach: false}
		layouts := []string{"/tmp/layout1.kdl"}
		script := BuildZellijScript("test-session", cfg, layouts)
		lines := strings.Split(script, "\n")

		// Find the layout creation section
		foundCreatedCheck := false
		for _, line := range lines {
			if strings.Contains(line, `if [ "$created" = "true" ]`) {
				foundCreatedCheck = true
			}
			if foundCreatedCheck && strings.Contains(line, "zellij action new-tab") {
				// Found layout creation inside created check
				return
			}
		}
		assert.True(t, foundCreatedCheck, "script should only create tabs when session is new")
	})
}
