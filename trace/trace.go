/*
 *  *******************************************************************************
 *  * Copyright (c) 2019 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package trace

import (
	"context"

	"github.com/eclipse-iofog/iofog-kubelet/v2/log"
	"go.opencensus.io/trace"
)

// Status is an alias to opencensus's trace status.
// The main reason we use this instead of implementing our own is library re-use,
// namely for converting an error to a tracing status.
// In the future this may be defined completely in this package.
type Status = trace.Status

// Tracer is the interface used for creating a tracing span
type Tracer interface {
	// StartSpan starts a new span. The span details are emebedded into the returned
	// context
	StartSpan(context.Context, string) (context.Context, Span)
}

var (
	// T is the Tracer to use this should be initialized before starting up
	// iofog-kubelet
	T Tracer = nopTracer{}
)

// StartSpan starts a span from the configured default tracer
func StartSpan(ctx context.Context, name string) (context.Context, Span) {
	ctx, span := T.StartSpan(ctx, name)
	ctx = log.WithLogger(ctx, span.Logger())
	return ctx, span
}

// Span encapsulates a tracing event
type Span interface {
	End()
	SetStatus(Status)

	// WithField and WithFields adds attributes to an entire span
	//
	// This interface is a bit weird, but allows us to manage loggers in the context
	// It is expected that implementations set `log.WithLogger` so the logger stored
	// in the context is updated with the new fields.
	WithField(context.Context, string, interface{}) context.Context
	WithFields(context.Context, log.Fields) context.Context

	// Logger is used to log individual entries.
	// Calls to functions like `WithField` and `WithFields` on the logger should
	// not affect the rest of the span but rather individual entries.
	Logger() log.Logger
}
