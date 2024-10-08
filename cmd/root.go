// Package cmd contains the command-line configuration of tulpa.
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jeffrom/tulpa/server"
	"github.com/spf13/cobra"
)

var version = "none"

func newRootCmd() *cobra.Command {
	cfg := &server.Config{}
	rootCmd := &cobra.Command{
		Use:   "tulpa",
		Short: "Development proxy with reload-after-change semantics",
		Long:  `tulpa is a command line utility for live reloading applications.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return start(cfg, args)
		},
		Version: version,
	}
	rootCmd.SetVersionTemplate("{{.Version}}\n")

	flags := rootCmd.Flags()
	flags.IntVarP(&cfg.AppPort, "app-port", "a", 3000, "application port")
	flags.IntVarP(&cfg.ProxyPort, "proxy-port", "p", 4000, "proxy port")
	flags.DurationVar(&cfg.Timeout, "timeout", 10*time.Second, "request timeout")
	flags.DurationVar(&cfg.Debounce, "debounce", 200*time.Millisecond, "file watch debounce interval")
	flags.DurationVar(&cfg.DebouncePoll, "debounce-poll", 1*time.Second, "poll interval while debounce is saturated")
	flags.DurationVar(&cfg.Latency, "latency", 0, "Duration to wait to respond to requests")
	flags.DurationVar(&cfg.LatencyJitter, "latency-jitter", 2*time.Second, "introduce randomness to latency duration")
	// TODO ignore pattern is better
	flags.StringArrayVarP(&cfg.IgnoreDirs, "ignore", "x", []string{"node_modules", "log", "tmp", "vendor", ".make"}, "directories to ignore")
	flags.BoolVarP(&cfg.Wait, "wait", "w", false, "wait for command to finish before serving request")
	flags.BoolVarP(&cfg.Verbose, "verbose", "v", false, "print extra debugging info")

	return rootCmd
}

func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func start(cfg *server.Config, args []string) error {
	cfg.Initialize()
	stop := make(chan os.Signal, 1)
	signal.Notify(
		stop,
		os.Interrupt,
		// syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	srv := server.New(cfg, args)

	go func() {
		if err := srv.Start(); err != nil {
			srv.Stop()
			log.Fatal(err)
		}
	}()

	<-stop
	srv.Stop()
	return nil
}
