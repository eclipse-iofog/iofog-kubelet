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

type nopLogger struct{}

func (nopLogger) Debug(...interface{})          {}
func (nopLogger) Debugf(string, ...interface{}) {}
func (nopLogger) Info(...interface{})           {}
func (nopLogger) Infof(string, ...interface{})  {}
func (nopLogger) Warn(...interface{})           {}
func (nopLogger) Warnf(string, ...interface{})  {}
func (nopLogger) Error(...interface{})          {}
func (nopLogger) Errorf(string, ...interface{}) {}
func (nopLogger) Fatal(...interface{})          {}
func (nopLogger) Fatalf(string, ...interface{}) {}

func (l nopLogger) WithField(string, interface{}) Logger { return l }
func (l nopLogger) WithFields(Fields) Logger             { return l }
func (l nopLogger) WithError(error) Logger               { return l }
