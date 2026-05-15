package data_test

import (
	"database/sql"
	"testing"

	"github.com/m00nk0d3/nexus/internal/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

// tableExists queries sqlite_master to verify that a table was created by
// the schema migration.  It accepts the raw *sql.DB so the helper stays
// independent of the data.DB wrapper.
func tableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var found string
	err := db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name=?",
		name,
	).Scan(&found)
	return err == nil && found == name
}

// TestNewDB_CreatesSchema verifies that NewDB opens a valid connection and
// runs the DDL migration that produces all five required tables.
func TestNewDB_CreatesSchema(t *testing.T) {
	db, err := data.NewDB(":memory:")
	require.NoError(t, err, "NewDB should open without error")
	require.NotNil(t, db, "NewDB should return a non-nil DB")
	t.Cleanup(func() { _ = db.Close() })

	tables := []string{
		"worktrees",
		"github_prs",
		"github_issues",
		"agent_history",
		"context_snapshots",
	}

	for _, table := range tables {
		assert.Truef(t, tableExists(t, db.DB, table),
			"expected table %q to exist after schema migration", table)
	}
}

// TestNewDB_WorktreeInsertAndQuery verifies that the worktrees table accepts
// a row with (path, branch) and that the row can be retrieved by path.
func TestNewDB_WorktreeInsertAndQuery(t *testing.T) {
	db, err := data.NewDB(":memory:")
	require.NoError(t, err)
	require.NotNil(t, db)
	t.Cleanup(func() { _ = db.Close() })

	const (
		wantPath   = "/repo/worktrees/feat-issue-10"
		wantBranch = "feat/issue-10-sqlite-db-setup"
	)

	_, err = db.DB.Exec(
		"INSERT INTO worktrees (path, branch) VALUES (?, ?)",
		wantPath, wantBranch,
	)
	require.NoError(t, err, "inserting a worktree row should succeed")

	var gotPath, gotBranch string
	err = db.DB.QueryRow(
		"SELECT path, branch FROM worktrees WHERE path = ?",
		wantPath,
	).Scan(&gotPath, &gotBranch)
	require.NoError(t, err, "querying worktree by path should return exactly one row")

	assert.Equal(t, wantPath, gotPath)
	assert.Equal(t, wantBranch, gotBranch)
}

// TestNewDB_GithubPRInsertAndQuery verifies that the github_prs table accepts
// a row with (number, title, branch, state) and that the row can be retrieved
// by PR number.
func TestNewDB_GithubPRInsertAndQuery(t *testing.T) {
	db, err := data.NewDB(":memory:")
	require.NoError(t, err)
	require.NotNil(t, db)
	t.Cleanup(func() { _ = db.Close() })

	const (
		wantNumber = 42
		wantTitle  = "feat: add SQLite persistence layer"
		wantBranch = "feat/issue-10-sqlite-db-setup"
		wantState  = "open"
	)

	_, err = db.DB.Exec(
		"INSERT INTO github_prs (number, title, branch, state) VALUES (?, ?, ?, ?)",
		wantNumber, wantTitle, wantBranch, wantState,
	)
	require.NoError(t, err, "inserting a github_pr row should succeed")

	var gotNumber int
	var gotTitle, gotBranch, gotState string
	err = db.DB.QueryRow(
		"SELECT number, title, branch, state FROM github_prs WHERE number = ?",
		wantNumber,
	).Scan(&gotNumber, &gotTitle, &gotBranch, &gotState)
	require.NoError(t, err, "querying github_pr by number should return exactly one row")

	assert.Equal(t, wantNumber, gotNumber)
	assert.Equal(t, wantTitle, gotTitle)
	assert.Equal(t, wantBranch, gotBranch)
	assert.Equal(t, wantState, gotState)
}

// TestDB_Close verifies that calling Close on an open DB returns no error and
// does not panic.
func TestDB_Close(t *testing.T) {
	db, err := data.NewDB(":memory:")
	require.NoError(t, err)
	require.NotNil(t, db)

	err = db.Close()
	assert.NoError(t, err, "Close should return no error on a live connection")
}
