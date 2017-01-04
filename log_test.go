// Copyright (c) 2017 Josh Rickmar
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mill

import (
	"bytes"
	"context"
	"strconv"
	"sync"
	"testing"
	"time"
)

type concurrentSafeBuffer struct {
	bytes.Buffer
	mu sync.Mutex
}

func (w *concurrentSafeBuffer) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	n, err = w.Buffer.Write(p)
	w.mu.Unlock()
	return
}

type blockingConcurrentSafeBuffer struct {
	buf bytes.Buffer
	c   chan struct{}
}

func (w *blockingConcurrentSafeBuffer) Write(p []byte) (n int, err error) {
	<-w.c
	return w.buf.Write(p)
}

func TestLogWithTextCodecIsAsync(t *testing.T) {
	// This will deadlock or timeout if it is not async.
	w := &blockingConcurrentSafeBuffer{c: make(chan struct{})}
	ctx := WithLogger(context.Background(), TextCodec(w))
	Log(ctx, "message 1")
	Log(ctx, "message 2")
	Log(ctx, "message 3")
	if len(w.buf.Bytes()) != 0 {
		t.Fatal("blockingWriter isn't blocking")
	}
	close(w.c)
	Sync()
	t.Log("\n" + w.buf.String())
}

func TestLogWithJSONCodecIsAsync(t *testing.T) {
	// This will deadlock or timeout if it is not async.
	w := &blockingConcurrentSafeBuffer{c: make(chan struct{})}
	ctx := WithLogger(context.Background(), JSONCodec(w))
	Log(ctx, "message 1")
	Log(ctx, "message 2")
	Log(ctx, "message 3")
	if len(w.buf.Bytes()) != 0 {
		t.Fatal("blockingWriter isn't blocking")
	}
	close(w.c)
	Sync()
	t.Log("\n" + w.buf.String())
}

type intStringer struct{ i int }

func (s *intStringer) String() string { return strconv.FormatInt(int64(s.i), 10) }

func TestLogEncodesBeforeReturn(t *testing.T) {
	w := &blockingConcurrentSafeBuffer{c: make(chan struct{})}
	ctx := WithLogger(context.Background(), TextCodec(w))
	i := &intStringer{123}
	data := Any("mutable ref parameter", i)
	Log(ctx, "message", data)
	i.i++
	Log(ctx, "message", data)
	close(w.c)
	Sync()
	t.Log("\n" + w.buf.String())

	lines := bytes.Split(w.buf.Bytes(), []byte("\n"))
	if len(lines) != 3 || len(lines[2]) != 0 {
		t.Fatal("expected 2 lines")
	}
	if !bytes.HasSuffix(lines[0], []byte("123")) {
		t.Error("data was mutated before first encode")
	}
	if !bytes.HasSuffix(lines[1], []byte("124")) {
		t.Error("second log did not use mutated data")
	}
}

func TestLogEntriesAreOrdered(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := WithLogger(context.Background(), TextCodec(buf))
	var wg sync.WaitGroup
	wg.Add(10000)
	for i := 0; i < 10000; i++ {
		i := i
		go func() {
			Log(ctx, "message", Int64("i", int64(i)))
			wg.Done()
		}()
	}
	wg.Wait()
	Sync()

	// timestamps can be lexicographically compared
	prevTime := time.Time{} // epoch
	prevTimeBytes := []byte(prevTime.Format(time.RFC3339))
	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	if len(lines) != 10001 || len(lines[10000]) != 0 {
		t.Fatal("expected 10000 lines")
	}
	for _, line := range lines[:len(lines)-1] {
		timeBytes := bytes.SplitN(line, []byte(" [] "), 2)[0]
		if bytes.Compare(prevTimeBytes, timeBytes) == 1 {
			t.Errorf("lexicographical comparison failed: %s is before %s", prevTimeBytes, timeBytes)
		}
		prevTimeBytes = timeBytes

		// TODO: compare data parameter i
	}
}

func TestMultiContextLogEntriesAreOrdered(t *testing.T) {
	w := &concurrentSafeBuffer{}
	ctxA := WithLogger(context.Background(), TextCodec(w))
	ctxB := WithLogger(context.Background(), TextCodec(w))
	var wg sync.WaitGroup
	wg.Add(20000)
	for i := 0; i < 10000; i++ {
		i := i
		go func() {
			Log(ctxA, "message", Int64("i", int64(i)))
			wg.Done()
		}()
		go func() {
			Log(ctxB, "message", Int64("i", int64(i)))
			wg.Done()
		}()
	}
	wg.Wait()
	Sync()

	// timestamps can be lexicographically compared
	prevTime := time.Time{} // epoch
	prevTimeBytes := []byte(prevTime.Format(time.RFC3339))
	lines := bytes.Split(w.Buffer.Bytes(), []byte("\n"))
	if len(lines) != 20001 || len(lines[20000]) != 0 {
		t.Fatal("expected 20000 lines")
	}
	for _, line := range lines[:len(lines)-1] {
		timeBytes := bytes.SplitN(line, []byte(" [] "), 2)[0]
		if bytes.Compare(prevTimeBytes, timeBytes) == 1 {
			t.Errorf("lexicographical comparison failed: %s is before %s", prevTimeBytes, timeBytes)
		}
		prevTimeBytes = timeBytes

		// TODO: compare data parameter i
	}
}
