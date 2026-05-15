package data

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/m00nk0d3/nexus/internal/domain"
	"github.com/pelletier/go-toml/v2"
)

// DefaultConfigPath returns the default path to the Nexus config file.
func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".nexus", "config.toml")
}

// LoadConfig reads and parses the TOML config at path.
// If the file does not exist, it returns the default config.
func LoadConfig(path string) (*domain.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.DefaultConfig(), nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg domain.Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

// SaveConfig marshals cfg to TOML and writes it to path, creating parent directories as needed.
func SaveConfig(cfg *domain.Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
