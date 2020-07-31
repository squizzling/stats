package kstat

import (
	"bytes"
	"strconv"

	"go.uber.org/zap"

	"github.com/squizzling/stats/internal/iio"
)

type parser struct {
	logger *zap.Logger
}

func (p *parser) parseHeader(hdr []byte) *Kstat {
	// 6 1 0x01 91 4368 4175294187 21581160000870141
	/*
	   ksp->ks_kid, ksp->ks_type, ksp->ks_flags,
	   ksp->ks_ndata, (int)ksp->ks_data_size,
	   ksp->ks_crtime, ksp->ks_snaptime);
	*/

	parts := bytes.Split(hdr, []byte{' '})
	if len(parts) != 7 {
		return nil
	}

	id, err := strconv.ParseInt(string(parts[0]), 0, 32)
	if err != nil {
		p.logger.Info("ParseInt", zap.String("field", "kid"), zap.String("value", string(parts[0])))
		return nil
	}

	t, err := strconv.ParseInt(string(parts[1]), 0, 8)
	if err != nil {
		p.logger.Info("ParseInt", zap.String("field", "ks_type"), zap.String("value", string(parts[1])))
		return nil
	}

	f, err := strconv.ParseInt(string(parts[2]), 0, 8)
	if err != nil {
		p.logger.Info("ParseInt", zap.String("field", "ks_flags"), zap.String("value", string(parts[2])))
		return nil
	}

	ndata, err := strconv.ParseUint(string(parts[3]), 0, 32)
	if err != nil {
		p.logger.Info("ParseUint", zap.String("field", "ks_ndata"), zap.String("value", string(parts[3])))
		return nil
	}

	datasize, err := strconv.ParseUint(string(parts[4]), 0, 32)
	if err != nil {
		p.logger.Info("ParseUint", zap.String("field", "ks_data_size"), zap.String("value", string(parts[4])))
		return nil
	}

	crtime, err := strconv.ParseUint(string(parts[5]), 0, 64)
	if err != nil {
		p.logger.Info("ParseUint", zap.String("field", "ks_crtime"), zap.String("value", string(parts[5])))
		return nil
	}

	snaptime, err := strconv.ParseUint(string(parts[6]), 0, 64)
	if err != nil {
		p.logger.Info("ParseUint", zap.String("field", "ks_snaptime"), zap.String("value", string(parts[6])))
		return nil
	}

	ks := &Kstat{
		Id:           int(id),
		Type:         Type(t),
		Flags:        Flag(f),
		RecordCount:  uint(ndata),
		DataSize:     uint(datasize),
		CreationTime: crtime,
		SnapTime:     snaptime,
		UValues:      make(map[string]uint64, ndata),
		SValues:      make(map[string]int64, ndata),
	}
	return ks
}

func (p *parser) parseNamedLine(kst *Kstat, line []byte) {
	var name string
	var dataType Data
	var value string

	if len(line) == 0 {
		return
	}

	for idx, ch := range line {
		if ch == ' ' {
			name = string(line[:idx])
			line = line[idx:]
			break
		}
	}
	for idx, ch := range line {
		if ch != ' ' {
			line = line[idx:]
			break
		}
	}
	for idx, ch := range line {
		if ch == ' ' {
			dt, err := strconv.ParseUint(string(line[:idx]), 0, 8)
			dataType = Data(dt)
			if err != nil {
				p.logger.Warn("ParseUint", zap.String("dataType", string(line[:idx])))
			}
			line = line[idx:]
			break
		}
	}
	for idx, ch := range line {
		if ch != ' ' {
			line = line[idx:]
			break
		}
	}
	value = string(line)

	switch dataType {
	case KsdUint64, KsdUint32, KsdUlong:
		u, err := strconv.ParseUint(value, 0, 64)
		if err != nil {
			p.logger.Warn("ParseUint", zap.String("ksdUlong", value))
		}
		kst.UValues[name] = u
	case KsdInt64, KsdInt32, KsdLong:
		u, err := strconv.ParseInt(value, 0, 64)
		if err != nil {
			p.logger.Warn("ParseUint", zap.String("ksdLong", value))
		}
		kst.SValues[name] = u
	default:
		p.logger.Warn(
			"unknown type",
			zap.String("key", name),
			zap.Int("data-type", int(dataType)),
		)
	}
}

func (p *parser) parseNamed(lines [][]byte, kst *Kstat) *Kstat {
	for _, line := range lines[2:] {
		p.parseNamedLine(kst, line)
	}
	return kst
}

func (p *parser) parseIo(lines [][]byte, kst *Kstat) *Kstat {
	if len(lines) != 4 {
		p.logger.Info("lines is not 4", zap.Int("lines", len(lines)))
		return nil
	}

	headerValue, headerLine := iio.NextChunk(lines[1])
	valueValue, valueLine := iio.NextChunk(lines[2])

	for len(headerValue) > 0 && len(valueValue) > 0 {
		value, err := strconv.ParseUint(string(valueValue), 0, 64)
		if err != nil {
			p.logger.Warn("failed to parse", zap.String("key", string(headerValue)), zap.String("value", string(valueValue)))
		} else {
			kst.UValues[string(headerValue)] = value
		}
		headerValue, headerLine = iio.NextChunk(headerLine)
		valueValue, valueLine = iio.NextChunk(valueLine)
	}
	return kst
}

func (p *parser) loadKstat(file string) *Kstat {
	lines := iio.SplitLines(iio.ReadEntireFile(p.logger, file))
	if len(lines) == 0 {
		return nil
	}

	kst := p.parseHeader(lines[0])
	switch kst.Type {
	case KstNamed:
		return p.parseNamed(lines, kst)
	case KstIo:
		return p.parseIo(lines, kst)
	default:
		p.logger.Warn("unknown type", zap.Int("type", int(kst.Type)))
		return nil
	}
}

func LoadKstat(file string, logger *zap.Logger) *Kstat {
	p := &parser{
		logger: logger,
	}
	return p.loadKstat(file)
}
