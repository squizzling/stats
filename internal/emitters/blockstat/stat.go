package blockstat

import (
	"github.com/squizzling/glob/pkg/glob"
	"go.uber.org/zap"

	"github.com/squizzling/stats/internal/iio"
	"github.com/squizzling/stats/pkg/emitter"
	"github.com/squizzling/stats/pkg/sources"
	"github.com/squizzling/stats/pkg/statser"
)

type BlockStatEmitter struct {
	logger         *zap.Logger
	statsPool      statser.Pool
	devicePatterns glob.Matcher
}

func NewEmitter(logger *zap.Logger, statsPools statser.Pool, opt emitter.OptProvider) emitter.Emitter {
	opts := opt.Get("blockstat").(*BlockStatOpts)
	return &BlockStatEmitter{
		logger:         logger,
		statsPool:      statsPools,
		devicePatterns: glob.NewACL(opts.IncludeDevice, opts.ExcludeDevice, len(opts.IncludeDevice) == 0),
	}
}

const sysBlockRoot = "/sys/block/"

func (bse *BlockStatEmitter) Emit() {
	es := iio.ReadEntries(bse.logger, sysBlockRoot)
	for _, e := range es {
		if !bse.devicePatterns.Match(e.Name()) {
			continue
		}

		bs := LoadBlockStat(bse.logger, e.Name())
		if bs != nil {
			if bs.readIOs == 0 && bs.writeIOs == 0 {
				continue
			}
			c := bse.statsPool.Get("device", bs.name)
			c.Gauge("blockstat.read.requests", bs.readIOs)
			c.Gauge("blockstat.read.merges", bs.readMerges)
			c.Gauge("blockstat.read.sectors", bs.readSectors)
			c.Gauge("blockstat.read.ticks", bs.readTicks)

			c.Gauge("blockstat.write.requests", bs.writeIOs)
			c.Gauge("blockstat.write.merges", bs.writeMerges)
			c.Gauge("blockstat.write.sectors", bs.writeSectors)
			c.Gauge("blockstat.write.ticks", bs.writeTicks)

			c.Gauge("blockstat.inflight", bs.inFlight)
			c.Gauge("blockstat.ioticks", bs.ioTicks)
			c.Gauge("blockstat.timeinqueue", bs.timeInQueue)

			if bs.version >= v4_19 {
				c.Gauge("blockstat.discard.requests", bs.discardIOs)
				c.Gauge("blockstat.discard.merges", bs.discardMerges)
				c.Gauge("blockstat.discard.sectors", bs.discardSectors)
				c.Gauge("blockstat.discard.ticks", bs.discardTicks)
			}
			if bs.version >= v5_5 {
				c.Gauge("blockstat.flush.requests", bs.flushIOs)
				c.Gauge("blockstat.flush.ticks", bs.flushTicks)
			}
		}
	}
}

func init() {
	sources.Sources["blockstat"] = NewEmitter
}
