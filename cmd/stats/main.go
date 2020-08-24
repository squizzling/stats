package main

import (
	"fmt"
	"os"
	"time"

	"github.com/alexcesaro/statsd"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	_ "github.com/squizzling/stats/internal/emitters/meminfo"
	_ "github.com/squizzling/stats/internal/emitters/procstat"
	_ "github.com/squizzling/stats/internal/emitters/sysfs"
	_ "github.com/squizzling/stats/internal/emitters/zfs"

	"github.com/squizzling/stats/internal/istats"
	"github.com/squizzling/stats/pkg/emitter"
	"github.com/squizzling/stats/pkg/sources"
)

func createLogger(verbose bool) *zap.Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stdout"}
	cfg.DisableStacktrace = true
	if verbose {
		cfg.Level.SetLevel(zapcore.DebugLevel)
	} else {
		cfg.Level.SetLevel(zapcore.InfoLevel)
	}
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

	if opts.List {
		for key, _ := range sources.Sources {
			fmt.Printf("- %s\n", key)
		}
		return
	}

	logger := createLogger(opts.Verbose)
	defer func() {
		_ = logger.Sync()
	}()
	statsPool := istats.NewPool(createStatsClient(logger, opts.Target, *opts.Host))

	var emitters []emitter.Emitter
	for key, factory := range sources.Sources {
		if opts.enableDisable != nil {
			_, ok := opts.selected[key]
			delete(opts.selected, key)
			if ok != *opts.enableDisable {
				continue
			}
		}
		logger.Info("enabled", zap.String("emitter", key))
		emitters = append(emitters, factory(logger, statsPool))
	}

	for key, _ := range opts.selected {
		logger.Warn("unrecognized emitter", zap.String("emitter", key))
	}

	for {
		time.Sleep(time.Second)
		logger.Info("emitting")
		for _, e := range emitters {
			e.Emit()
		}
	}
}
