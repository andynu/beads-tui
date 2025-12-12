package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds persistent user configuration
type Config struct {
	Theme string `json:"theme"` // Current theme name
}

// CollapseState holds the collapse state for tree view nodes
// Keyed by issue ID, value is true if collapsed
type CollapseState struct {
	CollapsedNodes map[string]bool `json:"collapsed_nodes"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Theme: "gruvbox-dark",
	}
}

// ConfigPath returns the path to the config file
func ConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".beads-tui")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, "config.json"), nil
}

// Load reads the config file from disk, or returns defaults if not found
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return default config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Save writes the config to disk
func Save(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// CollapseStatePath returns the path for collapse state file for a given beads directory
// Uses a hash of the beads path to create a unique filename per project
func CollapseStatePath(beadsDir string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".beads-tui")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create a short hash of the beads directory path for uniqueness
	hash := sha256.Sum256([]byte(beadsDir))
	shortHash := hex.EncodeToString(hash[:])[:8]

	return filepath.Join(configDir, fmt.Sprintf("collapse-%s.json", shortHash)), nil
}

// LoadCollapseState reads the collapse state from disk for a given beads directory
func LoadCollapseState(beadsDir string) (*CollapseState, error) {
	path, err := CollapseStatePath(beadsDir)
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty state
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &CollapseState{CollapsedNodes: make(map[string]bool)}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read collapse state file: %w", err)
	}

	var state CollapseState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse collapse state file: %w", err)
	}

	if state.CollapsedNodes == nil {
		state.CollapsedNodes = make(map[string]bool)
	}

	return &state, nil
}

// SaveCollapseState writes the collapse state to disk for a given beads directory
func SaveCollapseState(beadsDir string, state *CollapseState) error {
	path, err := CollapseStatePath(beadsDir)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize collapse state: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write collapse state file: %w", err)
	}

	return nil
}
