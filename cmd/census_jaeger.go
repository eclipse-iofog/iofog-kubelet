// +build !no_jaeger_exporter

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
	"errors"
	"os"

	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/trace"
)

func init() {
	RegisterTracingExporter("jaeger", NewJaegerExporter)
}

// NewJaegerExporter creates a new opencensus tracing exporter.
func NewJaegerExporter(opts TracingExporterOptions) (trace.Exporter, error) {
	jOpts := jaeger.Options{
		Endpoint:      os.Getenv("JAEGER_ENDPOINT"),
		AgentEndpoint: os.Getenv("JAEGER_AGENT_ENDPOINT"),
		Username:      os.Getenv("JAEGER_USER"),
		Password:      os.Getenv("JAEGER_PASSWORD"),
		Process: jaeger.Process{
			ServiceName: opts.ServiceName,
		},
	}

	if jOpts.Endpoint == "" && jOpts.AgentEndpoint == "" {
		return nil, errors.New("Must specify either JAEGER_ENDPOINT or JAEGER_AGENT_ENDPOINT")
	}

	for k, v := range opts.Tags {
		jOpts.Process.Tags = append(jOpts.Process.Tags, jaeger.StringTag(k, v))
	}
	return jaeger.NewExporter(jOpts)
}
