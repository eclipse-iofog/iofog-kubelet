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

package vkubelet

import (
	"net/http"

	"github.com/eclipse-iofog/iofog-kubelet/log"
	"github.com/eclipse-iofog/iofog-kubelet/vkubelet/api"
	"github.com/gorilla/mux"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
)

// ServeMux defines an interface used to attach routes to an existing http
// serve mux.
// It is used to enable callers creating a new server to completely manage
// their own HTTP server while allowing us to attach the required routes to
// satisfy the Kubelet HTTP interfaces.
type ServeMux interface {
	Handle(path string, h http.Handler)
}

func instrumentRequest(r *http.Request) *http.Request {
	ctx := r.Context()
	logger := log.G(ctx).WithFields(log.Fields{
		"uri":  r.RequestURI,
		"vars": mux.Vars(r),
	})
	ctx = log.WithLogger(ctx, logger)

	return r.WithContext(ctx)
}

// InstrumentHandler wraps an http.Handler and injects instrumentation into the request context.
func InstrumentHandler(h http.Handler) http.Handler {
	instrumented := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req = instrumentRequest(req)
		h.ServeHTTP(w, req)
	})
	return &ochttp.Handler{
		Handler:     instrumented,
		Propagation: &b3.HTTPFormat{},
	}
}

// NotFound provides a handler for cases where the requested endpoint doesn't exist
func NotFound(w http.ResponseWriter, r *http.Request) {
	log.G(r.Context()).Debug("404 request not found")
	http.Error(w, "404 request not found", http.StatusNotFound)
}

// NotImplemented provides a handler for cases where a provider does not implement a given API
func NotImplemented(w http.ResponseWriter, r *http.Request) {
	log.G(r.Context()).Debug("501 not implemented")
	http.Error(w, "501 not implemented", http.StatusNotImplemented)
}

func AttachFogControllerRoutes(mux ServeMux, startFunc func(nodeId string), stopFunc func(nodeId string, deleteNode bool)) {
	mux.Handle("/", InstrumentHandler(FogControllerHandler(startFunc, stopFunc)))
}

func FogControllerHandler(startFunc func(nodeId string), stopFunc func(nodeId string, deleteNode bool)) http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/node", api.FogControllerHandlerStopFunc(stopFunc)).Queries("uuid", "{uuid}").Methods("DELETE")
	r.HandleFunc("/node", api.FogControllerHandlerStartFunc(startFunc)).Queries("uuid", "{uuid}").Methods("POST")

	r.NotFoundHandler = http.HandlerFunc(NotFound)
	return r
}
