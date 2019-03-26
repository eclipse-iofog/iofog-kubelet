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

package log

import (
	"context"
)

var (
	// G is an alias for GetLogger.
	G = GetLogger

	// L is the default logger. It should be initialized before using `G` or `GetLogger`
	// If L is unitialized and no logger is available in a provided context, a
	// panic will occur.
	L Logger = nopLogger{}
)

type (
	loggerKey struct{}
)

// Logger is the interface used for logging in iofog-kubelet
//
// iofog-kubelet will access the logger via context using `GetLogger` (or its alias, `G`)
// You can set the default logger to use by setting the `L` variable.
type Logger interface {
	Debug(...interface{})
	Debugf(string, ...interface{})
	Info(...interface{})
	Infof(string, ...interface{})
	Warn(...interface{})
	Warnf(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
	Fatal(...interface{})
	Fatalf(string, ...interface{})

	WithField(string, interface{}) Logger
	WithFields(Fields) Logger
	WithError(error) Logger
}

// Fields allows setting multiple fields on a logger at one time.
type Fields map[string]interface{}

// WithLogger returns a new context with the provided logger. Use in
// combination with logger.WithField(s) for great effect.
func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// GetLogger retrieves the current logger from the context. If no logger is
// available, the default logger is returned.
func GetLogger(ctx context.Context) Logger {
	logger := ctx.Value(loggerKey{})

	if logger == nil {
		if L == nil {
			panic("default logger not initialized")
		}
		return L
	}

	return logger.(Logger)
}
