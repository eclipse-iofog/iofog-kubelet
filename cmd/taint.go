// Copyright © 2017 The virtual-kubelet authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
