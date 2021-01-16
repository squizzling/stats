package emitter

import (
	"go.uber.org/zap"

	"github.com/squizzling/stats/pkg/statser"
)

type EmitterFactory func(logger *zap.Logger, statsPool statser.Pool, opts OptProvider) Emitter

type Emitter interface {
	Emit()
}

type OptProvider interface {
	Get(name string) interface{}
}
