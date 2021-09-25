package smart

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/squizzling/stats/internal/iio"
	"github.com/squizzling/stats/pkg/emitter"
	"github.com/squizzling/stats/pkg/sources"
	"github.com/squizzling/stats/pkg/statser"
)

type SmartEmitter struct {
	logger    *zap.Logger
	statsPool statser.Pool

	pauseUntil time.Time
}

func NewEmitter(logger *zap.Logger, statsPool statser.Pool, opt emitter.OptProvider) emitter.Emitter {
	return &SmartEmitter{
		logger:     logger,
		statsPool:  statsPool,
		pauseUntil: time.Unix(0, 0),
	}
}

func (se *SmartEmitter) Emit() {
	if time.Now().Before(se.pauseUntil) {
		return
	}

	for _, disk := range se.scanForDisks() {
		d := se.getSmartData(disk)
		if d == nil {
			continue
		}

		sn := d.information["Serial Number"]

		for _, attribute := range d.attributeByName {
			client := se.statsPool.Host("serial", sn, "attribute", attribute.name)
			client.Gauge("smart.attribute", attribute.rawValue)
			//fmt.Printf("%s %s %v\n", sn, attribute.name, attribute.rawValue)
		}
	}
}

var scanFormat = regexp.MustCompile("^(/[0-9a-zA-Z/_\\-]+).*$")

func (se *SmartEmitter) scanForDisks() []string {
	rawDiskInfo, err := iio.Execute("/usr/sbin/smartctl", "--scan")
	if se.check("scan", err) {
		return nil
	}

	var found []string
	for _, line := range iio.SplitLines(rawDiskInfo) {
		if matches := scanFormat.FindAllSubmatch(line, -1); matches != nil {
			found = append(found, string(matches[0][1]))
		}
	}
	return found
}

type data struct {
	information     map[string]string
	attributeById   map[int64]*attribute
	attributeByName map[string]*attribute
}

func (se *SmartEmitter) getSmartData(drive string) *data {
	rawData, err := iio.Execute("/usr/sbin/smartctl", "--attributes", "--info", drive)
	if se.check("get-smart-data", err) {
		return nil
	}

	d := &data{
		information:     make(map[string]string),
		attributeById:   make(map[int64]*attribute),
		attributeByName: make(map[string]*attribute),
	}
	mode := d.ignoreLine
	for _, line := range iio.SplitLines(rawData) {
		if string(line) == "=== START OF INFORMATION SECTION ===" {
			mode = d.parseInfoLine
		} else if bytes.HasPrefix(line, []byte("ID#")) {
			mode = d.parseAttributeLine
		} else if len(line) == 0 {
			mode = d.ignoreLine
		} else {
			mode(line)
		}
	}
	return d
}

func (d *data) ignoreLine(line []byte) {}

func (d *data) parseInfoLine(line []byte) {
	/*
		=== START OF INFORMATION SECTION ===
		Device Model:     ST6000DX000-1H217Z
		Serial Number:    Z4D08518
		LU WWN Device Id: 5 000c50 078ce307e
		Firmware Version: CC48
		User Capacity:    6,001,175,126,016 bytes [6.00 TB]
		Sector Sizes:     512 bytes logical, 4096 bytes physical
		Rotation Rate:    7200 rpm
		Form Factor:      3.5 inches
		Device is:        Not in smartctl database [for details use: -P showall]
		ATA Version is:   ACS-3 T13/2161-D revision 3b
		SATA Version is:  SATA 3.1, 6.0 Gb/s (current: 6.0 Gb/s)
		Local Time is:    Sat Sep  5 14:57:26 2020 AEST
		SMART support is: Available - device has SMART capability.
		SMART support is: Enabled
	*/
	if split := bytes.SplitN(line, []byte{':'}, 2); len(split) == 2 {
		d.information[string(split[0])] = string(bytes.TrimSpace(split[1]))
	}
}

type attribute struct {
	id            int64
	name          string
	value         int64
	worst         int64
	threshold     int64
	preFail       bool
	updatedAlways bool
	whenFailed    string
	rawValue      int64
}

var attributeFormat = regexp.MustCompile(`\s*([0-9]+)\s([^\s]+)\s+(0x[0-9a-fA-F]{4})\s+(\d+)\s+(\d+)\s+(\d+)\s+([a-zA-Z_\\-]+)\s+([a-zA-Z_\\-]+)\s+([a-zA-Z_\\-]+)\s+(.*)`)

func (d *data) parseAttributeLine(line []byte) {
	parts := attributeFormat.FindStringSubmatch(string(line))
	if len(parts) != 11 {
		return
	}
	a := &attribute{
		id:            0,
		name:          parts[2],
		value:         0,
		worst:         0,
		threshold:     0,
		preFail:       false,
		updatedAlways: false,
		whenFailed:    parts[9],
		rawValue:      0,
	}

	var err error
	if a.id, err = strconv.ParseInt(parts[1], 10, 32); err != nil {
		fmt.Printf("failed to parse 1 %s\n", err)
		return
	}
	if a.value, err = strconv.ParseInt(parts[4], 10, 32); err != nil {
		fmt.Printf("failed to parse 4 %s\n", err)
		return
	}
	if a.worst, err = strconv.ParseInt(parts[5], 10, 32); err != nil {
		fmt.Printf("failed to parse 5 %s\n", err)
		return
	}
	if a.threshold, err = strconv.ParseInt(parts[6], 10, 32); err != nil {
		fmt.Printf("failed to parse 6 %s\n", err)
		return
	}
	if parts[7] == "Pre-fail" {
		a.preFail = true
	} else if parts[7] != "Old_age" {
		fmt.Printf("7 = %s\n", parts[7])
		return
	}

	if parts[8] == "Always" {
		a.updatedAlways = true
	} else if parts[8] != "Offline" {
		fmt.Printf("8 = %s\n", parts[8])
		return
	}

	rawValueParts := strings.SplitN(parts[10], " ", 2)
	if a.rawValue, err = strconv.ParseInt(rawValueParts[0], 10, 64); err != nil {
		fmt.Printf("failed to parse 10 %s\n", err)
		return
	}

	d.attributeById[a.id] = a
	d.attributeByName[a.name] = a

	/*
	   === START OF READ SMART DATA SECTION ===
	   SMART Attributes Data Structure revision number: 10
	   Vendor Specific SMART Attributes with Thresholds:
	   ID# ATTRIBUTE_NAME          FLAG     VALUE WORST THRESH TYPE      UPDATED  WHEN_FAILED RAW_VALUE
	     1 Raw_Read_Error_Rate     0x000f   118   099   006    Pre-fail  Always       -       177049170
	     3 Spin_Up_Time            0x0003   095   095   000    Pre-fail  Always       -       0
	     4 Start_Stop_Count        0x0032   100   100   020    Old_age   Always       -       5
	     5 Reallocated_Sector_Ct   0x0033   100   100   010    Pre-fail  Always       -       0
	     7 Seek_Error_Rate         0x000f   090   060   030    Pre-fail  Always       -       966530877
	     9 Power_On_Hours          0x0032   067   067   000    Old_age   Always       -       29043
	    10 Spin_Retry_Count        0x0013   100   100   097    Pre-fail  Always       -       0
	    12 Power_Cycle_Count       0x0032   100   100   020    Old_age   Always       -       5
	   183 Runtime_Bad_Block       0x0032   100   100   000    Old_age   Always       -       0
	   184 End-to-End_Error        0x0032   100   100   099    Old_age   Always       -       0
	   187 Reported_Uncorrect      0x0032   100   100   000    Old_age   Always       -       0
	   188 Command_Timeout         0x0032   100   100   000    Old_age   Always       -       0
	   189 High_Fly_Writes         0x003a   100   100   000    Old_age   Always       -       0
	   190 Airflow_Temperature_Cel 0x0022   047   036   045    Old_age   Always   In_the_past 53 (255 255 61 43 0)
	   191 G-Sense_Error_Rate      0x0032   097   097   000    Old_age   Always       -       6873
	   192 Power-Off_Retract_Count 0x0032   100   100   000    Old_age   Always       -       2
	   193 Load_Cycle_Count        0x0032   100   100   000    Old_age   Always       -       1213
	   194 Temperature_Celsius     0x0022   053   064   000    Old_age   Always       -       53 (0 22 0 0 0)
	   195 Hardware_ECC_Recovered  0x001a   064   054   000    Old_age   Always       -       177049170
	   197 Current_Pending_Sector  0x0012   100   100   000    Old_age   Always       -       0
	   198 Offline_Uncorrectable   0x0010   100   100   000    Old_age   Offline      -       0
	   199 UDMA_CRC_Error_Count    0x003e   200   200   000    Old_age   Always       -       0
	   240 Head_Flying_Hours       0x0000   100   253   000    Old_age   Offline      -       29043 (190 138 0)
	   241 Total_LBAs_Written      0x0000   100   253   000    Old_age   Offline      -       39862642689
	   242 Total_LBAs_Read         0x0000   100   253   000    Old_age   Offline      -       7734202075417
	*/
}

func (se *SmartEmitter) check(action string, err error) bool {
	if err == nil {
		return false
	}
	se.logger.Error("error", zap.String("action", action), zap.Error(err))
	se.pauseUntil = time.Now().Add(5 * time.Minute)
	return true
}

func init() {
	sources.Sources["smart"] = NewEmitter
}
