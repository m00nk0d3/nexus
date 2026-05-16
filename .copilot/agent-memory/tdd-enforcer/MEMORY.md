# TDD Enforcer Memory - nexus (m00nk0d3/nexus)

## Project Stack
- Go TUI app (Bubbletea), package `main` in `cmd/nexus/`
- Test runner: `go test ./cmd/nexus/... -v`
- Build: `go build ./...`

## Test Commands
- Run all tests: `go test ./cmd/nexus/...`
- Run renderer tests: `go test ./cmd/nexus/... -run "TestRenderer"`
- Filter specific test: `go test ./cmd/nexus/... -run TestName`

## Architecture Pattern
- `internal/domain/` -- pure structs (Issue, PullRequest, Worktree)
- `cmd/nexus/app.go` -- BubbleTea Model, Update, View, NewModel()
- `cmd/nexus/renderer.go` -- pure render functions called from View()
- Tests use `NewModel()` + set unexported fields directly (same `main` package)

## TUI Test Conventions (renderer_test.go)
- All tests via `model.View()` -- no direct render function calls for context panel
- Set state: `model.view = viewWorktrees/viewIssues/viewPRs`, `model.Worktrees`, `model.selectedIdx`
- Lipgloss ANSI codes stripped in test env (no TTY), raw text is directly assertable
- Use `assert.Contains(t, view, want)` for positive assertions
- Use `assert.NotContains(t, view, notWant)` for negative assertions
- `model.width = 300` for wide-terminal tests to prevent truncation

## Key Model Fields (accessible in same package)
- `model.selectedIdx` -- worktree selection
- `model.selectedIssueIdx`, `model.selectedPRIdx`
- `model.view`, `model.Worktrees`, `model.issues`, `model.prs`
- `model.syncing`, `model.lastSynced`, `model.syncErr`, `model.width`

## Renderer Helpers
- `worktreeStatus(wt)` returns "Idle" | "Dirty" | "Locked"
- `truncateStr(s, n)` -- clips to n runes with "..."
- `formatLabels([]string)` -- "[label1][label2]" format (shared between issue + PR contexts)
- `prStateColor(state)` -- lipgloss.Color: OPEN=#00D9FF, MERGED=#9B59B6, CLOSED=#E74C3C

## GH:ID Column (renderWorktreePanel)
- Default ghID = "-" when LinkedPR == nil (not empty string)
- Non-selected rows: colorize with lipgloss (OPEN=cyan, MERGED/CLOSED=grey #4A5568)
- Selected rows: ghID embedded in fmt.Sprintf string (no individual coloring needed)

## Context Panel -- Worktree View (renderContextPanel default case)
- With LinkedPR: shows "Context: PR #N", title (truncated), "GH Title:", "Author: @", "Status: * STATE", "Labels: [x][y]", AGENT COMMANDS
- Without LinkedPR: shows "Context: basename", "Branch: ", "Path: ", AGENT COMMANDS
- Agent commands (always present): `[a] Spawn Claude Code`, `[c] Spawn Copilot`, `[s] Open Shell in WT`

## Completed Issues
- Issue #8: Added domain.PullRequest, GitHubClient with ListOpenPRs/GetPR
- Issue #14: Added Issues/PRs views, renderIssueList, renderPRList, context panel for Issues+PRs
- Issue #15: PR context panel in worktree view + GH:ID column colorization + "-" default when no PR

## exec package conventions (internal/exec/)
- Constructor pair: NewXxx(repoPath) calls NewXxxWithRunner(repoPath, runGhCommand)
- Tests always use NewXxxWithRunner with a mock runner func
- Error wrapping: `fmt.Errorf("verb phrase: %w", err)` pattern
- JSON inner structs (ghLabel, ghAuthor, etc.) kept unexported in exec package
