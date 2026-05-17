# NEXUS — Git Worktree Orchestrator & AI Agent Hub

> Manage Git worktrees, track GitHub PRs and Issues, and launch AI coding agents — all from a single terminal interface.

![Go version](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-green)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)

<!-- Screenshot placeholder — run `vhs demo.tape` to regenerate -->
![Nexus TUI — 3-pane worktree dashboard](demo/nexus.gif)

---

## Why Nexus?

Modern software development means juggling multiple things at once: several features in flight, a handful of open PRs waiting for review, GitHub issues to reference, and at least one AI agent that swears it can fix everything. Keeping all of that in sync — without losing your mind or your terminal history — is genuinely painful.

**The problems Nexus solves:**

- **Context-switching hell.** Stashing changes, checking out branches, losing your editor state, repeating. Git worktrees are the solution, but their CLI is clunky and easy to forget. Nexus puts your entire worktree landscape on one screen and lets you jump between them instantly.

- **The five-app shuffle.** Terminal for git, browser for GitHub, another terminal for the agent, Slack for the PR link, repeat. Nexus collapses all of that into a single pane of glass: worktrees, PRs, issues, and agent launchers — no browser tab required.

- **AI agents without context.** Spinning up Claude Code or Copilot in the wrong directory (or forgetting which worktree maps to which feature) wastes time and produces wrong answers. Nexus launches agents *inside the correct worktree* automatically, so your AI always has the right repo context.

- **"Which branch had that issue again?"** Nexus links GitHub issues and PRs to their worktrees so you always know what's where. No more `git branch -a | grep vague-memory`.

In short: if you work on multiple features simultaneously and use AI coding tools, Nexus removes the glue work so you can focus on the actual code.

---

## Features

- **3-pane TUI** — worktree list, GitHub context panel, and detail view in one terminal window
- **Full worktree management** — create, delete, switch shell, lock/unlock, and prune worktrees without leaving the terminal
- **GitHub sync** — pull requests and issues fetched via the `gh` CLI and kept fresh in the background
- **AI agent launchers** — spawn Claude Code, GitHub Copilot, or Aider in the correct worktree directory with a single keypress
- **Theme cycling** — switch between Digital Noir, Matrix, and Light themes on the fly with `t`
- **In-app help** — press `f1` or `?` at any time for a searchable keybindings and troubleshooting reference
- **Local persistence** — config lives in `~/.nexus/config.toml`; metadata is cached in SQLite so Nexus starts fast

---

## Prerequisites

| Requirement | Notes |
|---|---|
| [Git](https://git-scm.com/) | Must be in `PATH` |
| [GitHub CLI (`gh`)](https://cli.github.com/) | Run `gh auth login` before first use |
| Go 1.25+ | Only needed if building from source |
| Claude Code | Optional — enable with `claude_enabled = true` |
| GitHub Copilot CLI | Optional — `gh extension install github/gh-copilot` |
| Aider | Optional — `pip install aider-chat` |

---

## Installation

### Using go install (recommended)

```bash
go install github.com/m00nk0d3/nexus/cmd/nexus@latest
```

### Build from source

```bash
git clone https://github.com/m00nk0d3/nexus
cd nexus
go build -ldflags "-X github.com/m00nk0d3/nexus/internal/version.Version=v0.1.0" \
    -o nexus ./cmd/nexus
```

Move the resulting `nexus` binary somewhere on your `PATH`.

---

## Quick Start

1. `cd` into any Git repository
2. Run `nexus`
3. Nexus opens in the **Worktrees** view — use `j`/`k` to navigate the list
4. Press `Ctrl+N` to create a new worktree (optionally linked to a GitHub issue)
5. Press `Enter` or `s` to open a shell inside the selected worktree
6. Press `a`, `c`, or `Space` to launch an AI agent in that worktree's context

---

## Keybindings

### Navigation

| Key | Action |
|---|---|
| `↑` / `↓` or `j` / `k` | Navigate within panel |
| `←` / `→` or `h` / `l` | Switch between panels / tabs |
| `Tab` | Cycle panel focus / tab |
| `Enter` | Open shell in worktree / Select |

### Worktree Operations

| Key | Action |
|---|---|
| `Ctrl+N` | Create new worktree |
| `Ctrl+D` | Delete selected worktree |
| `s` | Open shell in worktree |

### AI Agents

| Key | Action |
|---|---|
| `a` | Spawn Claude Code |
| `c` | Spawn GitHub Copilot |
| `Space` | Unified agent launcher (shows all agents and availability) |

### Views

| Key | Action |
|---|---|
| `w` / `W` | Worktrees view |
| `i` / `I` | Issues view |
| `p` / `P` | PRs view |
| `t` | Cycle themes (Digital Noir → Matrix → Light) |

### Global

| Key | Action |
|---|---|
| `f1` / `?` | Open help modal |
| `g` | Open selected item in GitHub |
| `Esc` / `Ctrl+C` | Quit |

---

## Configuration

Nexus reads its config from `~/.nexus/config.toml`. The file is created with defaults on first run. All fields are optional.

```toml
[github]
# Automatically sync PRs and Issues in the background.
auto_sync = true

# How often to refresh GitHub data (in minutes).
sync_interval_minutes = 5

[appearance]
# UI theme. Options: "digital-noir", "matrix", "light"
theme = "digital-noir"

[ai_agents]
# Enable or disable individual agent launchers.
copilot_enabled = true
claude_enabled  = true
aider_enabled   = false

# Override the binary name/path if it differs from the default.
claude_binary = "claude"
aider_binary  = "aider"

[worktrees]
# The branch used as the base when creating new worktrees.
base_branch = "main"

# Where new worktree directories are created, relative to the repo root.
worktree_root = "../worktrees"
```

---

## AI Agent Setup

### GitHub Copilot

```bash
# Install the Copilot CLI extension
gh extension install github/gh-copilot

# Authenticate (if not already done)
gh auth login
```

Enable in config (default: `true`). Trigger in Nexus: **`c`**.

### Claude Code

Install via the official guide: <https://docs.anthropic.com/en/docs/claude-code>

The binary must be named `claude` (or set `claude_binary` to its full path). Enable in config (default: `true`). Trigger in Nexus: **`a`**.

### Aider

```bash
pip install aider-chat
```

Set `aider_enabled = true` in `~/.nexus/config.toml`. Trigger in Nexus: **`Space`** → select Aider in the unified launcher.

---

## Troubleshooting

| Symptom | Fix |
|---|---|
| Empty issues / PRs list or "gh: not logged in" | Run `gh auth login` and follow the prompts |
| "binary not found" when spawning an agent | Install the missing tool or set `*_enabled = false` in config |
| No worktrees visible — "git not in PATH" | Ensure Git is installed: `git --version` |
| Deleted worktrees still appear | Run `git worktree prune` in your repo, then press `r` in Nexus |
| Warning banner on startup — config not loading | Check `~/.nexus/config.toml` for TOML syntax errors; delete to restore defaults |

For more detail, press **`f1`** inside Nexus and open the **Troubleshooting** tab.

---

## Contributing

Contributions are welcome! Here's how to get set up and ship something.

### Developer setup

```bash
# 1. Clone the repo
git clone https://github.com/m00nk0d3/nexus
cd nexus

# 2. Verify Go 1.25+ is installed
go version

# 3. Install the GitHub CLI and authenticate
gh auth login

# 4. Build the project
go build ./...

# 5. Run Nexus locally (from inside any Git repo)
go run ./cmd/nexus
```

For a full developer reference including release procedures and configuration details, see [docs/RUNBOOK.md](docs/RUNBOOK.md).

### Workflow

1. **Open an issue first** — before starting anything non-trivial, open an issue so we can align on the approach and avoid wasted effort.
2. **Branch naming** — use `feat/issue-<number>-<short-description>` (e.g. `feat/issue-42-worktree-lock-ui`).
3. **Run tests** before pushing:
   ```bash
   go test ./...
   ```
4. **Commit messages** follow [Conventional Commits](https://www.conventionalcommits.org/) and must include a body explaining *why* the change was made (see `.copilot/` for the project commit guidelines).
5. **Open a PR** against `main` with a clear description: what changed, why, and how it was tested.

---

## License

Nexus is released under the [MIT License](LICENSE).
