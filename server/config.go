package server

import "time"

type Config struct {
	AppPort    int
	ProxyPort  int
	IgnoreDirs []string
	Timeout    time.Duration
	Wait       bool
	Verbose    bool
}
