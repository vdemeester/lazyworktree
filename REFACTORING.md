# App.go Refactoring Guide

This document outlines the refactoring of `internal/app/app.go` (4450 lines) into 7 focused files.

## Current Status

✅ **Completed:**
- `handlers.go` - 458 lines - Keyboard/input handling
- `messages.go` - 164 lines - Message dispatch and handling

🚧 **In Progress / Planned:**
- `commands.go` - ~800 lines - Worktree operations (create, delete, rename, etc.)
- `ui.go` - ~1200 lines - Rendering and layout
- `state.go` - ~800 lines - State management, caching, persistence
- `utils.go` - ~250 lines - Utility functions
- `app.go` - ~200 lines - Core types and Bubble Tea interface

## File Breakdown

### app.go (Core - ~200 lines)
**Keep only:**
- Package declaration and imports
- Constants (keyEnter, keyEsc, errBranchEmpty, etc.)
- Type definitions (Model, message types, layoutDims, etc.)
- `NewModel()` constructor
- `Init()` - Bubble Tea interface
- `Update()` - Bubble Tea interface (dispatches to handlers)
- `View()` - Bubble Tea interface (dispatches to UI)
- Core helpers: `screenName()`, `GetSelectedPath()`, `Close()`

### handlers.go (458 lines) ✅
**Keyboard & Input:**
- `handleKeyMsg()` - Main keyboard dispatcher
- `handleBuiltInKey()` - Built-in shortcuts
- `handleNavigationDown/Up()` - Arrow key navigation
- `handlePageDown/Up()` - Page navigation
- `handleEnterKey()` - Enter key handling
- `handleFilterNavigation()` - Filter pane navigation
- `setFilterToWorktree()` - Filter manipulation
- `selectFilteredWorktree()` - Selection
- `handleMouse()` - Mouse events

### messages.go (164 lines) ✅
**Message Handling:**
- `handleWorktreeMessages()` - Dispatcher
- `handleWorktreesLoaded()` - Worktree update
- `handleCachedWorktrees()` - Cache update
- `handlePruneResult()` - Prune completion
- `handleAbsorbResult()` - Absorb completion
- `handlePRMessages()` - PR dispatcher
- `handlePRDataLoaded()` - PR data update
- `handleCIStatusLoaded()` - CI update
- `handleOpenPRsLoaded()` - PR list
- `handleCreateFromChangesReady()` - Changes dialog
- `handleCherryPickResult()` - Cherry-pick completion

### commands.go (~800 lines) 🚧
**Worktree Operations:**
- `showCreateWorktree()` - Initiate creation
- `showCreateFromPR()` - PR creation
- `showCreateWorktreeFromChanges()` - Changes-based creation
- `showCreateFromChangesInput()` - Input handling for changes
- `handleCreateFromChangesReady()` - Ready callback
- `showDeleteWorktree()` - Delete confirmation
- `showDiff()` - Show diff
- `showRenameWorktree()` - Rename
- `showRunCommand()` - Arbitrary command
- `showPruneMerged()` - Prune merged
- `showAbsorbWorktree()` - Absorb merge
- `showCommandPalette()` - Command palette
- `customPaletteItems()` - Custom commands
- `customCommandKeys()` - Keys
- `customCommandLabel()` - Labels
- `deleteWorktreeCmd()` - Delete command
- `executeCustomCommand()` - Custom execution
- `executeCustomCommandWithPager()` - Paged output
- `executeArbitraryCommand()` - Arbitrary execution
- `openLazyGit()` - LazyGit integration
- `openTmuxSession()` - Tmux
- `openPR()` - Open PR in browser
- `showCherryPick()` - Cherry-pick UI
- `executeCherryPick()` - Cherry-pick execution
- `openCommitView()` - Commit details

### ui.go (~1200 lines) 🚧
**Rendering & Layout:**
- `View()` core rendering (overlay management)
- `overlayPopup()` - Popup overlay
- `renderScreen()` - Screen dispatcher
- `computeLayout()` - Layout calculation
- `applyLayout()` - Apply dimensions
- `setWindowSize()` - Window sizing
- `renderHeader()` - Header
- `renderFilter()` - Filter bar
- `renderBody()` - Main body
- `renderLeftPane()` - Worktree table
- `renderRightPane()` - Info/log panes
- `renderRightTopPane()` - Info pane
- `renderRightBottomPane()` - Log pane
- `renderFooter()` - Footer hints
- `renderPaneTitle()` - Pane titles
- `renderInnerBox()` - Inner boxes
- `buildInfoContent()` - Info formatting
- `buildStatusContent()` - Status formatting
- `updateTable()` - Table updates
- `updateDetailsView()` - Details update
- `debouncedUpdateDetailsView()` - Debounced update
- `updateTableColumns()` - Column layout
- `updateLogColumns()` - Log columns
- `customFooterHints()` - Footer hints
- `renderKeyHint()` - Key hints
- `basePaneStyle()` - Base style
- `paneStyle()` - Pane styling
- `baseInnerBoxStyle()` - Box styling
- `truncateToHeight()` - Text truncation

### state.go (~800 lines) 🚧
**State & Persistence:**
Data Fetching:
- `refreshWorktrees()` - Fetch worktrees
- `fetchPRData()` - Fetch PR data
- `fetchCIStatus()` - Fetch CI status
- `maybeFetchCIStatus()` - Conditional fetch
- `fetchRemotes()` - Fetch remotes

Caching:
- `loadCache()` - Load cache
- `saveCache()` - Save cache
- `getCachedDetails()` - Get cached details

Command History:
- `loadCommandHistory()` - Load history
- `saveCommandHistory()` - Save history
- `addToCommandHistory()` - Add to history

Persistence:
- `persistCurrentSelection()` - Save current
- `persistLastSelected()` - Save selection
- `ensureRepoConfig()` - Ensure config

Configuration & Utilities:
- `getRepoKey()` - Get repo key
- `getMainWorktreePath()` - Get main path
- `getWorktreeDir()` - Get directory
- `buildCommandEnv()` - Build env vars
- `showInfo()` - Show info screen
- `debugf()` - Debug logging
- `pagerCommand()` - Pager selection
- `pagerEnv()` - Pager environment

### utils.go (~250 lines) 🚧
**Utility Functions (package-level):**
- `parseCommitMeta()` - Parse commit metadata
- `sanitizePRURL()` - Sanitize PR URLs
- `isEscKey()` - Escape key detection
- `pagerIsLess()` - Detect less pager
- `expandWithEnv()` - Environment expansion
- `envMapToList()` - Convert env map
- `shellQuote()` - Shell quoting
- `resolveTmuxWindows()` - Tmux window resolution
- `buildTmuxWindowCommand()` - Build tmux command
- `exportEnvCommand()` - Build env export
- `buildTmuxScript()` - Build tmux script
- `buildTmuxInfoMessage()` - Tmux info message
- `readTmuxSessionFile()` - Read session file

## Implementation Notes

### Import Changes
Each new file imports only what it needs:
- `handlers.go`: `path/filepath`, `sort`, `textinput`, `tea`, `models`
- `messages.go`: `fmt`, `os`, `path/filepath`, `strings`, `time`, `textinput`, `tea`, `models`
- `commands.go`: Would add `exec`, `runtime`, `config` (for custom commands)
- `ui.go`: `lipgloss`, `wrap`, layout types
- `state.go`: Core imports for caching, file I/O
- `utils.go`: Minimal imports, mostly standard library

### Function Dependencies
- `handlers.go` calls `updateTable()`, `debouncedUpdateDetailsView()`, `executeCustomCommand()` (moved to commands.go)
- `messages.go` calls `buildInfoContent()`, `updateTable()` (moved to ui.go and state.go)
- `commands.go` calls many state and UI functions
- `ui.go` is relatively self-contained except for `buildInfoContent()`, `buildStatusContent()`
- `state.go` is relatively self-contained

### Testing Strategy
1. After each file creation, run `make build` to check compilation
2. Run `make sanity` (golangci-lint, gofumpt, go test) before final commit
3. Verify no functional changes - only reorganization

## Benefits

✅ **Improved Maintainability**
- Each file ~200-1200 lines (manageable)
- Clear separation of concerns
- Easier to find related functionality

✅ **Better Testability**
- Can test handlers independently
- Can mock messages
- UI rendering isolated

✅ **Cleaner Codebase**
- Reduced cognitive load
- Easier onboarding for new developers
- Better git blame history

✅ **Flexible Architecture**
- Can refactor UI without touching handlers
- Can swap out command implementations
- Can extend handlers independently

## Migration Order

1. ✅ Create `handlers.go` - No dependencies on new files
2. ✅ Create `messages.go` - Can reference app.go functions
3. ⏭️ Create `utils.go` - Pure functions, no Model dependencies
4. ⏭️ Create `commands.go` - Calls UI and state functions
5. ⏭️ Create `ui.go` - Calls state functions
6. ⏭️ Create `state.go` - Mostly independent
7. ⏭️ Refactor `app.go` - Remove all moved functions
8. ⏭️ Verify `make sanity` passes

## Notes for Implementers

When creating each file:
1. Copy the function implementations from `app_original.go`
2. Add appropriate imports
3. Ensure no duplicate declarations
4. Test compilation after each file
5. Update app.go to remove moved functions (be careful with line ranges!)
6. Run full test suite to verify no regressions

The refactoring is structural only - no logic changes.
