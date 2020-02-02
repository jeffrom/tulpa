package server

import (
	"sync"
	"sync/atomic"
	"time"
)

// forked from github.com/bep/debounce

// newDebouncer returns a debounced function that takes another functions as
// its argument.  This function will be called when the debounced function
// stops being called for the given duration.  The debounced function can be
// invoked with different functions, if needed, the last one will win.
func newDebouncer(cfg *Config) func(f func()) {
	d := &debouncer{cfg: cfg}

	return func(f func()) {
		d.add(f)
	}
}

type debouncer struct {
	cfg       *Config
	mu        sync.Mutex
	timer     *time.Timer
	longTimer *time.Timer
	called    int32
}

func (d *debouncer) add(f func()) {
	if d.cfg.Debounce == 0 {
		f()
		return
	}
	if atomic.CompareAndSwapInt32(&d.called, 0, 1) {
		f()
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.cfg.Debounce, func() {
		f()
		atomic.StoreInt32(&d.called, 0)

		d.mu.Lock()
		defer d.mu.Unlock()
		if d.longTimer != nil {
			d.longTimer.Stop()
		}
		d.longTimer = nil
	})

	if d.longTimer != nil || d.cfg.DebouncePoll <= 0 {
		return
	}

	d.longTimer = time.AfterFunc(d.cfg.DebouncePoll, func() {
		f()

		d.mu.Lock()
		defer d.mu.Unlock()
		if d.longTimer != nil {
			d.longTimer.Stop()
		}
		d.longTimer = nil
	})
}
