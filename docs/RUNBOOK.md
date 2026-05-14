# Nexus Runbook

This runbook covers day-to-day operation of Nexus and the expected procedures around setup, configuration, troubleshooting, and maintenance.

## Purpose

Nexus is a terminal UI for managing Git worktrees, syncing GitHub data, and launching AI coding agents with the right repository context.

## Current status

The repository currently contains the project plan and documentation scaffolding. The operational procedures below describe the intended runtime behavior once the app is implemented.

## Owners

- Primary owner: project maintainer
- Secondary owner: anyone responsible for GitHub auth, local config, or release packaging

## Prerequisites

- Git
- GitHub CLI (`gh`) authenticated to the target account
- Go toolchain
- A git repository with worktrees enabled
- Optional: Claude Code and Aider binaries if those launchers are enabled

## Setup

1. Clone the repository.
2. Ensure `gh auth status` succeeds.
3. Create the Nexus config directory: `~/.nexus/`
4. Add `~/.nexus/config.toml`
5. Start Nexus from the root of a git repository

## Configuration

Primary config file:

`~/.nexus/config.toml`

Key settings:

| Setting | Purpose |
| --- | --- |
| `github.token` | Optional PAT if `gh` auth is not used |
| `github.auto_sync` | Enables background sync |
| `github.sync_interval_minutes` | Refresh cadence |
| `appearance.theme` | UI theme selection |
| `ai_agents.*_enabled` | Toggles individual launchers |
| `worktrees.root_dir` | Base repo path |
| `worktrees.auto_link_prs` | Auto-associate PRs to branches |

## Startup procedure

1. Confirm the repo is a valid git repository.
2. Confirm the current worktree list is readable.
3. Load local config and cached data.
4. Start the UI.
5. Start background GitHub sync if enabled.

## Normal operations

### Worktree management

- Create a worktree from the dashboard.
- Switch to a worktree when you need repository context there.
- Lock worktrees before long-lived work.
- Prune stale worktrees after branch deletion or directory removal.

### GitHub sync

- Refresh manually with `r` when current data looks stale.
- Use auto-sync for steady-state updates.
- Prefer cached data during temporary API failures.

### Agent launchers

- Use Copilot for quick guidance and suggestions.
- Use Claude for broader reasoning and multi-step changes.
- Use Aider when you want file-scoped editing assistance.

## Troubleshooting

### Nexus will not start

Check:

- you are inside a git repository
- `gh auth status` succeeds
- config file exists and is readable

### GitHub data is stale

Check:

- background sync is enabled
- network access is available
- the auth token still works

### Agent launcher fails

Check:

- the binary exists in `PATH`
- the launcher is enabled in config
- the worktree path is valid

### Worktree operations fail

Check:

- the repository is clean enough for the intended operation
- the branch name is valid
- the target worktree is not locked

## Incident response

### GitHub API outage

1. Fall back to cached metadata.
2. Stop relying on live refreshes.
3. Retry after connectivity recovers.

### Corrupted local config

1. Back up `~/.nexus/`.
2. Remove or repair the broken TOML file.
3. Restart with a minimal config.

### Missing agent binary

1. Disable the launcher in config.
2. Install the missing tool.
3. Re-enable the launcher after verification.

## Backup and restore

Back up:

- `~/.nexus/config.toml`
- SQLite cache or local database files
- logs under `~/.nexus/logs/`

Restore by copying the files back into the same paths and restarting Nexus.

## Logging

Operational logs should live under:

`~/.nexus/logs/nexus.log`

Use logs for:

- sync failures
- launch failures
- git command errors
- config validation errors

## Maintenance tasks

- Review config defaults when adding new features
- Update dependency notes after architecture changes
- Keep keybindings in sync with the UI
- Document new error states and recovery steps

## Release checklist

- README updated
- runbook updated
- keybindings documented
- config defaults documented
- edge cases documented
- license present

## Open questions

- Should local persistence use a single SQLite file or separate caches?
- Should Nexus prefer `gh` auth only, or allow PAT fallback by default?
- Should the app resume to the original shell context after agent launch or keep a visible handoff screen?
