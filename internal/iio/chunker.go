package iio

import (
	"io"
	"strconv"
)

type Chunker struct {
	line []byte
	next int
	err  error
}

func NewChunker(data []byte) *Chunker {
	return &Chunker{
		line: data,
	}
}

func (c *Chunker) Err() error {
	return c.err
}

func (c *Chunker) SkipWhitespace() {
	if c.err != nil {
		return
	}
	for ; c.next < len(c.line); c.next++ {
		if c.line[c.next] != ' ' {
			return
		}
	}
	c.err = io.EOF
}

func (c *Chunker) NextChunk() []byte {
	c.SkipWhitespace()
	if c.err != nil {
		return nil
	}
	start := c.next
	for ; c.next < len(c.line); c.next++ {
		if c.line[c.next] == ' ' {
			return c.line[start:c.next]
		}
	}
	return c.line[start:]
}

func (c *Chunker) NextString() string {
	return string(c.NextChunk())
}

func (c *Chunker) NextInt64() int64 {
	next := c.NextString()
	if c.err != nil {
		return 0
	}
	value, err := strconv.ParseInt(next, 0, 64)
	if err != nil {
		c.err = err
		return 0
	}
	return value
}

func (c *Chunker) NextUint64() uint64 {
	next := c.NextString()
	if c.err != nil {
		return 0
	}
	value, err := strconv.ParseUint(next, 0, 64)
	if err != nil {
		c.err = err
		return 0
	}
	return value
}
