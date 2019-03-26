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
	"context"
	"github.com/iofog/iofog-kubelet/log"
	"github.com/iofog/iofog-kubelet/vkubelet"
	"github.com/pkg/errors"
	"net"
	"net/http"
)

func serveHTTP(ctx context.Context, s *http.Server, l net.Listener, name string) {
	if err := s.Serve(l); err != nil {
		select {
		case <-ctx.Done():
		default:
			log.G(ctx).WithError(err).Errorf("Error setting up %s http server", name)
		}
	}
	l.Close()
}

func setupControllerServer(ctx context.Context, startFunc func(nodeId string), stopFun func(nodeId string, deleteNode bool)) (*http.Server, error) {
	l, err := net.Listen("tcp", "localhost:1234")
	if err != nil {
		return nil, errors.Wrap(err, "could not setup listener for ioFog controller http server")
	}

	mux := http.NewServeMux()
	vkubelet.AttachFogControllerRoutes(mux, startFunc, stopFun)
	s := &http.Server{
		Handler: mux,
	}
	go serveHTTP(ctx, s, l, "iofog controller")
	return s, nil
}