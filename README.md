# Nexus

Nexus is a terminal-based worktree and AI-agent command center for managing Git worktrees, syncing GitHub metadata, and launching coding agents with the right repository context.

## What it is

Nexus is designed for developers who juggle multiple branches, worktrees, pull requests, and AI-assisted workflows. The goal is to keep that workflow in one place instead of bouncing between shells, browser tabs, and separate tools.

## Planned capabilities

- Manage Git worktrees from a TUI
- Create, delete, switch, lock, unlock, and prune worktrees
- Sync GitHub pull requests, issues, and branches
- Link PRs to worktrees by branch name
- Launch AI agents with current repo context
- Persist config, history, and cached metadata locally
- Use a high-contrast "Digital Noir" interface theme

## Core ideas

### Git worktree control

Nexus treats worktrees as the primary unit of navigation. The app will surface:

- worktree path
- branch name
- short commit SHA
- clean/dirty state
- lock state
- related GitHub PRs

### GitHub sync

The app will periodically fetch repository metadata and keep the UI updated without blocking the main event loop. Planned data includes:

- pull requests
- issues
- branches
- reviewers, labels, authors, and statuses

### AI agent launchers

Nexus is planned to support three launch targets:

- GitHub Copilot CLI
- Claude Code
- Aider

Each launcher will receive the current worktree context, changed files, and related git state before it starts.

## Proposed architecture

Nexus is planned as a Go TUI built with the Charm.sh ecosystem:

- `bubbletea` for MVU state management
- `lipgloss` for styling
- `bubbles` for reusable UI components
- `gh` CLI or `go-github` for GitHub integration
- SQLite for local persistence
- TOML for user config

## Planned layout

```text
nexus/
├── cmd/
│   └── nexus/
│       └── main.go
├── internal/
│   ├── tui/
│   │   ├── models/
│   │   ├── styles/
│   │   └── components/
│   ├── domain/
│   ├── data/
│   │   └── migrations/
│   └── exec/
├── docs/
│   └── PLAN.md
└── README.md
```

## UI concept

The target UI is a three-column dashboard:

1. Left rail for mode switching
2. Middle panel for lists and tables
3. Right panel for contextual details and actions

Planned screens:

- Worktree dashboard
- PR browser
- Issue browser
- Theme selector
- Help popup

## Keybindings

Planned keyboard shortcuts:

| Key | Action |
| --- | --- |
| `j` / `k` | Move within a panel |
| `h` / `l` | Switch panels |
| `Enter` | Open selected item |
| `Ctrl+N` | Create new worktree |
| `Ctrl+D` | Delete selected worktree |
| `Ctrl+L` | Lock or unlock worktree |
| `s` | Open shell in worktree |
| `a` | Spawn Claude Code |
| `c` | Spawn Copilot |
| `F1` / `?` | Open help |
| `t` | Open theme selector |
| `g` | Open selected item in GitHub |
| `r` | Refresh GitHub data |
| `Esc` | Close popup or quit |

## Configuration

Planned config file:

`~/.nexus/config.toml`

Example:

```toml
[github]
token = "ghp_xxxx"
auto_sync = true
sync_interval_minutes = 5

[appearance]
theme = "digital_noir"
color_scheme = "high_contrast"

[ai_agents]
copilot_enabled = true
claude_enabled = true
aider_enabled = true
claude_binary = "claude"

[worktrees]
root_dir = "/path/to/repo"
auto_link_prs = true
```

## Local data

Planned SQLite tables:

- `worktrees`
- `github_prs`
- `github_issues`
- `agent_history`
- `context_snapshots`

## Roadmap

### Phase 1

- Project setup
- Core Bubble Tea app structure
- Git worktree integration
- Worktree list UI
- Create/delete modals
- Theme styling

### Phase 2

- GitHub client setup
- TOML config loading
- SQLite schema and migrations
- PR and issue sync
- Background refresh
- Worktree-to-PR linking

### Phase 3

- Copilot launcher
- Claude launcher
- Aider launcher
- Context builder
- Agent history

### Phase 4

- Error handling
- Help system
- Settings screen
- Performance tuning
- Tests and packaging

## Error handling goals

- Show cached GitHub data when the network fails
- Provide clear messages when an agent binary is missing
- Warn when Nexus is launched outside a Git repo
- Prevent unsafe git command execution

## Security goals

- Prefer `gh` auth over raw tokens
- Never log sensitive data
- Keep local config permissions tight
- Validate input before shelling out

## Status

This repository currently contains the project plan and documentation scaffolding. The implementation is still pending.

## Documentation

- [Project plan](docs/PLAN.md)
- [Runbook](docs/RUNBOOK.md)

## License

Add a license before publishing the project publicly.
