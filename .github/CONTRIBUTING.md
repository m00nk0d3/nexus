# Contributing to Nexus

## Development Workflow

Nexus uses a **Pull Request (PR)-based workflow** with strict TDD enforcement. All changes to `main` must go through PR review.

### Branch Naming Convention

```
feature/issue-{NUMBER}    # For feature development
fix/issue-{NUMBER}        # For bug fixes
chore/issue-{NUMBER}      # For documentation, deps, cleanup
```

Example: `feature/issue-2`, `fix/issue-15`

### Development Cycle (Red-Green-Refactor)

1. **Check out a feature branch**
   ```bash
   git checkout -b feature/issue-XX main
   ```

2. **RED Phase** - Write failing test first
   - Use the `tdd-enforcer` global agent
   - Tests must fail before implementation
   - Example: `TestWorktreeList_NoWorktrees_ReturnsEmpty`

3. **GREEN Phase** - Implement minimal code
   - Use the `go-specialist` global agent
   - Make the test pass with minimal implementation
   - Don't add extra features yet

4. **REFACTOR Phase** - Improve code quality
   - Use the `testing-specialist` global agent
   - Ensure all tests still pass
   - Improve readability, performance, error handling

5. **Push to feature branch**
   ```bash
   git push origin feature/issue-XX
   ```

6. **Open a Pull Request**
   - Reference the issue: `Closes #XX`
   - Link all related 29 GitHub issues by phase
   - Ensure `copilot-setup-steps` workflow passes ✅

7. **Merge the PR**
   - Use **Squash and merge** for cleaner history (optional)
   - Feature branch auto-deletes after merge

## Testing Framework

We use **Testify** for all tests, not Go's standard `testing` package.

### Test Naming Convention

```go
func TestCOMPONENT_Scenario_ExpectedResult(t *testing.T)
```

**Example:**
```go
func TestWorktreeList_MultipleWorktrees_ReturnsSorted(t *testing.T) {
    // Arrange
    // Act
    // Assert
}
```

### Table-Driven Tests

All tests follow table-driven patterns for clarity and maintainability:

```go
tests := []struct {
    name      string
    input     interface{}
    expected  interface{}
    wantErr   bool
}{
    {
        name:     "valid input",
        input:    "test",
        expected: "result",
        wantErr:  false,
    },
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test logic here
    })
}
```

### Test Commands

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/domain

# Run tests matching a pattern
go test -run TestWorktree ./...
```

### Coverage Targets

- **Domain logic**: 90%+
- **Data layer**: 85%+
- **Execution layer**: 80%+
- **TUI layer**: 70%+

## Global Agents

### When to Use Each Agent

**`go-specialist`** - For implementation
- Writing production code (after test passes)
- Error handling and recovery
- Performance-critical sections

**`tdd-enforcer`** - For test-first development
- RED phase: Write failing tests first
- GREEN phase: Implement minimal code
- Ensures strict Red-Green-Refactor cycle

**`testing-specialist`** - For test architecture
- Test case design and coverage
- Mocking and test isolation
- Performance testing strategies

**`github-worktree-specialist`** - For Git operations
- Creating feature branches
- Managing worktrees
- GitHub API operations

## Project Structure

```
nexus/
├── cmd/
│   └── nexus/
│       └── main.go              # Application entry point
├── internal/
│   ├── domain/                  # Business logic (90%+ coverage)
│   │   ├── worktree.go
│   │   └── worktree_test.go
│   ├── data/                    # SQLite persistence (85%+ coverage)
│   │   ├── sqlite.go
│   │   └── sqlite_test.go
│   ├── exec/                    # Git command execution (80%+ coverage)
│   │   ├── git.go
│   │   └── git_test.go
│   └── tui/                     # Bubbletea TUI (70%+ coverage)
│       ├── app.go
│       └── app_test.go
├── docs/
│   └── PLAN.md                  # Architecture and design docs
├── go.mod                       # Dependencies
└── .github/
    ├── workflows/
    │   └── copilot-setup-steps.yml  # Copilot environment setup
    └── CONTRIBUTING.md          # This file
```

## Dependencies

- **Charm.sh** - TUI framework (bubbletea, lipgloss, bubbles)
- **github.com/go-github/v60** - GitHub API client (Phase 2)
- **sqlite** - Local data persistence (Phase 2+)
- **testify** - Testing framework (all phases)

## GitHub Issues by Phase

### [Phase 1: Foundation](https://github.com/m00nk0d3/nexus/issues?q=label%3Aphase-1)
Basic worktree management in TUI - 7 issues

### [Phase 2: GitHub Integration](https://github.com/m00nk0d3/nexus/issues?q=label%3Aphase-2)
Real-time GitHub sync + PR linking - 8 issues

### [Phase 3: AI Agent Integration](https://github.com/m00nk0d3/nexus/issues?q=label%3Aphase-3)
Spawn Copilot, Claude, Aider with context - 7 issues

### [Phase 4: Polish & Refinement](https://github.com/m00nk0d3/nexus/issues?q=label%3Aphase-4)
Production-ready v1.0 - 7 issues

## Tips for Success

1. **Always start with a failing test** - RED phase is non-negotiable
2. **Keep commits small and focused** - One issue per PR
3. **Use issue templates** - Reference GitHub issue in PR description
4. **Test your test** - Verify it fails before implementing
5. **Run full test suite before pushing** - `go test ./...`
6. **Keep the main branch deployable** - Tests must pass for merge
7. **Review your own code in the PR** - Catch issues before merge

## Questions?

For development environment questions, refer to `.copilot/README.md` for Copilot cloud agent integration and global agent usage.
