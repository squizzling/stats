package procstat

import (
	"fmt"
	"strconv"

	"go.uber.org/zap"

	"github.com/squizzling/stats/pkg/emitter"
	"github.com/squizzling/stats/pkg/sources"
	"github.com/squizzling/stats/pkg/statser"
)

type ProcStatEmitter struct {
	logger    *zap.Logger
	statsPool statser.Pool
}

func NewEmitter(logger *zap.Logger, statsPools statser.Pool) emitter.Emitter {
	return &ProcStatEmitter{
		logger:    logger,
		statsPool: statsPools,
	}
}

func emitProcStatCpu(c statser.Statser, s string, cpu *ProcStatCpu) {
	c.Gauge(fmt.Sprintf("procstat.cpu.%s.user", s), cpu.User)
	c.Gauge(fmt.Sprintf("procstat.cpu.%s.nice", s), cpu.Nice)
	c.Gauge(fmt.Sprintf("procstat.cpu.%s.system", s), cpu.System)
	c.Gauge(fmt.Sprintf("procstat.cpu.%s.idle", s), cpu.Idle)
	c.Gauge(fmt.Sprintf("procstat.cpu.%s.iowait", s), cpu.IoWait)
	c.Gauge(fmt.Sprintf("procstat.cpu.%s.irq", s), cpu.Irq)
	c.Gauge(fmt.Sprintf("procstat.cpu.%s.softirq", s), cpu.SoftIrq)
	c.Gauge(fmt.Sprintf("procstat.cpu.%s.steal", s), cpu.Steal)
	c.Gauge(fmt.Sprintf("procstat.cpu.%s.guest", s), cpu.Guest)
	c.Gauge(fmt.Sprintf("procstat.cpu.%s.guestnice", s), cpu.GuestNice)

	active := 0 + // because gofmt is awesome
		cpu.User +
		cpu.Nice +
		cpu.System +
		cpu.IoWait +
		cpu.Irq +
		cpu.SoftIrq +
		cpu.Steal +
		cpu.Guest +
		cpu.GuestNice
	total := active + cpu.Idle
	c.Gauge(fmt.Sprintf("procstat.cpu.%s.active", s), active)
	c.Gauge(fmt.Sprintf("procstat.cpu.%s.total", s), total)
}

func (pse *ProcStatEmitter) Emit() {
	ps := LoadProcStat(pse.logger)
	if ps.CPUTotal != nil {
		emitProcStatCpu(pse.statsPool.Get(), "total", ps.CPUTotal)
	}
	for idx, perCPUStats := range ps.CPUs {
		emitProcStatCpu(pse.statsPool.Get("cpu", strconv.Itoa(idx)), "per", perCPUStats)
	}
}

func init() {
	sources.Sources["procstat"] = NewEmitter
}
