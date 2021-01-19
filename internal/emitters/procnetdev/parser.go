package procnetdev

import (
	"strings"

	"github.com/squizzling/glob/pkg/glob"

	"github.com/squizzling/stats/internal/iio"
)

type Interface struct {
	name         string
	rxBytes      uint64
	rxPackets    uint64
	rxErrors     uint64
	rxDropped    uint64
	rxOverrun    uint64
	rxFrame      uint64
	rxCompressed uint64
	rxMulticast  uint64
	txBytes      uint64
	txPackets    uint64
	txErrors     uint64
	txDropped    uint64
	txOverrun    uint64
	txCollisions uint64
	txCarrier    uint64
	txCompressed uint64
}

func parseInterface(line []byte) *Interface {
	iface := &Interface{}

	c := iio.NewChunker(line)
	iface.name = strings.TrimRight(c.NextString(), ":")
	iface.rxBytes = c.NextUint64()
	iface.rxPackets = c.NextUint64()
	iface.rxErrors = c.NextUint64()
	iface.rxDropped = c.NextUint64()
	iface.rxOverrun = c.NextUint64()
	iface.rxFrame = c.NextUint64()
	iface.rxCompressed = c.NextUint64()
	iface.rxMulticast = c.NextUint64()
	iface.txBytes = c.NextUint64()
	iface.txPackets = c.NextUint64()
	iface.txErrors = c.NextUint64()
	iface.txDropped = c.NextUint64()
	iface.txOverrun = c.NextUint64()
	iface.txCollisions = c.NextUint64()
	iface.txCarrier = c.NextUint64()
	iface.txCompressed = c.NextUint64()

	if c.Err() != nil {
		return nil
	}

	return iface
}

func (pnde *ProcNetDevEmitter) loadInterfaceStats(filename string, m glob.Matcher) []*Interface {
	var is []*Interface
	lines := iio.SplitLines(iio.ReadEntireFile(pnde.logger, filename))
	for _, line := range lines {
		if i := parseInterface(line); i != nil {
			if m.Match(i.name) {
				is = append(is, i)
			}
		}
	}
	return is
}
