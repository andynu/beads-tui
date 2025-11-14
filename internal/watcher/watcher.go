package watcher

import (
	"fmt"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher monitors a file for changes and triggers a callback
type Watcher struct {
	watcher       *fsnotify.Watcher
	path          string
	debounceDelay time.Duration
	onChange      func()
	stopCh        chan struct{}
}

// New creates a new file watcher
func New(path string, debounceDelay time.Duration, onChange func()) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	w := &Watcher{
		watcher:       fsWatcher,
		path:          path,
		debounceDelay: debounceDelay,
		onChange:      onChange,
		stopCh:        make(chan struct{}),
	}

	return w, nil
}

// Start begins watching the file (and SQLite WAL file if applicable)
func (w *Watcher) Start() error {
	if err := w.watcher.Add(w.path); err != nil {
		return fmt.Errorf("failed to watch file: %w", err)
	}

	// For SQLite databases, also watch the WAL file where changes are written
	walPath := w.path + "-wal"
	if err := w.watcher.Add(walPath); err != nil {
		// WAL file might not exist yet, which is fine
		// We'll still catch changes to the main DB file
	}

	go w.watchLoop()
	return nil
}

// Stop stops watching the file
func (w *Watcher) Stop() error {
	close(w.stopCh)
	return w.watcher.Close()
}

// watchLoop runs the main watch loop with debouncing
func (w *Watcher) watchLoop() {
	var debounceTimer *time.Timer

	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Only respond to write and create events
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				// Debounce: reset timer if it's already running
				if debounceTimer != nil {
					debounceTimer.Stop()
				}

				debounceTimer = time.AfterFunc(w.debounceDelay, func() {
					w.onChange()
				})
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			// In a real app, you might want to log this or handle it
			_ = err

		case <-w.stopCh:
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			return
		}
	}
}
