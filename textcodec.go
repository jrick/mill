// Copyright (c) 2017 Josh Rickmar
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mill

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"strconv"
	"sync"
	"time"
)

const TimeFormat = "2006-01-02 15:04:05.000000-0700"

type textCodec struct {
	writer io.Writer
	pool   sync.Pool
}

// TextCodec creates a Codec that writes encoded human-readable log entries to
// w.
//
// The format is appropiate for both stdout/stderr logging and persistent
// logs written to a log file.  Timestamps are formatted using RFC 3339.
func TextCodec(w io.Writer) Codec {
	return &textCodec{
		writer: w,
		pool: sync.Pool{
			New: func() interface{} { return bytes.NewBuffer(make([]byte, 0, 256)) },
		},
	}
}

func (c *textCodec) EncodeLogEntry(t time.Time, tags []KV, message string, data []Data, encodeDone func(), writeReady <-chan struct{}) {
	buf := c.pool.Get().(*bytes.Buffer)

	buf.WriteString(t.Format(TimeFormat))
	buf.WriteString(" [")
	for i, tag := range tags {
		buf.WriteString(tag.Key)
		if tag.Value != "" {
			buf.WriteByte('=')
			buf.WriteString(tag.Value)
		}
		if i != len(tags)-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteString("] ")
	buf.WriteString(message)

	for _, d := range data {
		ty := d.Type()
		if ty == ValueTypeUnknown || ty > valueTypeMaxValue {
			continue
		}

		buf.WriteString(", ")
		buf.WriteString(d.name)
		buf.WriteByte('=')
		switch d.Type() {
		case ValueTypeString:
			buf.WriteString(d.string)
		case ValueTypeInt64:
			b := strconv.AppendInt(buf.Bytes(), int64(d.numBits), 10)
			*buf = *bytes.NewBuffer(b)
		case ValueTypeUint64:
			b := strconv.AppendUint(buf.Bytes(), d.numBits, 10)
			*buf = *bytes.NewBuffer(b)
		case ValueTypeFloat64:
			b := strconv.AppendFloat(buf.Bytes(), math.Float64frombits(d.numBits), 'g', -1, 64)
			*buf = *bytes.NewBuffer(b)
		case ValueTypeAny:
			fmt.Fprintf(buf, "%v", d.any)
		}
	}

	buf.WriteByte('\n')
	encodeDone()

	<-writeReady
	_, _ = buf.WriteTo(c.writer)
	buf.Reset()
	c.pool.Put(buf)
}
