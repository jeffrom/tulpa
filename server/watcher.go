package server

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/MichaelTJones/walk"
)

type watcher struct {
	cfg     *Config
	lastRun time.Time
	mu      sync.Mutex
}

func newWatcher(cfg *Config) *watcher {
	return &watcher{
		cfg:     cfg,
		lastRun: time.Now(),
	}
}

func (w *watcher) scan() bool {
	w.cfg.Debug("start scan")
	start := time.Now()

	modified := walk.Walk(".", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && w.shouldSkipDir(path) {
			return walk.SkipDir
		}

		if info.ModTime().After(w.getLastRun()) {
			w.cfg.Debugf("found modified file: %v", path)
			return errors.New(path)
		}

		return nil
	})

	w.cfg.Printf("scan done in %v", time.Since(start))
	return modified != nil
}

func (w *watcher) getLastRun() time.Time {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.lastRun
}

func (w *watcher) setLastRun(t time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.lastRun = t
}

// Checks to see if this directory should be watched. Don't want to watch
// hidden directories (like .git) or ignored directories.
func (w *watcher) shouldSkipDir(path string) bool {
	if len(path) > 1 && strings.HasPrefix(filepath.Base(path), ".") {
		return true
	}

	for _, dir := range w.cfg.IgnoreDirs {
		if dir == path {
			return true
		}
	}

	return false
}
