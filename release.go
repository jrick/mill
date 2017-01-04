// Copyright (c) 2017 Josh Rickmar
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

//+build release

// This file contains noop implementations for all debugging mechanisms that are
// removed from release builds.

package mill

import (
	"context"
)

func debug(ctx context.Context, message string, data ...Data) {}

func setDebuggingEnabled(ctx context.Context, enabled bool) {}

func debuggingEnabled(ctx context.Context) bool { return false }

func withDebuggingInitialized(ctx context.Context) context.Context { return ctx }

func setGlobalDebuggingEnabled(enabled bool) {}
