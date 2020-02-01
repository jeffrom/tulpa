package server

import "time"

type Config struct {
	AppPort    int
	ProxyPort  int
	IgnoreDirs []string
	Timeout    time.Duration
	Debounce   time.Duration
	Wait       bool
	Verbose    bool
}
