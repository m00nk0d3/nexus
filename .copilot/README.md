# Copilot Agent Integration Guide

This document outlines how to use your global Copilot agents with the Nexus project, with a focus on TDD-first development.

## Global Agents Available

### 1. **tdd-enforcer** (Primary for new features)
- **Purpose**: Enforce strict Test-Driven Development (Red-Green-Refactor cycle)
- **Model**: Claude Sonnet 4.6 (with project memory)
- **When to use**: 
  - Starting any new feature (Phase 1-4)
  - Fixing bugs (write test first to reproduce)
  - Any change to `internal/` packages
- **Workflow**:
  1. Write a failing test (RED)
  2. Implement minimal code to pass (GREEN)
  3. Refactor while tests stay green (REFACTOR)
  4. Repeat for next behavior

**Example invocation**:
```
"Use the tdd-enforcer agent to implement the git worktree list functionality with tests first"
```

### 2. **go-specialist** (Implementation after tests)
- **Purpose**: Idiomatic Go code with strong error handling
- **When to use**:
  - After test structure is in place
  - For refactoring existing code
  - For performance optimization
- **Principles**:
  - Prefer clarity over cleverness
  - Small, focused functions
  - Explicit error handling
  - Context propagation for I/O

**Example invocation**:
```
"Use the go-specialist agent to implement the git command wrapper following the test we just wrote"
```

### 3. **testing-specialist** (Test architecture)
- **Purpose**: Comprehensive test strategy and coverage
- **Model**: Claude Sonnet 4.6 (with project memory)
- **When to use**:
  - Designing test strategy for complex features
  - Fixing failing tests
  - Test infrastructure improvements
  - Coverage analysis
- **Focus Areas**:
  - Unit tests (pure domain logic)
  - Integration tests (API/database operations)
  - Test patterns and helpers

**Example invocation**:
```
"Use the testing-specialist agent to design tests for the SQLite data layer"
```

### 4. **github-worktree-specialist** (Git operations)
- **Purpose**: Git worktree operations and GitHub interactions
- **When to use**:
  - Creating/deleting worktrees during development
  - GitHub issue/PR work
  - Git workflow questions

## Testing Framework: Testify

This project uses **Testify**, not the standard Go `testing` package.

### Key Testify Features Used

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
)

// Assertions
assert.Equal(t, expected, actual)
assert.NotNil(t, value)
require.NoError(t, err)

// Mocking
mockRepo := new(RepositoryMock)
mockRepo.On("Get", mock.MatchedBy(...)).Return(data, nil)
```

## Test Organization

```
internal/
├── domain/
│   ├── worktree.go          (domain logic)
│   └── worktree_test.go     (unit tests)
├── data/
│   ├── sqlite.go            (data access)
│   └── sqlite_test.go       (repository tests)
├── exec/
│   ├── git.go               (git command wrapper)
│   └── git_test.go          (command tests)
└── tui/
    ├── app.go               (TUI logic)
    └── app_test.go          (component tests)
```

## TDD Workflow for Nexus

### Step 1: RED - Write Failing Test
Create a test that defines the expected behavior. Run `go test ./...` to confirm it fails.

**Example**:
```go
func TestListWorktrees_ValidRepo_ReturnsWorktrees(t *testing.T) {
    // Arrange
    cmd := NewGitCommand("/tmp/test-repo")
    
    // Act
    worktrees, err := cmd.ListWorktrees()
    
    // Assert
    require.NoError(t, err)
    assert.NotEmpty(t, worktrees)
}
```

### Step 2: GREEN - Minimal Implementation
Write the absolute minimum code to make the test pass. Do NOT add extra logic.

**Example**:
```go
func (g *GitCommand) ListWorktrees() ([]Worktree, error) {
    return []Worktree{}, nil  // Minimal: passes test
}
```

### Step 3: REFACTOR - Clean Up
With tests passing, improve code quality. Run tests after each change.

**Example**:
```go
func (g *GitCommand) ListWorktrees() ([]Worktree, error) {
    // Execute: git worktree list --porcelain
    // Parse output and return results
    return g.executeWorktreeCommand()
}
```

## Running Tests

### Run all tests
```bash
go test ./...
```

### Run specific package tests
```bash
go test ./internal/domain
go test ./internal/data
```

### Run with verbose output
```bash
go test -v ./...
```

### Run specific test
```bash
go test -run TestListWorktrees ./internal/exec
```

### Run with coverage
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Naming Convention

Use the pattern: **MethodName_Scenario_ExpectedResult**

Examples:
- `TestListWorktrees_ValidRepo_ReturnsWorktrees`
- `TestCreateWorktree_InvalidBranch_ReturnsError`
- `TestSyncGitHub_NetworkFailure_UsesCache`
- `TestParseConfig_MissingToken_ReturnsError`

## Code Coverage Goals

Aim for these coverage targets:

| Package | Target Coverage |
|---------|-----------------|
| `internal/domain` | 90%+ (business logic is critical) |
| `internal/data` | 85%+ (persistence logic) |
| `internal/exec` | 80%+ (external command execution) |
| `internal/tui` | 70%+ (UI components harder to test) |

## Agent Invocation in Tasks

When delegating work to agents, provide complete context:

```
Use the tdd-enforcer agent to:
1. Write a failing test for the git worktree list operation
2. Implement the minimal code to pass
3. Refactor with proper error handling
4. Verify all tests pass

Context:
- Repository: m00nk0d3/nexus
- Issue: #3 (Implement - Git wrapper)
- Framework: Testify
- Pattern: Table-driven tests
```

## Troubleshooting

### Tests won't compile
- Check `go mod download` ran successfully
- Verify testify is installed: `go list github.com/stretchr/testify`

### Tests fail in Copilot but pass locally
- Check Go version: `go version` (should be 1.22+)
- Verify environment variables in `.github/workflows/copilot-setup-steps.yml`

### Coverage gaps
- Use testing-specialist agent to identify what's untested
- Prioritize domain logic and error paths

## Links

- [Nexus Plan](../../docs/PLAN.md)
- [GitHub Issues](https://github.com/m00nk0d3/nexus/issues)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Go Testing](https://pkg.go.dev/testing)
