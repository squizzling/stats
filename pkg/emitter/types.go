package emitter

import (
	"go.uber.org/zap"

	"github.com/squizzling/stats/pkg/statser"
)

type EmitterFactory func(logger *zap.Logger, statsPool statser.Pool) Emitter

type Emitter interface {
	Emit()
}
