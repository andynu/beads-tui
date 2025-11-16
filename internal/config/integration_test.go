package config

import (
	"os"
	"testing"
)

// TestConfigPersistenceWorkflow simulates the full user workflow
func TestConfigPersistenceWorkflow(t *testing.T) {
	// Use a temporary directory for testing
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Simulate first run: no config exists
	cfg1, err := Load()
	if err != nil {
		t.Fatalf("First Load() failed: %v", err)
	}
	if cfg1.Theme != "gruvbox-dark" {
		t.Errorf("First run: expected default theme 'gruvbox-dark', got %q", cfg1.Theme)
	}

	// Simulate user changing theme to 'nord' and saving
	cfg1.Theme = "nord"
	if err := Save(cfg1); err != nil {
		t.Fatalf("Save() after theme change failed: %v", err)
	}

	// Simulate second run: config should persist
	cfg2, err := Load()
	if err != nil {
		t.Fatalf("Second Load() failed: %v", err)
	}
	if cfg2.Theme != "nord" {
		t.Errorf("Second run: expected persisted theme 'nord', got %q", cfg2.Theme)
	}

	// Simulate user changing theme to 'dracula'
	cfg2.Theme = "dracula"
	if err := Save(cfg2); err != nil {
		t.Fatalf("Save() after second theme change failed: %v", err)
	}

	// Simulate third run: should load 'dracula'
	cfg3, err := Load()
	if err != nil {
		t.Fatalf("Third Load() failed: %v", err)
	}
	if cfg3.Theme != "dracula" {
		t.Errorf("Third run: expected persisted theme 'dracula', got %q", cfg3.Theme)
	}
}

// TestConfigResilience tests that config handles edge cases gracefully
func TestConfigResilience(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Test: Corrupted config file should fall back to defaults
	path, _ := ConfigPath()
	if err := os.WriteFile(path, []byte("invalid json {{{"), 0644); err != nil {
		t.Fatalf("Failed to create corrupted config: %v", err)
	}

	// Load should fail gracefully and return error
	_, err := Load()
	if err == nil {
		t.Error("Expected error when loading corrupted config, got nil")
	}
}
