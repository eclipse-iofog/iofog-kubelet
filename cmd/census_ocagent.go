// +build !no_ocagent_exporter

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
