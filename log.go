// Copyright (c) 2017 Josh Rickmar
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mill

import (
	"context"
	"sync"
	"time"
)

var globalLogSyncer struct {
	writeReady chan struct{}
	mu         sync.Mutex
}

func init() {
	globalLogSyncer.writeReady = make(chan struct{})
	close(globalLogSyncer.writeReady)
}

type loggerKey struct{}

// WithLogger creates a copy of the context with a logger attached.
func WithLogger(ctx context.Context, c Codec) context.Context {
	ctx = withDebuggingInitialized(ctx)

	var loggers []Codec
	if v := ctx.Value(loggerKey{}); v != nil {
		loggers = v.([]Codec)
	}
	return context.WithValue(ctx, loggerKey{}, append(loggers, c))
}

type contextTags struct{}

// KV describes a log tag.  When Value is the empty string, this is interpreted
// as being a single tag (with the tag being the Key) rather than a tag pair.
type KV struct {
	Key, Value string
}

// WithLogTag creates a copy of the context with a logging tag.  All calls to
// Log and Debug will include this tag.
func WithLogTag(ctx context.Context, tag string) context.Context {
	var tags []KV
	if v := ctx.Value(contextTags{}); v != nil {
		tags = v.([]KV)
	}
	return context.WithValue(ctx, contextTags{}, append(tags, KV{Key: tag}))
}

// WithLogTagPair creates a copy of the context with a logging tag key/value
// pair.  All calls to Log and Debug will include this tag pair.
func WithLogTagPair(ctx context.Context, k, v string) context.Context {
	var tags []KV
	if v := ctx.Value(contextTags{}); v != nil {
		tags = v.([]KV)
	}
	return context.WithValue(ctx, contextTags{}, append(tags, KV{k, v}))
}

// Codec is used to encode a log entry and write it to an underlying writer.
type Codec interface {
	// EncodeLogEntry encodes a log entry and writes it to an underlying writer.
	//
	// EncodeLogEntry may be called concurrently and the underlying writer must
	// be safe for concurrent writes (or the codec must perform this
	// synchronization itself).
	//
	// This method is designed to support asynchronous logging.  Once all
	// encoding has finished (and none of the data values will be read again),
	// encodeDone must be called.  This signals to the Log function that it is
	// safe to return since there is no posibility of data racing on mutable
	// data passed to Log.
	//
	// To prevent out-of-order written log entries, the write of the encoded
	// entry must only be written once a read of the writeReady channel
	// unblocks.
	EncodeLogEntry(t time.Time, tags []KV, message string, data []Data, encodeDone func(), writeReady <-chan struct{})
}

// Log logs to all attached loggers of the context.  The message parameter
// describes what event or situation is being logged, and additional values of
// importance can be passed to the data slice.
//
// Logging is performed asynchronously, but no references to the additional data
// are held after Log returns.  Instead, all log entries are completely encoded
// before returning, and the writing of the encoded entry to Codec's underlying
// writers is done in the background.  This prevents the possibility of inducing
// a data race by trying to log a mutable parameter.
func Log(ctx context.Context, message string, data ...Data) {
	var loggers []Codec
	if v := ctx.Value(loggerKey{}); v != nil {
		loggers = v.([]Codec)
	} else {
		return
	}

	var tags []KV
	if v := ctx.Value(contextTags{}); v != nil {
		tags = v.([]KV)
	}

	// to prevent entries from showing out of order depending on how long they
	// took to encode, block the write until the previous (if any) has finished.
	// The encoding operation itself is not blocked at all.
	globalLogSyncer.mu.Lock()
	writeReady := globalLogSyncer.writeReady
	nextWriteReady := make(chan struct{})
	globalLogSyncer.writeReady = nextWriteReady
	// Getting the current time must be done before any other calls to Log add
	// other writers, or log timestamps may appear out of order, even though the
	// logs messages themselves are ordered correctly.
	t := time.Now()
	globalLogSyncer.mu.Unlock()

	var writesDone, encodesDone sync.WaitGroup
	writesDone.Add(len(loggers))
	encodesDone.Add(len(loggers))
	for _, c := range loggers {
		go func(c Codec) {
			c.EncodeLogEntry(t, tags, message, data, encodesDone.Done, writeReady)
			writesDone.Done()
		}(c)
	}
	go func() {
		writesDone.Wait()
		close(nextWriteReady)
	}()

	// Only safe to return to caller once all encoding has been completed, even
	// if the entry has not yet been written to the underlying writer.  This
	// allows for async logging without fear of holding references to mutable
	// data and causing a data race.
	encodesDone.Wait()
}

// Debug is a debugging log function that adds an extra "debug" log tag to each
// log entry.  Debugging is not turned on by default but can be enabled at
// runtime either per-context or globally (see SetDebuggingEnabled and
// SetGlobalDebuggingEnabled).
//
// If this package was built with the "release" build tag, all debugging is
// turned off and is not included in the generated code.
func Debug(ctx context.Context, message string, data ...Data) {
	debug(ctx, message, data...)
}

// SetDebuggingEnabled enables or disables per-context debug logging in
// non-release builds.  If global logging is enabled (see
// SetGlobalDebuggingEnabled) all debug logs to this context's loggers are
// enabled regardless of this value.
//
// It is an intentional design decision to not allow for querying whether
// compile time or runtime debugging is enabled or not, since the presense or
// lack of debugging should not affect other systems of the program.
func SetDebuggingEnabled(ctx context.Context, enabled bool) {
	setDebuggingEnabled(ctx, enabled)
}

// SetGlobalDebuggingEnabled enables or disables all debug output in non-release
// builds.  If per-context debugging is enabled, debug log entries will still be
// created even if global debugging is disabled.
//
// It is an intentional design decision to not allow for querying whether
// compile time or runtime debugging is enabled or not, since the presense or
// lack of debugging should not affect other systems of the program.
func SetGlobalDebuggingEnabled(enabled bool) {
	setGlobalDebuggingEnabled(enabled)
}

// Sync blocks until all loggers have finished writing all log entries created
// up to now.  Note that does not also block on any concurrent logs started
// after Sync is called.
//
// Sync should be called before flushing each codec's underlying writer to
// ensure that all log entries created before now are written.
func Sync() {
	globalLogSyncer.mu.Lock()
	writesDone := globalLogSyncer.writeReady
	globalLogSyncer.mu.Unlock()
	<-writesDone
}
