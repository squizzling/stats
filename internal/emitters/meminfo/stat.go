package meminfo

import (
	"go.uber.org/zap"

	"github.com/squizzling/stats/pkg/emitter"
	"github.com/squizzling/stats/pkg/sources"
	"github.com/squizzling/stats/pkg/statser"
)

type MemInfoEmitter struct {
	logger         *zap.Logger
	statsClient    statser.Statser
	trackedMetrics map[string]string
}

func NewEmitter(logger *zap.Logger, statsPool statser.Pool) emitter.Emitter {
	return &MemInfoEmitter{
		logger:      logger,
		statsClient: statsPool.Get(),
		trackedMetrics: map[string]string{
			"mem_total":     "MemTotal",
			"mem_free":      "MemFree",
			"mem_available": "MemAvailable",
			"buffers":       "Buffers",
			"cached":        "Cached",
			"slab":          "Slab",
		},
	}
}

func (mse *MemInfoEmitter) tryEmit(ms *MemInfo, metricSuffix, memStatName string) {
	metricName := "procmeminfo." + metricSuffix
	if value, ok := ms.Values[memStatName]; ok {
		mse.statsClient.Gauge(metricName, value)
	} else {
		mse.logger.Warn("not found", zap.String("key", memStatName))
	}
}

func (mse *MemInfoEmitter) Emit() {
	ms := LoadMemInfo(mse.logger)
	for metricSuffix, memStatName := range mse.trackedMetrics {
		mse.tryEmit(ms, metricSuffix, memStatName)
	}
}

func init() {
	sources.Sources["meminfo"] = NewEmitter
}
