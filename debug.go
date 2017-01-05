// Copyright (c) 2017 Josh Rickmar
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

//+build !release

// This file contains implementations for all debugging mechanisms.  These
// implementations are only used for non-release builds (the default).

package mill

import (
	"context"
	"sync"
)

type debugKey struct{}

type debugValue struct {
	enabled bool
	mu      sync.Mutex
}

func (v *debugValue) isEnabled() bool {
	v.mu.Lock()
	enabled := v.enabled
	v.mu.Unlock()
	return enabled
}

func (v *debugValue) setEnabled(enabled bool) {
	v.mu.Lock()
	v.enabled = enabled
	v.mu.Unlock()
}

var globalDebugging debugValue

func setGlobalDebuggingEnabled(enabled bool) {
	globalDebugging.setEnabled(enabled)
}

func withDebuggingInitialized(ctx context.Context) context.Context {
	if ctx.Value(debugKey{}) != nil {
		return ctx
	}
	return context.WithValue(ctx, debugKey{}, &debugValue{enabled: false})
}

func setDebuggingEnabled(ctx context.Context, enabled bool) {
	if v := ctx.Value(debugKey{}); v != nil {
		v.(*debugValue).setEnabled(enabled)
	}
}

func debuggingEnabled(ctx context.Context) bool {
	if v := ctx.Value(debugKey{}); v != nil {
		return v.(*debugValue).isEnabled()
	}
	return false
}

func debug(ctx context.Context, message string, data ...Data) {
	if globalDebugging.isEnabled() || debuggingEnabled(ctx) {
		Log(WithLogTag(ctx, "debug"), message, data...)
	}
}
