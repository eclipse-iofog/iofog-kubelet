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
	"fmt"

	"github.com/eclipse-iofog/iofog-kubelet/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the version of the program",
	Long:  `Show the version of the program`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s, Built: %s", version.Version, version.BuildTime)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
