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

package providers

const (
	// OperatingSystemLinux is the configuration value for defining Linux.
	OperatingSystemLinux = "Linux"
	// OperatingSystemWindows is the configuration value for defining Windows.
	OperatingSystemWindows = "Windows"
)

type OperatingSystems map[string]bool

var (
	// ValidOperatingSystems defines the group of operating systems
	// that can be used as a kubelet node.
	ValidOperatingSystems = OperatingSystems{
		OperatingSystemLinux:   true,
		OperatingSystemWindows: true,
	}
)

func (o OperatingSystems) Names() []string {
	keys := make([]string, 0, len(o))
	for k := range o {
		keys = append(keys, k)
	}
	return keys
}
