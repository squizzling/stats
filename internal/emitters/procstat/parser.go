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

	c := iio.NewChunker(line)
	cpu.User = c.NextInt64()
	cpu.Nice = c.NextInt64()
	cpu.System = c.NextInt64()
	cpu.Idle = c.NextInt64()
	cpu.IoWait = c.NextInt64()
	cpu.Irq = c.NextInt64()
	cpu.SoftIrq = c.NextInt64()
	cpu.Steal = c.NextInt64()
	cpu.Guest = c.NextInt64()
	cpu.GuestNice = c.NextInt64()

	if c.Err() != nil {
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
