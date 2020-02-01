package server

import (
	"log"
	"net"
	"time"
)

type Server struct {
	cfg     *Config
	proxy   *proxy
	runner  *runner
	watcher *watcher
}

func New(cfg *Config, args []string) *Server {
	return &Server{
		cfg:     cfg,
		proxy:   newProxy(cfg),
		runner:  newRunner(cfg, args),
		watcher: newWatcher(cfg),
	}
}

func (s *Server) Addr() net.Addr { return s.proxy.ln.Addr() }

func (s *Server) GoStart() chan error {
	errC := make(chan error)
	stop := make(chan error, 1)
	ready := make(chan error)

	go func() {
		if err := s.start(stop, ready); err != nil {
			errC <- err
		}
		errC <- nil
	}()

	if err := <-ready; err != nil {
		panic(err)
	}
	return errC
}

func (s *Server) Start() error {
	stop := make(chan error, 1)

	return s.start(stop, nil)
}

func (s *Server) start(stop chan error, ready chan error) error {
	go func() {
		if err := s.proxy.start(ready); err != nil {
			stop <- err
		}
	}()

	if err := s.runner.run(); err != nil {
		s.proxy.setError(err)
	}

	for {
		select {
		case <-s.proxy.requests:
			modified := s.watcher.scan()
			if modified {
				log.Printf("server: fs modified, rerunning...")

				if err := s.runner.run(); err != nil {
					s.proxy.setError(err)
					s.proxy.unpause <- struct{}{}
					continue
				}

				s.proxy.clearError()
			}

			s.watcher.lastRun = time.Now()
			s.proxy.unpause <- struct{}{}

		case err := <-s.runner.errors:
			log.Print("runner: error")
			s.proxy.setError(err)
		case err := <-stop:
			s.Stop()
			return err
		}
	}
}

func (s *Server) Stop() {
	s.runner.kill()
}
