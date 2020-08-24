package procstat

import (
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/squizzling/stats/internal/iio"
)

type ProcStatCpu struct {
	User      int64
	Nice      int64
	System    int64
	Idle      int64
	IoWait    int64
	Irq       int64
	SoftIrq   int64
	Steal     int64
	Guest     int64
	GuestNice int64
}

type ProcStat struct {
	CPUTotal *ProcStatCpu
	CPUs     map[int]*ProcStatCpu
}

func (ps *ProcStat) parseProcStatCpu(line []byte) *ProcStatCpu {
	var cpu ProcStatCpu
	var err error
	cpu.User, line, _ = iio.NextInt64(line)
	cpu.Nice, line, _ = iio.NextInt64(line)
	cpu.System, line, _ = iio.NextInt64(line)
	cpu.Idle, line, _ = iio.NextInt64(line)
	cpu.IoWait, line, _ = iio.NextInt64(line)
	cpu.Irq, line, _ = iio.NextInt64(line)
	cpu.SoftIrq, line, _ = iio.NextInt64(line)
	cpu.Steal, line, _ = iio.NextInt64(line)
	cpu.Guest, line, _ = iio.NextInt64(line)
	cpu.GuestNice, line, err = iio.NextInt64(line)
	if err != nil {
		return nil
	}
	return &cpu
}

func (ps *ProcStat) handleCpu(cpuId int, line []byte) {
	cpu := ps.parseProcStatCpu(line)
	if cpu == nil {
		return
	}
	if cpuId == -1 {
		ps.CPUTotal = cpu
	} else {
		ps.CPUs[cpuId] = cpu
	}
}

func LoadProcStat(logger *zap.Logger) *ProcStat {
	lines := iio.SplitLines(iio.ReadEntireFile(logger, "/proc/stat"))
	ps := &ProcStat{
		CPUs: make(map[int]*ProcStatCpu),
	}
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		chunk, remainder := iio.NextChunk(line)
		statType := string(chunk)

		if strings.HasPrefix(statType, "cpu") {
			cpuId := -1
			if statType != "cpu" {
				var err error
				cpuId, err = strconv.Atoi(statType[3:])
				if err != nil {
					logger.Warn("Atoi", zap.String("stat-type", statType))
					continue
				}
			}
			ps.handleCpu(cpuId, remainder)
			continue
		}

		switch statType {
		case "intr":
		case "ctxt":
		case "btime":
		case "processes":
		case "procs_running":
		case "procs_blocked":
		case "softirq":
		default:
			logger.Info("Unrecognized stat-type", zap.String("stat-type", string(statType)))
		}

	}

	return ps
}
