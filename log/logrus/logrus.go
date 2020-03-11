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

package logrus

import (
	"github.com/Sirupsen/logrus"
	"github.com/eclipse-iofog/iofog-kubelet/v2/log"
)

// Adapter implements the `log.Logger` interface for logrus
type Adapter struct {
	*logrus.Entry
}

// FromLogrus creates a new `log.Logger` from the provided entry
func FromLogrus(entry *logrus.Entry) log.Logger {
	return &Adapter{entry}
}

// WithField adds a field to the log entry.
func (l *Adapter) WithField(key string, val interface{}) log.Logger {
	return FromLogrus(l.Entry.WithField(key, val))
}

// WithFields adds multiple fields to a log entry.
func (l *Adapter) WithFields(f log.Fields) log.Logger {
	return FromLogrus(l.Entry.WithFields(logrus.Fields(f)))
}

// WithError adds an error to the log entry
func (l *Adapter) WithError(err error) log.Logger {
	return FromLogrus(l.Entry.WithError(err))
}
