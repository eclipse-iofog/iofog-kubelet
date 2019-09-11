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
	"context"
	"github.com/eclipse-iofog/iofog-kubelet/log"
	"github.com/eclipse-iofog/iofog-kubelet/vkubelet"
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
