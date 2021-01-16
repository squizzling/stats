package ipmi

import (
	"bytes"
	"encoding/csv"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/squizzling/stats/internal/iio"
	"github.com/squizzling/stats/pkg/emitter"
	"github.com/squizzling/stats/pkg/sources"
	"github.com/squizzling/stats/pkg/statser"
)

type IPMIEmitter struct {
	logger    *zap.Logger
	statsPool statser.Pool

	pauseUntil time.Time
}

func NewEmitter(logger *zap.Logger, statsPool statser.Pool, opt emitter.OptProvider) emitter.Emitter {
	return &IPMIEmitter{
		logger:     logger,
		statsPool:  statsPool,
		pauseUntil: time.Unix(0, 0),
	}
}

func (ie *IPMIEmitter) readIPMI() map[int64]map[string]string {
	rawData, err := iio.Execute("/usr/sbin/ipmi-sensors", "--comma-separated-output")
	if ie.check("ipmi-sensors", err) {
		return nil
	}

	r := csv.NewReader(bytes.NewBuffer(rawData))
	recs, err := r.ReadAll()
	if err != nil {
		ie.logger.Error("readall", zap.Error(err))
		return nil
	}

	readings := make(map[int64]map[string]string)

	columnNames := recs[0]
	for _, rec := range recs[1:] {
		reading := make(map[string]string)
		for columnIndex, columnName := range columnNames {
			reading[columnName] = rec[columnIndex]
		}

		id, err := strconv.ParseInt(reading["ID"], 10, 64)
		if err != nil {
			ie.logger.Error("parse id", zap.String("id", reading["ID"]), zap.Error(err))
			continue
		}
		readings[id] = reading
	}

	return readings
}

func (ie *IPMIEmitter) Emit() {
	if time.Now().Before(ie.pauseUntil) {
		return
	}

	sensors := ie.readIPMI()
	for _, sensor := range sensors {
		switch sensor["Type"] {
		case "Temperature": // process
			ie.emitTemp(sensor)
		case "Voltage": // process
			ie.emitVoltage(sensor)
		case "Fan": // ignore
		case "Physical Security": // ignore
		default:
			ie.logger.Warn("unknown sensor type", zap.String("type", sensor["Type"]))
		}
	}
}

func (ie *IPMIEmitter) emitTemp(sensor map[string]string) {
	logger := ie.logger.With(zap.String("sensor", sensor["Name"]))

	if sensor["Units"] != "C" {
		logger.Warn("temperature unit not C", zap.String("unit", sensor["Units"]))
		return
	}

	sensorValue, err := strconv.ParseFloat(sensor["Reading"], 64)
	if err != nil {
		logger.Warn("failed to parse temperature", zap.String("reading", sensor["Reading"]), zap.Error(err))
		return
	}

	client := ie.statsPool.Get("sensor", sensor["Name"])
	client.Gauge("ipmi.temperature", sensorValue)
}

func (ie *IPMIEmitter) emitVoltage(sensor map[string]string) {
	logger := ie.logger.With(zap.String("sensor", sensor["Name"]))

	if sensor["Units"] != "V" {
		logger.Warn("temperature unit not V", zap.String("unit", sensor["Units"]))
		return
	}

	sensorValue, err := strconv.ParseFloat(sensor["Reading"], 64)
	if err != nil {
		logger.Warn("failed to parse voltage", zap.String("reading", sensor["Reading"]), zap.Error(err))
		return
	}

	client := ie.statsPool.Get("sensor", sensor["Name"])
	client.Gauge("ipmi.voltage", sensorValue)
}

func (ie *IPMIEmitter) check(action string, err error) bool {
	if err == nil {
		return false
	}
	ie.logger.Error("error", zap.String("action", action), zap.Error(err))
	ie.pauseUntil = time.Now().Add(5 * time.Minute)
	return true
}

func init() {
	sources.Sources["ipmi"] = NewEmitter
}
