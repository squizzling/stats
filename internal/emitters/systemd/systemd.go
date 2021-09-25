package systemd

import (
	"github.com/godbus/dbus"
	"go.uber.org/zap"

	"github.com/squizzling/stats/pkg/emitter"
	"github.com/squizzling/stats/pkg/sources"
	"github.com/squizzling/stats/pkg/statser"
)

type SystemdEmitter struct {
	logger      *zap.Logger
	statsClient statser.Statser
	obj         dbus.BusObject
}

const (
	destSystemd     = "org.freedesktop.systemd1"
	pathSystemd     = "/org/freedesktop/systemd1"
	propFailedUnits = "org.freedesktop.systemd1.Manager.NFailedUnits"
)

func NewEmitter(logger *zap.Logger, statsPool statser.Pool, opt emitter.OptProvider) emitter.Emitter {
	b, err := dbus.SystemBus()
	if err != nil {
		logger.Error("failed to connect to system bus", zap.Error(err))
		return nil
	}

	sde := &SystemdEmitter{
		logger:      logger,
		statsClient: statsPool.Host(),
		obj:         b.Object(destSystemd, pathSystemd),
	}

	v := sde.failedUnits()
	if v < 0 {
		logger.Error("failed to read from systemd", zap.Int64("code", v))
		return nil
	}

	return sde
}

func (sde *SystemdEmitter) failedUnits() int64 {
	v, err := sde.obj.GetProperty(propFailedUnits)
	if err != nil {
		sde.logger.Error("failed to read NFailedUnits", zap.Error(err))
		return -1
	}
	if u, ok := v.Value().(uint32); ok {
		return int64(u)
	}
	return -2
}

func (sde *SystemdEmitter) Emit() {
	sde.statsClient.Gauge("systemd.failed_units", sde.failedUnits())
}

func init() {
	sources.Sources["systemd"] = NewEmitter
}
