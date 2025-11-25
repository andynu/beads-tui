package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcher(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(testFile, []byte("initial content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Track if onChange was called
	called := make(chan bool, 10)
	onChange := func() {
		called <- true
	}

	// Create watcher with short debounce
	w, err := New(testFile, 50*time.Millisecond, onChange)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	// Start watching
	if err := w.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}
	defer func() { _ = w.Stop() }()

	// Wait a bit for watcher to be ready
	time.Sleep(100 * time.Millisecond)

	// Modify the file
	if err := os.WriteFile(testFile, []byte("modified content"), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Wait for debounce + processing
	select {
	case <-called:
		// Success - onChange was called
	case <-time.After(500 * time.Millisecond):
		t.Fatal("onChange was not called within timeout")
	}
}

func TestWatcherDebounce(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(testFile, []byte("initial"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	callCount := 0
	onChange := func() {
		callCount++
	}

	w, err := New(testFile, 100*time.Millisecond, onChange)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	if err := w.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}
	defer func() { _ = w.Stop() }()

	time.Sleep(100 * time.Millisecond)

	// Write multiple times rapidly
	for i := 0; i < 5; i++ {
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to modify test file: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for debounce to settle
	time.Sleep(200 * time.Millisecond)

	// Should only be called once due to debouncing
	if callCount != 1 {
		t.Errorf("Expected 1 call due to debouncing, got %d", callCount)
	}
}

func TestWatcherStop(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(testFile, []byte("initial"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	called := false
	onChange := func() {
		called = true
	}

	w, err := New(testFile, 50*time.Millisecond, onChange)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	if err := w.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Stop immediately
	if err := w.Stop(); err != nil {
		t.Fatalf("Failed to stop watcher: %v", err)
	}

	// Modify after stop - should not trigger
	if err := os.WriteFile(testFile, []byte("modified"), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	if called {
		t.Error("onChange was called after watcher was stopped")
	}
}
