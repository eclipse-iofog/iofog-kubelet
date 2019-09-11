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

package register

import (
	"github.com/cpuguy83/strongerrors"
	"github.com/eclipse-iofog/iofog-kubelet/manager"
	"github.com/eclipse-iofog/iofog-kubelet/providers"
	"github.com/pkg/errors"
)

var providerInits = make(map[string]initFunc)

// InitConfig is the config passed to initialize a registered provider.
type InitConfig struct {
	ConfigPath      string
	NodeName        string
	OperatingSystem string
	InternalIP      string
	DaemonPort      int32
	ResourceManager *manager.ResourceManager
	ControllerToken string
	ControllerUrl   string
	NodeId          string
}

type initFunc func(InitConfig) (providers.Provider, error)

// GetProvider gets the provider specified by the given name
func GetProvider(name string, cfg InitConfig) (providers.Provider, error) {
	f, ok := providerInits[name]
	if !ok {
		return nil, strongerrors.NotFound(errors.Errorf("provider not found: %s", name))
	}
	return f(cfg)
}

func register(name string, f initFunc) {
	providerInits[name] = f
}
