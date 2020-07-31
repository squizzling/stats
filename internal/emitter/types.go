package emitter

import (
	"go.uber.org/zap"

	"github.com/squizzling/stats/internal/stats"
)

type EmitterFactory func(logger *zap.Logger, statsPool *stats.Pool) Emitter

type Emitter interface {
	Emit()
}
