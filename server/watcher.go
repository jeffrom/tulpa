package server

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MichaelTJones/walk"
)

type watcher struct {
	cfg     *Config
	lastRun time.Time
}

func newWatcher(cfg *Config) *watcher {
	return &watcher{
		cfg:     cfg,
		lastRun: time.Now(),
	}
}

func (w *watcher) scan() bool {
	log.Print("watcher: start")
	start := time.Now()

	modified := walk.Walk(".", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && w.shouldSkipDir(path) {
			return walk.SkipDir
		}

		if info.ModTime().After(w.lastRun) {
			log.Printf("Watcher: Found modified file: %v", path)
			return errors.New(path)
		}

		return nil
	})

	log.Printf("Watcher: Scan finished: %v", time.Since(start))
	return modified != nil
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
