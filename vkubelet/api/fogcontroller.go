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

package api

import (
	"net/http"
)

func FogControllerHandlerStopFunc(stopFunc func(nodeId string, deleteNode bool)) http.HandlerFunc {
	return handleError(func(w http.ResponseWriter, req *http.Request) error {
		nodeId := req.FormValue("uuid")
		go stopFunc(nodeId, true)

		return nil
	})
}

func FogControllerHandlerStartFunc(startFunc func(nodeId string)) http.HandlerFunc {
	return handleError(func(w http.ResponseWriter, req *http.Request) error {
		nodeId := req.FormValue("uuid")
		go startFunc(nodeId)

		return nil
	})
}