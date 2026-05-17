package data

import (
	"fmt"
	"time"
)

// AgentHistoryEntry holds the data for a single AI agent invocation to be
// persisted in the agent_history table.
type AgentHistoryEntry struct {
	AgentName  string // Name of the agent, e.g. "copilot"
	WorktreeID *int64 // Optional FK to worktrees.id; nil if not linked
	Prompt     string // The prompt that was sent to the agent
	ExitCode   int    // Process exit code (0 = success)
	StartedAt  time.Time
	EndedAt    time.Time
}

// LogAgentRun inserts an agent invocation record into the agent_history table.
// Returns a wrapped error on failure so callers can detect "log agent run" context.
func LogAgentRun(db *DB, entry AgentHistoryEntry) error {
	_, err := db.Conn.Exec(
		`INSERT INTO agent_history (agent_name, worktree_id, prompt, exit_code, started_at, ended_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		entry.AgentName,
		entry.WorktreeID,
		entry.Prompt,
		entry.ExitCode,
		entry.StartedAt,
		entry.EndedAt,
	)
	if err != nil {
		return fmt.Errorf("log agent run: %w", err)
	}
	return nil
}
