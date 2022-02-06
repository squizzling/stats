package diskfree

import (
	"strings"

	"go.uber.org/zap"
	"golang.org/x/sys/unix"

	"github.com/squizzling/glob/pkg/glob"

	"github.com/squizzling/stats/internal/iio"
	"github.com/squizzling/stats/pkg/emitter"
	"github.com/squizzling/stats/pkg/sources"
	"github.com/squizzling/stats/pkg/statser"
)

type DiskFreeEmitter struct {
	logger         *zap.Logger
	statsPool      statser.Pool
	mountPatterns  glob.Matcher
	fsTypePatterns glob.Matcher
}

func NewEmitter(logger *zap.Logger, statsPools statser.Pool, opt emitter.OptProvider) emitter.Emitter {
	opts := opt.Get("diskfree").(*DiskFreeOpts)

	return &DiskFreeEmitter{
		logger:         logger,
		statsPool:      statsPools,
		mountPatterns:  glob.NewACL(opts.IncludeMountPoint, opts.ExcludeMountPoint, len(opts.IncludeMountPoint) == 0),
		fsTypePatterns: glob.NewACL(opts.IncludeFsType, opts.ExcludeFsType, len(opts.IncludeFsType) == 0),
	}
}

const sysBlockRoot = "/sys/block/"

func (dfe *DiskFreeEmitter) Emit() {
	data := iio.ReadEntireFile(dfe.logger, "/proc/self/mountinfo")
	lines := iio.SplitLines(data)
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		chunker := iio.NewChunker(line)
		chunker.NextChunk()                // mount ID
		chunker.NextChunk()                // parent ID
		chunker.NextChunk()                // major:minor
		mountRoot := chunker.NextString()  // root of the mount within the filesystem
		mountPoint := chunker.NextString() // mountpoint relative to process root
		chunker.NextChunk()                // mount options
		for i, optField := 0, chunker.NextString(); optField != "-" && i < 10; i, optField = i+1, chunker.NextString() {
		}
		fsType := chunker.NextString() // fstype
		chunker.NextChunk()            // mount source
		chunker.NextChunk()            // super options

		if !dfe.fsTypePatterns.Match(fsType) {
			continue
		}

		if mountRoot != "/" {
			//dfe.logger.Warn("skipping root", zap.String("mount-point", mountPoint), zap.String("mount-root", mountRoot))
			continue
		}

		// https://github.com/torvalds/linux/blob/v5.16/fs/proc_namespace.c#L155
		mountPoint = strings.Replace(mountPoint, "\\011", "\t", -1)
		mountPoint = strings.Replace(mountPoint, "\\012", "\n", -1)
		mountPoint = strings.Replace(mountPoint, "\\040", " ", -1)
		mountPoint = strings.Replace(mountPoint, "\\134", "\\", -1)

		if !dfe.mountPatterns.Match(mountPoint) {
			continue
		}

		var fs unix.Statfs_t
		err := unix.Statfs(mountPoint, &fs)
		if err != nil {
			dfe.logger.Warn("failed to statfs", zap.String("mount-point", mountPoint), zap.Error(err))
			continue
		}

		availableBytes := fs.Bavail * uint64(fs.Bsize)
		capacityBytes := fs.Blocks * uint64(fs.Bsize)
		usedBytes := capacityBytes - availableBytes
		c := dfe.statsPool.Host("fstype", fsType, "mount", mountPoint)
		c.Gauge("diskfree.available", availableBytes)
		c.Gauge("diskfree.capacity", capacityBytes)
		c.Gauge("diskfree.used", usedBytes)
	}
}

func init() {
	sources.Sources["diskfree"] = NewEmitter
}
