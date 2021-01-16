package iio

import (
	"bytes"
	"io/ioutil"
	"os"

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

func ReadEntries(logger *zap.Logger, path string) []os.FileInfo {
	dir, err := os.Open(path)
	if err != nil {
		logger.Warn("failed to open directory for reading", zap.String("path", path), zap.Error(err))
		return nil
	}
	entries, err := dir.Readdir(-1)
	_ = dir.Close()
	if err != nil {
		logger.Warn("failed to read directories", zap.String("path", path), zap.Error(err))
		return nil
	}
	return entries
}
