// +build !no_ocagent_exporter

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

	"contrib.go.opencensus.io/exporter/ocagent"
	"github.com/cpuguy83/strongerrors"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

func init() {
	RegisterTracingExporter("ocagent", NewOCAgentExporter)
}

// NewOCAgentExporter creates a new opencensus tracing exporter using the opencensus agent forwarder.
func NewOCAgentExporter(opts TracingExporterOptions) (trace.Exporter, error) {
	agentOpts := append([]ocagent.ExporterOption{}, ocagent.WithServiceName(opts.ServiceName))

	if endpoint := os.Getenv("OCAGENT_ENDPOINT"); endpoint != "" {
		agentOpts = append(agentOpts, ocagent.WithAddress(endpoint))
	} else {
		return nil, strongerrors.InvalidArgument(errors.New("must set endpoint address in OCAGENT_ENDPOINT"))
	}

	switch os.Getenv("OCAGENT_INSECURE") {
	case "0", "no", "n", "off", "":
	case "1", "yes", "y", "on":
		agentOpts = append(agentOpts, ocagent.WithInsecure())
	default:
		return nil, strongerrors.InvalidArgument(errors.New("invalid value for OCAGENT_INSECURE"))
	}

	return ocagent.NewExporter(agentOpts...)
}
