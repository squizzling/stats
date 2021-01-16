package procnetdev

import (
	"strings"

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
	var err error

	iface.name, line, _ = iio.NextString(line)
	iface.name = strings.TrimRight(iface.name, ":")
	iface.rxBytes, line, _ = iio.NextUint64(line)
	iface.rxPackets, line, _ = iio.NextUint64(line)
	iface.rxErrors, line, _ = iio.NextUint64(line)
	iface.rxDropped, line, _ = iio.NextUint64(line)
	iface.rxOverrun, line, _ = iio.NextUint64(line)
	iface.rxFrame, line, _ = iio.NextUint64(line)
	iface.rxCompressed, line, _ = iio.NextUint64(line)
	iface.rxMulticast, line, _ = iio.NextUint64(line)
	iface.txBytes, line, _ = iio.NextUint64(line)
	iface.txPackets, line, _ = iio.NextUint64(line)
	iface.txErrors, line, _ = iio.NextUint64(line)
	iface.txDropped, line, _ = iio.NextUint64(line)
	iface.txOverrun, line, _ = iio.NextUint64(line)
	iface.txCollisions, line, _ = iio.NextUint64(line)
	iface.txCarrier, line, _ = iio.NextUint64(line)
	iface.txCompressed, line, err = iio.NextUint64(line)

	if err != nil {
		return nil
	}

	return iface
}

func (pnde *ProcNetDevEmitter) loadInterfaceStats(filename string, m *matchers) []*Interface {
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
