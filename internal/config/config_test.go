package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Theme != "gruvbox-dark" {
		t.Errorf("expected default theme to be 'gruvbox-dark', got %q", cfg.Theme)
	}
}

func TestLoadSaveConfig(t *testing.T) {
	// Use a temporary directory for testing
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Load should return default config when file doesn't exist
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.Theme != "gruvbox-dark" {
		t.Errorf("expected default theme, got %q", cfg.Theme)
	}

	// Modify and save
	cfg.Theme = "nord"
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file was created
	path, _ := ConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Load again and verify
	cfg2, err := Load()
	if err != nil {
		t.Fatalf("Load() after save failed: %v", err)
	}
	if cfg2.Theme != "nord" {
		t.Errorf("expected saved theme 'nord', got %q", cfg2.Theme)
	}
}

func TestConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() failed: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, ".beads-tui", "config.json")
	if path != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, path)
	}

	// Verify directory was created
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("config directory was not created")
	}
}
