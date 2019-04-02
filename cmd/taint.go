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

package cmd

import (
	"os"

	corev1 "k8s.io/api/core/v1"
)

// Default taint values
const (
	DefaultTaintEffect = corev1.TaintEffectNoSchedule
	DefaultTaintKey    = "resource-type"
	DefaultTaintValue  = "iofog-custom-resource"
)

func getEnv(key, defaultValue string) string {
	value, found := os.LookupEnv(key)
	if found {
		return value
	}
	return defaultValue
}

func getTaint() (*corev1.Taint, error) {
	return &corev1.Taint{
		Key:    DefaultTaintKey,
		Value:  DefaultTaintValue,
		Effect: DefaultTaintEffect,
	}, nil
}
