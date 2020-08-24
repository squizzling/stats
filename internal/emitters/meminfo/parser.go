package meminfo

import (
	"bytes"
	"strconv"

	"go.uber.org/zap"

	"github.com/squizzling/stats/internal/iio"
)

type MemInfo struct {
	Values map[string]int64
}

func LoadMemInfo(logger *zap.Logger) *MemInfo {
	ms := &MemInfo{
		Values: make(map[string]int64),
	}
	lines := iio.SplitLines(iio.ReadEntireFile(logger, "/proc/meminfo"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		nameChunk, line := iio.NextChunk(line)  // name
		valueChunk, line := iio.NextChunk(line) // value
		typeChunk, line := iio.NextChunk(line)  // unit
		nameChunk = bytes.TrimSuffix(nameChunk, []byte{':'})
		multiplier := int64(1)
		if string(typeChunk) == "kB" {
			multiplier = 1024
		}
		value, err := strconv.ParseInt(string(valueChunk), 10, 64)
		if err != nil {
			logger.Warn("ParseInt", zap.String("value", string(valueChunk)), zap.Error(err))
			continue
		}
		value *= multiplier
		ms.Values[string(nameChunk)] = value
	}

	return ms
}
