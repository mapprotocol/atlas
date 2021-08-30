package params

import (
	"encoding/csv"
	"fmt"
	"io"
	"sync"
)

type WriterCloser interface {
	io.Writer
	io.Closer
}

// A CSVRecorder enables easy writing of CSV data a specified writer.
// The header is written on creation. Writing is thread safe.
type CSVRecorder struct {
	writer    csv.Writer
	backingWC WriterCloser
	writeMu   sync.Mutex
}

// NewCSVRecorder creates a CSV recorder that writes to the supplied writer.
// The writer is retained and can be closed by calling CSVRecorder.Close()
// The header is immediately written upon construction.
func NewCSVRecorder(wc WriterCloser, fields ...string) *CSVRecorder {
	c := &CSVRecorder{
		writer:    *csv.NewWriter(wc),
		backingWC: wc,
	}
	c.writer.Write(fields)
	return c
}

// WriteRow writes out as csv row. Will convert the values to a string using "%v".
func (c *CSVRecorder) Write(values ...interface{}) {
	if c == nil {
		return
	}
	strs := make([]string, 0, len(values))
	for _, v := range values {
		strs = append(strs, fmt.Sprintf("%v", v))
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	c.writer.Write(strs)
}

// Close closes the writer. This is a no-op for a nil receiver.
func (c *CSVRecorder) Close() error {
	if c == nil {
		return nil
	}
	c.writer.Flush()
	return c.backingWC.Close()
}
