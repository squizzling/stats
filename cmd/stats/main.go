package main

import (
	"fmt"
	"os"
	"time"

	"github.com/alexcesaro/statsd"
	"go.uber.org/zap"

	"github.com/squizzling/stats/internal/emitter"
	"github.com/squizzling/stats/internal/emitter/meminfo"
	"github.com/squizzling/stats/internal/emitter/procstat"
	"github.com/squizzling/stats/internal/emitter/zfs"
	"github.com/squizzling/stats/internal/stats"
)

func createLogger() *zap.Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stdout"}
	logger, err := cfg.Build()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error creating logger: %v\n", err)
		os.Exit(1)
	}
	return logger
}

func createStatsClient(logger *zap.Logger, target, host string) *statsd.Client {
	c, err := statsd.New(
		statsd.Address(target),
		statsd.Network("udp4"),
		statsd.FlushPeriod(1*time.Second),
		statsd.TagsFormat(statsd.Datadog),
		statsd.Tags("host", host),
	)
	if err != nil {
		logger.Error(
			"failed to create statsd client",
			zap.Error(err),
		)
		_ = logger.Sync()
		os.Exit(1)
	}
	return c
}

func main() {
	opts := parseArgs(os.Args[1:])
	logger := createLogger()
	defer func() {
		_ = logger.Sync()
	}()
	statsPool := stats.NewPool(createStatsClient(logger, opts.Target, *opts.Host))

	emitterFactories := []emitter.EmitterFactory{
		zfs.NewEmitter,
		procstat.NewEmitter,
		meminfo.NewEmitter,
	}

	var emitters []emitter.Emitter
	for _, factory := range emitterFactories {
		emitters = append(emitters, factory(logger, statsPool))
	}

	for {
		time.Sleep(time.Second)
		logger.Info("emitting")
		for _, e := range emitters {
			e.Emit()
		}
	}
}
