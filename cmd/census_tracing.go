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
	"net/http"
	"os"

	"github.com/cpuguy83/strongerrors"
	"github.com/eclipse-iofog/iofog-kubelet/log"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
	"go.opencensus.io/zpages"
)

var (
	tracingExporters = make(map[string]TracingExporterInitFunc)

	reservedTagNames = map[string]bool{
		"operatingSystem": true,
		"provider":        true,
		"nodeName":        true,
	}
)

// TracingExporterOptions is used to pass options to the configured tracer
type TracingExporterOptions struct {
	Tags        map[string]string
	ServiceName string
}

// TracingExporterInitFunc is the function that is called to initialize an exporter.
// This is used when registering an exporter and called when a user specifed they want to use the exporter.
type TracingExporterInitFunc func(TracingExporterOptions) (trace.Exporter, error)

// RegisterTracingExporter registers a tracing exporter.
// For a user to select an exporter, it must be registered here.
func RegisterTracingExporter(name string, f TracingExporterInitFunc) {
	tracingExporters[name] = f
}

// GetTracingExporter gets the specified tracing exporter passing in the options to the exporter init function.
// For an exporter to be availbale here it must be registered with `RegisterTracingExporter`.
func GetTracingExporter(name string, opts TracingExporterOptions) (trace.Exporter, error) {
	f, ok := tracingExporters[name]
	if !ok {
		return nil, strongerrors.NotFound(errors.Errorf("tracing exporter %q not found", name))
	}
	return f(opts)
}

// AvailableTraceExporters gets the list of registered exporters
func AvailableTraceExporters() []string {
	out := make([]string, 0, len(tracingExporters))
	for k := range tracingExporters {
		out = append(out, k)
	}
	return out
}

func setupZpages() {
	ctx := context.TODO()
	p := os.Getenv("ZPAGES_PORT")
	if p == "" {
		log.G(ctx).Error("Missing ZPAGES_PORT env var, cannot setup zpages endpoint")
	}
	mux := http.NewServeMux()
	zpages.Handle(mux, "/debug")
	http.ListenAndServe(p, mux)
}
