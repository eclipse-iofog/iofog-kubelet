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
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/eclipse-iofog/iofog-kubelet/v2/log"
)

func TestImplementsLoggerInterface(t *testing.T) {
	l := FromLogrus(&logrus.Entry{})

	if _, ok := l.(log.Logger); !ok {
		t.Fatal("does not implement log.Logger interface")
	}
}
