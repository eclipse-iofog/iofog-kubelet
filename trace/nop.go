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

type nopTracer struct{}

func (nopTracer) StartSpan(ctx context.Context, _ string) (context.Context, Span) {
	return ctx, &nopSpan{}
}

type nopSpan struct{}

func (nopSpan) End()                   {}
func (nopSpan) SetStatus(trace.Status) {}
func (nopSpan) Logger() log.Logger     { return nil }

func (nopSpan) WithField(ctx context.Context, _ string, _ interface{}) context.Context { return ctx }
func (nopSpan) WithFields(ctx context.Context, _ log.Fields) context.Context           { return ctx }
