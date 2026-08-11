package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"golang-clean-architecture/internal/bootstrap"
	"golang-clean-architecture/internal/buildinfo"
	"golang-clean-architecture/internal/config"
	"golang-clean-architecture/internal/delivery/messaging"
	"golang-clean-architecture/internal/logging"
)

func main() {
	var log *logging.Logger
	defer bootstrap.RecoverMain("worker", &log)

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
	).Component("worker")

	ctx, cancelRuntime := context.WithCancel(context.Background())
	defer cancelRuntime()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer signal.Stop(signals)
	consumers := []struct {
		topic string
		new   func(*logging.Logger) messaging.ConsumerHandler
	}{
		{topic: "contacts", new: func(log *logging.Logger) messaging.ConsumerHandler { return messaging.NewContactConsumer(log).Consume }},
		{topic: "addresses", new: func(log *logging.Logger) messaging.ConsumerHandler { return messaging.NewAddressConsumer(log).Consume }},
	}

	var wg sync.WaitGroup
	var bootstrapErr error
	for _, item := range consumers {
		consumer, err := config.OpenKafkaConsumerGroup(cfg.Kafka)
		if err != nil {
			log.Errorw("worker_bootstrap_failed", "topic", item.topic, "error", err)
			bootstrapErr = err
			cancelRuntime()
			break
		}
		topic, handler := item.topic, item.new(log.Component("kafka_consumer").With("topic", item.topic))
		wg.Go(func() { messaging.ConsumeTopic(ctx, consumer, topic, log, handler) })
	}
	if bootstrapErr == nil {
		received := <-signals
		log.Infow("shutdown_signal_received", "signal", received.String())
		cancelRuntime()
	}
	log.Infow("worker_shutdown_started")
	wg.Wait()
	log.Infow("worker_stopped")
	if bootstrapErr != nil {
		_ = log.Sync()
		os.Exit(1)
	}
}
