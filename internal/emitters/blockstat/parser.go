package blockstat

import (
	"bytes"
	"fmt"
	"path"

	"go.uber.org/zap"

	"github.com/squizzling/stats/internal/iio"
)

// https://www.kernel.org/doc/html/latest/block/stat.html
const (
	vEarly = 0
	v4_19  = 1
	v5_5   = 2
)

type BlockStat struct {
	name    string
	version int

	readIOs     uint64
	readMerges  uint64
	readSectors uint64
	readTicks   uint64

	writeIOs     uint64
	writeMerges  uint64
	writeSectors uint64
	writeTicks   uint64

	inFlight    uint64
	ioTicks     uint64
	timeInQueue uint64

	// v4.19+
	discardIOs     uint64
	discardMerges  uint64
	discardSectors uint64
	discardTicks   uint64

	// v5.5+
	flushIOs   uint64
	flushTicks uint64
}

func LoadBlockStat(logger *zap.Logger, deviceName string) *BlockStat {
	blockStatFilename := path.Join(sysBlockRoot, deviceName, "stat")
	line := iio.ReadEntireFile(logger, blockStatFilename)
	if line == nil {
		fmt.Printf("?\n")
		return nil
	}
	if bytes.HasSuffix(line, []byte{'\n'}) {
		line = line[:len(line)-1]
	}

	c := iio.NewChunker(line)
	bs := &BlockStat{}
	bs.name = deviceName
	bs.readIOs = c.NextUint64()
	bs.readMerges = c.NextUint64()
	bs.readSectors = c.NextUint64()
	bs.readTicks = c.NextUint64()
	bs.writeIOs = c.NextUint64()
	bs.writeMerges = c.NextUint64()
	bs.writeSectors = c.NextUint64()
	bs.writeTicks = c.NextUint64()
	bs.inFlight = c.NextUint64()
	bs.ioTicks = c.NextUint64()
	bs.timeInQueue = c.NextUint64()
	if c.IsEOF() {
		bs.version = vEarly
		return bs
	}

	bs.discardIOs = c.NextUint64()
	bs.discardMerges = c.NextUint64()
	bs.discardSectors = c.NextUint64()
	bs.discardTicks = c.NextUint64()

	if c.IsEOF() {
		bs.version = v4_19
		return bs

	}
	bs.flushIOs = c.NextUint64()
	bs.flushTicks = c.NextUint64()

	if c.Err() != nil {
		fmt.Printf("err %s\n", c.Err())
		return nil
	}

	bs.version = v5_5
	return bs
}
