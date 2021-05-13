package sysfs

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/squizzling/stats/internal/iio"
	"github.com/squizzling/stats/pkg/emitter"
	"github.com/squizzling/stats/pkg/sources"
	"github.com/squizzling/stats/pkg/statser"
)

const hwmonBasePath = "/sys/class/hwmon"

// SysfsEmitter is sort of a sysfs reader, but really it's just an hwmon reader.  It will
// likely be refactored at some point.
type SysfsEmitter struct {
	logger    *zap.Logger
	statsPool statser.Pool
}

type sensor struct {
	label  string
	values map[string]int64
}

type device struct {
	name string

	// Numbering usually starts from 1, except for voltages which start
	// from 0 (because most data sheets use this).
	// ^
	// |
	// +-- Use a map, this is not a guarantee.
	temperatures map[int]*sensor
	pwms         map[int]*sensor
}

func NewEmitter(logger *zap.Logger, statsPool statser.Pool, opt emitter.OptProvider) emitter.Emitter {
	return &SysfsEmitter{
		logger:    logger,
		statsPool: statsPool,
	}
}

func (d *device) getTemperatureSensor(index int) *sensor {
	s, ok := d.temperatures[index]
	if ok {
		return s
	}
	s = &sensor{
		label:  fmt.Sprintf("unnamed_temp_sensor_%d", index),
		values: make(map[string]int64),
	}
	d.temperatures[index] = s
	return s
}

func (d *device) getPWMSensor(index int) *sensor {
	s, ok := d.pwms[index]
	if ok {
		return s
	}
	s = &sensor{
		label:  fmt.Sprintf("unnamed_pwm_sensor_%d", index),
		values: make(map[string]int64),
	}
	d.pwms[index] = s
	return s
}

func (se *SysfsEmitter) parseFileName(devicePath, dataFileName string) (string, int, string, bool) {
	prefixLength := 0
	if strings.HasPrefix(dataFileName, "temp") {
		prefixLength = 4
	} else if strings.HasPrefix(dataFileName, "pwm") { // pwm sensors have no _suffix, so we fake it
		prefixLength = 3
		dataFileName += "_value"
	} else if strings.HasPrefix(dataFileName, "in") {
		se.logger.Debug("ignoring voltage", zap.String("device", devicePath), zap.String("file", dataFileName))
		return "", 0, "", false
	} else if strings.HasPrefix(dataFileName, "def") { // unsure what this is meant to define, but it's something with PWM
		return "", 0, "", false
	} else {
		se.logger.Warn("sensor with unknown type", zap.String("device", devicePath), zap.String("file", dataFileName))
		return "", 0, "", false
	}

	underscoreIndex := strings.IndexByte(dataFileName, '_')
	if underscoreIndex == -1 {
		se.logger.Warn(
			"sensor with no underscore",
			zap.String("device", devicePath),
			zap.String("file", dataFileName),
		)
		return "", 0, "", false
	}
	sensorIndex, err := strconv.Atoi(dataFileName[prefixLength:underscoreIndex])
	if err != nil {
		se.logger.Warn("sensor with invalid index",
			zap.String("device", devicePath),
			zap.String("file", dataFileName),
		)
		return "", 0, "", false
	}
	sensorName := dataFileName[underscoreIndex+1:]
	return dataFileName[:prefixLength], sensorIndex, sensorName, true
}

func (se *SysfsEmitter) readDevice(devicePath string) *device {
	dataFiles := iio.ReadEntries(se.logger, devicePath)
	dev := &device{
		temperatures: make(map[int]*sensor),
		pwms:         make(map[int]*sensor),
	}
	for _, dataFile := range dataFiles {
		dataFileName := dataFile.Name()
		if iio.IsDir(dataFile.Mode()) {
			se.logger.Debug("ignoring file", zap.String("device-path", devicePath), zap.String("data-file", dataFileName))
			continue
		}
		if iio.IsSymlink(dataFile.Mode()) {
			se.logger.Debug("ignoring link", zap.String("device-path", devicePath), zap.String("data-file", dataFileName))
			continue
		}

		valueString := strings.TrimSpace(string(iio.ReadEntireFile(se.logger, path.Join(devicePath, dataFileName))))
		if dataFileName == "name" {
			dev.name = valueString
		} else if dataFileName == "update_interval" { // in milliseconds
		} else if dataFileName == "uevent" { // ignore
		} else if sensorType, sensorIndex, sensorName, ok := se.parseFileName(devicePath, dataFileName); ok {
			// temperature sensor
			switch sensorType {
			case "temp":
				s := dev.getTemperatureSensor(sensorIndex)
				if sensorName == "label" {
					s.label = valueString
				} else {
					value, err := strconv.ParseInt(valueString, 10, 64)
					if err != nil {
						se.logger.Error("failed to read sensor value", zap.String("device-path", devicePath), zap.String("data-file", dataFileName), zap.String("data", valueString), zap.Error(err))
					} else {
						s.values[sensorName] = value
					}
				}
			case "pwm":
				s := dev.getPWMSensor(sensorIndex)
				if sensorName == "label" {
					s.label = valueString
				} else {
					value, err := strconv.ParseInt(valueString, 10, 64)
					if err != nil {
						se.logger.Error("failed to read sensor value", zap.String("device-path", devicePath), zap.String("data-file", dataFileName), zap.String("data", valueString), zap.Error(err))
					} else {
						s.values[sensorName] = value
					}
				}
			default:
				se.logger.Info(
					"unknown sensor",
					zap.String("device", devicePath),
					zap.String("sensor-type", sensorType),
					zap.Int("sensor-index", sensorIndex),
					zap.String("sensor-name", sensorName),
				)
			}
		}
	}

	return dev
}

func (se *SysfsEmitter) Emit() {
	potentialDeviceNames := iio.ReadEntries(se.logger, hwmonBasePath)

	for _, potentialDeviceName := range potentialDeviceNames {
		if !iio.IsDir(potentialDeviceName.Mode()) && !iio.IsSymlink(potentialDeviceName.Mode()) {
			continue
		}
		if len(potentialDeviceName.Name()) < 6 || !strings.HasPrefix(potentialDeviceName.Name(), "hwmon") {
			continue
		}

		fullDevicePath := path.Join(hwmonBasePath, potentialDeviceName.Name())
		device := se.readDevice(fullDevicePath)

		if device == nil {
			continue
		}
		if device.name == "" {

			continue
		}
		//fmt.Printf("dumping device %s\n", device.name)
		for _, sensor := range device.temperatures {
			//fmt.Printf("dumping sensor %s\n", sensor.label)
			for valueName, value := range sensor.values {
				if valueName == "input" {
					se.statsPool.Get(
						"device", device.name,
						"sensor", sensor.label,
					).Gauge("sysfs.hwmon.temperature", float64(value)/1000)
				}
			}
		}

		for _, sensor := range device.pwms {
			//fmt.Printf("dumping sensor %s\n", sensor.label)
			for valueName, value := range sensor.values {
				if valueName == "value" {
					se.statsPool.Get(
						"device", device.name,
						"sensor", sensor.label,
					).Gauge("sysfs.hwmon.pwm", (float64(value) / 255)*100)
				}
			}
		}
	}
}

func init() {
	sources.Sources["sysfs"] = NewEmitter
}
