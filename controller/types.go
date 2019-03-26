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

package controller

import (
	"context"
	"github.com/iofog/iofog-kubelet/vkubelet"
)

type IOFogsReponse struct {
	Fogs []IOFog `json:"fogs"`
}

type IOFog struct {
	UUID         string  `json:"uuid"`
	Name         string  `json:"name"`
	DaemonStatus string  `json:"daemonStatus"`
	MemoryUsage  float32 `json:"memoryUsage"`
	DiskUsage    float32 `json:"diskUsage"`
	CPUUsage     float32 `json:"cpuUsage"`
	IPAddress    string  `json:"ipAddress"`
	//Location                  string  `json:"location"`
	//GPSMode                   string  `json:"gpsMode"`
	//Latitude                  string  `json:"latitude"`
	//Longitude                 string  `json:"longitude"`
	//Description               string  `json:"description"`
	//Time                      int32   `json:"lastActive"`
	//DaemonOperatingDuration   int32   `json:"daemonOperatingDuration"`
	//DaemonLastStart           int32   `json:"daemonLastStart"`
	//MemoryViolation           bool    `json:"memoryViolation"`
	//DiskViolation             bool    `json:"diskViolation"`
	//CPUViolation              bool    `json:"cpuViolation"`
	//SystemAvailableDisk       int32   `json:"systemAvailableDisk"`
	//SystemAvailableMemory     int32   `json:"systemAvailableMemory"`
	//SystemTotalCpu            int32   `json:"systemTotalCpu"`
	//CatalogItemStatus         string  `json:"catalogItemStatus"`
	//RepositoryCount           int32   `json:"repositoryCount"`
	//RepositoryStatus          string  `json:"repositoryStatus"`
	//SystemTime                int32   `json:"systemTime"`
	//LastStatusTime            int32   `json:"lastStatusTime"`
	//ProcessedMessages         int32   `json:"processedMessages"`
	//CatalogItemMessageCounts  int32   `json:"catalogItemMessageCounts"`
	//MessageSpeed              int32   `json:"messageSpeed"`
	//LastCommandTime           int32   `json:"lastCommandTime"`
	//NetworkInterface          string  `json:"networkInterface"`
	//DockerUrl                 string  `json:"dockerUrl"`
	//DiskLimit                 int32   `json:"diskLimit"`
	//DiskDirectory             string  `json:"diskDirectory"`
	//MemoryLimit               int32   `json:"memoryLimit"`
	//CPULimit                  int32   `json:"cpuLimit"`
	//LogLimit                  int32   `json:"logLimit"`
	//LogDirectory              string  `json:"logDirectory"`
	//BluetoothEnabled          bool    `json:"bluetoothEnabled"`
	//AbstractedHardwareEnabled bool    `json:"abstractedHardwareEnabled"`
	//LogFileCount              int32   `json:"logFileCount"`
	//Version                   string  `json:"version"`
	//IsReadyToUpgrade          bool    `json:"isReadyToUpgrade"`
	//IsReadyToRollback         bool    `json:"isReadyToRollback"`
	//StatusFrequency           int32   `json:"statusFrequency"`
	//ChangeFrequency           int32   `json:"changeFrequency"`
	//DeviceScanFrequency       int32   `json:"deviceScanFrequency"`
	//Tunnel                    string  `json:"tunnel"`
	//WatchdogEnabled           bool    `json:"watchdogEnabled"`
	//CreatedAt                 string  `json:"created_at"`
	//UpdatedAt                 string  `json:"updated_at"`
	//FogTypeId                 int8    `json:"fogTypeId"`
	//UserId                    int32   `json:"userId"`
}

type IOFogKubelet struct {
	KubeletInstance        *vkubelet.Server
	NodeContextCancel      context.CancelFunc
	NodeContext            context.Context
}
