package pmbus

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/squizzling/stats/pkg/emitter"
	"github.com/squizzling/stats/pkg/sources"
	"github.com/squizzling/stats/pkg/statser"
)

type CorsairEmitter struct {
	logger *zap.Logger

	statsPool statser.Pool
}

const vidCorsair = 0x1b1c
const pidHX750i = 0x1c05
const pidHX1000i = 0x1c07

func (ce *CorsairEmitter) Emit() {
	dev := newPmbusDevice(ce.logger, vidCorsair, pidHX750i)
	if dev == nil {
		return
	}
	defer dev.close()

	client := ce.statsPool.Get()
	b := dev.execReadFromPage(0, pmbusReadTemperature1)
	ce.statsPool.Get("sensor", "1").Gauge("pmbus.temperature", linearToFloat64(b[2:4]))
	b = dev.execReadFromPage(0, pmbusReadTemperature2)
	ce.statsPool.Get("sensor", "2").Gauge("pmbus.temperature", linearToFloat64(b[2:4]))
	b = dev.execReadFromPage(0, pmbusReadFanSpeed1)
	ce.statsPool.Get("fan", "1").Gauge("pmbus.fanspeed", linearToFloat64(b[2:4]))
	b = dev.execReadFromPage(0, pmbusReadVin)
	client.Gauge("pmbus.voltage_in", linearToFloat64(b[2:4]))
	b = dev.execReadFromPage(0, pmbusMfrSpecific30)
	client.Gauge("pmbus.power_in", linearToFloat64(b[2:4]))

	for page, name := range []string{"12", "5", "3.3"} {
		client = ce.statsPool.Get("rail", name)
		b = dev.execReadFromPage(byte(page), pmbusReadVOut)
		client.Gauge("pmbus.voltage_out", linearToFloat64(b[2:4]))
		b = dev.execReadFromPage(byte(page), pmbusReadIOut)
		client.Gauge("pmbus.current_out", linearToFloat64(b[2:4]))
		b = dev.execReadFromPage(byte(page), pmbusReadPOut)
		client.Gauge("pmbus.power_out", linearToFloat64(b[2:4]))
	}
}

func NewEmitter(logger *zap.Logger, statsPool statser.Pool) emitter.Emitter {
	dev := newPmbusDevice(logger, vidCorsair, pidHX750i)
	if dev == nil {
		return nil
	}
	// []byte{0xfe, 0x3, 0x48, 0x58, 0x31, 0x30, 0x30, 0x30, 0x69, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	b := dev.execWriteAddress(0xfe, pmbusClearFaults)
	fmt.Printf("%#v\n", b)
	dev.close()
	return &CorsairEmitter{
		logger:    logger,
		statsPool: statsPool,
	}
}

func init() {
	sources.Sources["pmbus"] = NewEmitter
}
