# TDD Enforcer Memory — nexus (m00nk0d3/nexus)

## Project Stack
- Go (backend only), located at repo root
- Test runner: go test ./internal/...

## Test Commands
- Run all tests: go test ./internal/exec/... ./internal/domain/...
- Run with verbose output: go test -v ./internal/exec/... ./internal/domain/...
- Filter specific test: go test -run TestName ./internal/exec/...

## Architecture Pattern
- internal/domain/ — pure structs, no deps (Issue, PullRequest, Worktree…)
- internal/exec/ — gh/git CLI wrappers; use injected commandRunner for testability
- Constructor pair: NewXxx(repoPath) calls NewXxxWithRunner(repoPath, runGhCommand)
- Tests always use NewXxxWithRunner with a mock runner func

## Test Conventions (exec package)
- Mock runner: unc(_ string, args ...string) (string, error) { return output, err }
- Capture args: ar capturedArgs []string inside the runner closure
- Table-driven tests for happy path, error, empty, edge cases in one 	ests := []struct{...} block
- Use 	estify/assert + 	estify/require; equire.NoError, ssert.Contains, ssert.Equal

## Error Wrapping Convention
- mt.Errorf("list open prs: %w", err) — verb phrase prefix, then %w
- Parse errors: mt.Errorf("parse pr list: %w", err)
- Single-item fetch: mt.Errorf("get pr: %w", err) / mt.Errorf("parse pr: %w", err)

## JSON Mapping Pattern (gh CLI)
- Inner structs (ghLabel, ghAuthor, ghIssue, ghPR) kept unexported in the exec package
- Labels always mapped via make([]string, len(g.Labels)) to guarantee non-nil slice
- Shared mapping extracted to ghXxxToDomain(g ghXxx) domain.Xxx helper to avoid duplication

## Completed Issues
- Issue #8: Added domain.PullRequest, GitHubClient with ListOpenPRs/GetPR — 12 new tests, all green
