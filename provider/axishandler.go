// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

// Package cameradiscoveryprovider implements a CameraDiscovery provider which scans
// network for existence of compatible/requested IP cameras and registers these
// as EdgeX devices, with corresponding management/control interfaces.
package cameradiscoveryprovider

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func (p *CameraDiscoveryProvider) getAxisCameraDetails(address string, credentialsPath string) (CameraInfo, error) {
	p.lc.Debug(fmt.Sprintf("Querying AXIS details from: %v", address))
	client := &http.Client{}
	parameters := []string{
		"Brand.ProdFullName",
		"Properties.Firmware.Version",
		"Properties.System.SerialNumber",
		"Properties.Image.Resolution",
		"Properties.Image.Format",
		"Properties.API.RTSP.RTSPAuth",
	}
	url := "http://" + address + "/axis-cgi/admin/param.cgi?action=list&group=" + strings.Join(parameters, ",")
	// Providing credentials will cause Axis API to produce errors in some camera configurations.
	// Adjustment is needed for cases where cameras a configured for basic vs digest authentication.
	/* cameraUser, cameraPassword, err := readCredentialsFromFile(credentialsPath)
	if err != nil {
		return CameraInfo{}, err
	} */
	// Nullify credentials here
	cameraUser := ""
	cameraPassword := ""
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return CameraInfo{}, err
	}
	if cameraUser != "" && cameraPassword != "" {
		req.SetBasicAuth(cameraUser, cameraPassword)
	}
	resp, err := client.Do(req)
	if err != nil {
		return CameraInfo{}, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return CameraInfo{}, err
	}
	entries := strings.Split(string(body), "\n")
	entries = entries[:len(entries)-1]
	if len(entries) != len(parameters) {
		err := fmt.Errorf("Invalid answer from camera API for host: %v", address)
		return CameraInfo{}, err
	}
	for i := range entries {
		result := strings.Split(entries[i], "=")
		if len(result) > 1 {
			entries[i] = result[1]
		} else {
			err := fmt.Errorf("Invalid data structure returned from camera API for host: %v", address)
			return CameraInfo{}, err
		}
	}
	defaultResolution := strings.Split(entries[3], ",")
	cameraInfo := CameraInfo{
		IPAddress:       address,
		ProductName:     entries[0],
		FirmwareVersion: entries[1],
		SerialNumber:    entries[2],
		VendorInfoAxis: VendorInfoAxis{
			SupportedResolutions: defaultResolution,
			SupportedFormats:     strings.Split(entries[4], ","),
			RTSPAuthEnabled:      parseAxisBoolean(entries[5]),
			RTSPPath:             "rtsp://" + address + "/axis-media/media.amp",
			ImagePath:            "http://" + address + "/jpg/1/image.jpg?resolution=" + defaultResolution[len(defaultResolution)-1],
		},
	}
	p.lc.Debug(fmt.Sprintf("%v", cameraInfo))
	return cameraInfo, nil
}

func parseAxisBoolean(value string) bool {
	return value == "yes"
}
