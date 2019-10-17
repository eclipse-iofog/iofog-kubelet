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
	"github.com/eclipse-iofog/iofog-kubelet/providers"
	"github.com/eclipse-iofog/iofog-kubelet/providers/iofog"
)

func init() {
	register("iofog", initWeb)
}

func initWeb(cfg InitConfig) (providers.Provider, error) {
	return iofog.NewBrokerProvider(
		cfg.DaemonPort,
		cfg.NodeName,
		cfg.OperatingSystem,
		cfg.Controller,
		cfg.ControllerClient,
		cfg.NodeId,
		cfg.Store)
}
