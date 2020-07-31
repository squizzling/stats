package iio

import (
	"bytes"
	"io"
	"io/ioutil"
	"strconv"

	"go.uber.org/zap"
)

func ReadEntireFile(logger *zap.Logger, file string) []byte {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		logger.Warn("ReadEntireFile failed", zap.Error(err))
		return nil
	}
	return b
}

func SplitLines(b []byte) [][]byte {
	return bytes.Split(b, []byte{'\n'})
}

func SkipWhitespace(data []byte) []byte {
	for idx, ch := range data {
		if ch != ' ' {
			return data[idx:]
		}
	}
	return nil
}

func NextChunk(data []byte) ([]byte, []byte) {
	data = SkipWhitespace(data)
	if len(data) == 0 {
		return nil, nil
	}
	for idx, ch := range data {
		if ch == ' ' {
			return data[:idx], SkipWhitespace(data[idx:])
		}
	}
	return data, nil // no whitespace, return the whole chunk
}

func NextInt64(data []byte) (int64, []byte, error) {
	next, remainder := NextChunk(data)
	if len(next) == 0 {
		return 0, nil, io.EOF
	}
	value, err := strconv.ParseInt(string(next), 0, 64)
	if err != nil {
		return 0, nil, err
	}
	return value, remainder, nil
}
