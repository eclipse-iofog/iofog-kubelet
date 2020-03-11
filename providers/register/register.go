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

package register

import (
	"github.com/cpuguy83/strongerrors"
	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/apps"
	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofog-kubelet/v2/manager"
	"github.com/eclipse-iofog/iofog-kubelet/v2/providers"
	"github.com/eclipse-iofog/iofog-kubelet/v2/vkubelet/api"
	"github.com/pkg/errors"
)

var providerInits = make(map[string]initFunc)

// InitConfig is the config passed to initialize a registered provider.
type InitConfig struct {
	ConfigPath       string
	NodeName         string
	OperatingSystem  string
	InternalIP       string
	DaemonPort       int32
	ResourceManager  *manager.ResourceManager
	Controller       apps.IofogController
	ControllerClient *client.Client
	NodeId           string
	Store            *api.KeyValueStore
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
