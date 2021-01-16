package procnetdev

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/squizzling/stats/internal/glob"
	"github.com/squizzling/stats/pkg/emitter"
	"github.com/squizzling/stats/pkg/sources"
	"github.com/squizzling/stats/pkg/statser"
)

type matchers struct {
	include []glob.Matcher
	exclude []glob.Matcher
}

func (m *matchers) Match(value string) bool {
	for _, m := range m.exclude {
		if m.Match(value) {
			return false
		}
	}

	for _, m := range m.include {
		if m.Match(value) {
			return true
		}
	}

	// if there's no include listed, include everything
	return len(m.include) == 0
}

func newMatchers(include []string, exclude []string) *matchers {
	m := &matchers{}
	for _, pattern := range include {
		m.include = append(m.include, glob.NewMatcher(pattern))
	}
	for _, pattern := range exclude {
		m.exclude = append(m.exclude, glob.NewMatcher(pattern))
	}
	return m
}

type ProcNetDevEmitter struct {
	logger                *zap.Logger
	statsPool             statser.Pool
	hostInterfacePatterns *matchers
	containerPatterns     *matchers
	ethMatcher            *matchers
}

func NewEmitter(logger *zap.Logger, statsPools statser.Pool, opt emitter.OptProvider) emitter.Emitter {
	opts := opt.Get("procnetdev").(*ProcNetDevOpts)
	pnde := &ProcNetDevEmitter{
		logger:                logger,
		statsPool:             statsPools,
		hostInterfacePatterns: newMatchers(opts.IncludeInterface, opts.ExcludeInterface),
		containerPatterns:     newMatchers(opts.IncludeContainer, opts.ExcludeContainer),
		ethMatcher:            newMatchers([]string{"eth*"}, nil),
	}

	return pnde
}

func (pnde *ProcNetDevEmitter) Emit() {
	ids := getDockerContainerIDs()
	for _, id := range ids {
		d := getDockerContainerDetail(id)
		if d.Id != id {
			pnde.logger.Warn("unexpected id", zap.String("original", id), zap.String("found", d.Id))
			continue
		}
		if d.State.Status != "running" {
			pnde.logger.Info("not running, skipping", zap.String("container", id), zap.String("state", d.State.Status))
			continue
		}
		if !d.State.Running {
			pnde.logger.Info("not running, skipping", zap.String("container", id))
			continue
		}
		if d.State.Pid == 0 {
			pnde.logger.Info("pid is 0, skipping", zap.String("container", id))
			continue
		}
		if !pnde.containerPatterns.Match(d.Name) {
			pnde.logger.Debug("ignored container, skipping", zap.String("container", id))
			continue
		}
		is := pnde.loadInterfaceStats(fmt.Sprintf("/proc/%d/net/dev", d.State.Pid), pnde.ethMatcher)
		for _, i := range is {
			c := pnde.statsPool.Get("interface", i.name, "container", d.Name)
			pnde.emitInterfaceStats(c, "net.docker.", i)
		}
	}

	is := pnde.loadInterfaceStats("/proc/net/dev", pnde.hostInterfacePatterns)
	for _, i := range is {
		c := pnde.statsPool.Get("interface", i.name)
		pnde.emitInterfaceStats(c, "net.host.", i)
	}
}

func (pnde *ProcNetDevEmitter) emitInterfaceStats(c statser.Statser, prefix string, i *Interface) {
	c.Gauge(prefix+"rx.bytes", i.rxBytes)
	c.Gauge(prefix+"rx.packets", i.rxPackets)
	//c.Gauge(prefix+"rx.errors", i.rxErrors)
	//c.Gauge(prefix+"rx.dropped", i.rxDropped)
	//c.Gauge(prefix+"rx.overrun", i.rxOverrun)
	//c.Gauge(prefix+"rx.frame", i.rxFrame)
	//c.Gauge(prefix+"rx.compressed", i.rxCompressed)
	//c.Gauge(prefix+"rx.multicast", i.rxMulticast)
	c.Gauge(prefix+"tx.bytes", i.txBytes)
	c.Gauge(prefix+"tx.packets", i.txPackets)
	//c.Gauge(prefix+"tx.errors", i.txErrors)
	//c.Gauge(prefix+"tx.dropped", i.txDropped)
	//c.Gauge(prefix+"tx.overrun", i.txOverrun)
	//c.Gauge(prefix+"tx.collisions", i.txCollisions)
	//c.Gauge(prefix+"tx.carrier", i.txCarrier)
	//c.Gauge(prefix+"tx.compressed", i.txCompressed)
}

func init() {
	sources.Sources["procnetdev"] = NewEmitter
}
