// Copyright (c) 2017 Josh Rickmar
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mill

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type jsonCodec struct {
	writer io.Writer
}

type jsonSchema struct {
	Date        string                 `json:"date"`
	DateUnix    int64                  `json:"dateunix"`
	NanoSeconds int64                  `json:"nanoseconds"`
	Tags        []string               `json:"tags,omitempty"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

type jsonKV struct {
	k, v string
}

func (kv *jsonKV) MarshalJSON() ([]byte, error) {
	if kv.v == "" {
		return json.Marshal(kv.k)
	}
	return json.Marshal(fmt.Sprintf("%s=%s", kv.k, kv.v))
}

type jsonDataObject struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// JSONCodec creates a Codec that writes encoded log entries as JSON objects to
// w.
func JSONCodec(w io.Writer) Codec {
	return &jsonCodec{w}
}

func mapKV(tags []KV) []string {
	r := make([]string, len(tags))
	for i := range tags {
		if tags[i].Value == "" {
			r[i] = tags[i].Key
		} else {
			r[i] = tags[i].Key + "=" + tags[i].Value
		}
	}
	return r
}

func mapData(data []Data) map[string]interface{} {
	r := make(map[string]interface{})
	for i := range data {
		r[data[i].Name()] = data[i].Value()
	}
	return r
}

func (c *jsonCodec) EncodeLogEntry(t time.Time, tags []KV, message string, data []Data, encodeDone func(), writeReady <-chan struct{}) {
	b, err := json.Marshal(jsonSchema{
		Date:        t.Format(TimeFormat),
		DateUnix:    t.Unix(),
		NanoSeconds: int64(t.Nanosecond()),
		Tags:        mapKV(tags),
		Message:     message,
		Data:        mapData(data),
	})
	encodeDone()
	if err != nil {
		return
	}
	<-writeReady
	c.writer.Write(b)
}
