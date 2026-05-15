-- worktrees: tracks local git worktrees with metadata
CREATE TABLE IF NOT EXISTS worktrees (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    path        TEXT NOT NULL UNIQUE,
    branch      TEXT NOT NULL,
    linked_pr   INTEGER,
    is_locked   BOOLEAN DEFAULT FALSE,
    last_used   DATETIME,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- github_prs: cache of GitHub PR data
CREATE TABLE IF NOT EXISTS github_prs (
    number      INTEGER PRIMARY KEY,
    title       TEXT NOT NULL,
    branch      TEXT NOT NULL,
    author      TEXT,
    state       TEXT,
    is_draft    BOOLEAN DEFAULT FALSE,
    labels      TEXT,
    synced_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- github_issues: cache of GitHub issue data
CREATE TABLE IF NOT EXISTS github_issues (
    number      INTEGER PRIMARY KEY,
    title       TEXT NOT NULL,
    state       TEXT,
    labels      TEXT,
    assignees   TEXT,
    synced_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- agent_history: log of AI agent invocations
CREATE TABLE IF NOT EXISTS agent_history (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_name  TEXT NOT NULL,
    worktree_id INTEGER REFERENCES worktrees(id),
    prompt      TEXT,
    exit_code   INTEGER,
    started_at  DATETIME,
    ended_at    DATETIME
);

-- context_snapshots: worktree context captured before agent launch
CREATE TABLE IF NOT EXISTS context_snapshots (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    worktree_id INTEGER REFERENCES worktrees(id),
    git_status  TEXT,
    file_list   TEXT,
    recent_log  TEXT,
    captured_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
