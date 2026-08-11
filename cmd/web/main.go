package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang-clean-architecture/internal/bootstrap"
	"golang-clean-architecture/internal/buildinfo"
	"golang-clean-architecture/internal/config"
	"golang-clean-architecture/internal/logging"
)

func main() {
	var log *logging.Logger
	defer bootstrap.RecoverMain("web", &log)

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
	}
	log, err = logging.New(cfg.Log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger error: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = log.Sync() }()
	log = log.With(
		"service_name", cfg.App.Name,
		"app_env", cfg.App.Env,
		"app_version", buildinfo.Version,
		"build_commit", buildinfo.Commit,
		"build_time", buildinfo.BuildTime,
	)

	ctx, cancelRuntime := context.WithCancel(context.Background())
	defer cancelRuntime()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer signal.Stop(signals)
	app, err := bootstrap.NewApplication(ctx, cfg, log)
	if err != nil {
		log.Errorw("application_bootstrap_failed", "error", err)
		_ = log.Sync()
		os.Exit(1)
	}

	serverErr := make(chan error, 1)
	go func() { serverErr <- app.Start() }()

	exitCode := 0
	serverExited := false
	select {
	case received := <-signals:
		log.Infow("shutdown_signal_received", "signal", received.String())
		cancelRuntime()
	case err := <-serverErr:
		serverExited = true
		if err != nil {
			log.Errorw("http_server_stopped_unexpectedly", "error", err)
			exitCode = 1
		}
		cancelRuntime()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := app.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
		log.Errorw("application_shutdown_failed", "error", err)
		_ = log.Sync()
		os.Exit(1)
	}
	if !serverExited {
		select {
		case err := <-serverErr:
			if err != nil {
				log.Errorw("http_server_shutdown_failed", "error", err)
				exitCode = 1
			}
		case <-shutdownCtx.Done():
			log.Errorw("http_server_shutdown_timed_out", "error", shutdownCtx.Err())
			exitCode = 1
		}
	}
	log.Infow("application_stopped", "shutdown_timeout", cfg.HTTP.ShutdownTimeout.Round(time.Millisecond))
	if exitCode != 0 {
		_ = log.Sync()
		os.Exit(exitCode)
	}
}
