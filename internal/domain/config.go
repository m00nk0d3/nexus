package domain

type Config struct {
	GitHub     GitHubConfig     `toml:"github"`
	Appearance AppearanceConfig `toml:"appearance"`
	AIAgents   AIAgentsConfig   `toml:"ai_agents"`
	Worktrees  WorktreesConfig  `toml:"worktrees"`
}

type GitHubConfig struct {
	AutoSync            bool `toml:"auto_sync"`
	SyncIntervalMinutes int  `toml:"sync_interval_minutes"`
}

type AppearanceConfig struct {
	Theme string `toml:"theme"`
}

type AIAgentsConfig struct {
	CopilotEnabled bool   `toml:"copilot_enabled"`
	ClaudeEnabled  bool   `toml:"claude_enabled"`
	AiderEnabled   bool   `toml:"aider_enabled"`
	ClaudeBinary   string `toml:"claude_binary"`
}

type WorktreesConfig struct {
	BaseBranch   string `toml:"base_branch"`
	WorktreeRoot string `toml:"worktree_root"`
}

func DefaultConfig() *Config {
	return &Config{
		Appearance: AppearanceConfig{Theme: "digital-noir"},
		Worktrees:  WorktreesConfig{BaseBranch: "main", WorktreeRoot: "../worktrees"},
		GitHub:     GitHubConfig{AutoSync: true, SyncIntervalMinutes: 5},
	}
}
