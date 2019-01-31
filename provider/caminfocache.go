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
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"path/filepath"
)

// CamInfo is a struct containing info to associate with cameras in a form of
// map where serial number is the key and info map is the value
type CamInfo struct {
	//	CamInfo map[string]map[CameraInfo]interface{}
	OnvifCameras []CameraInfo `json:"onvif-cameras"`
	AxisCameras  []CameraInfo `json:"axis-cameras"`
	// ... additional vendors
}

// LoadInfo loads camera info from a json file
func (st *CamInfo) LoadInfo(lc AppLoggingClient, filePathOnvif string, filePathAxis string) error {
	if filePathOnvif != "" {
		bytes, err := ioutil.ReadFile(filepath.Clean(filePathOnvif))
		if err != nil {
			return errors.Wrap(err, "Loading file failed for ONVIF cameras. Using empty camerainfo cache.")
		}
		err = json.Unmarshal(bytes, &(st.OnvifCameras))
		if err != nil {
			return errors.Wrap(err, "Parsing JSON failed for ONVIF cameras. Using empty camerainfo cache.")
		}
		lc.Info(fmt.Sprintf("*** Loaded %d ONVIF devices from CameraInfoCache... ***", len(st.OnvifCameras)))
	}
	if filePathAxis != "" {
		bytes, err := ioutil.ReadFile(filepath.Clean(filePathAxis))
		if err != nil {
			return errors.Wrap(err, "Loading file failed for Axis cameras. Using empty camerainfo cache.")
		}
		err = json.Unmarshal(bytes, &(st.AxisCameras))
		if err != nil {
			return errors.Wrap(err, "Parsing JSON failed for Axis cameras. Using empty camerainfo cache.")
		}
		lc.Info(fmt.Sprintf("*** Loaded %d Axis devices from CameraInfoCache... ***", len(st.AxisCameras)))
	}
	return nil
}

// SaveInfo flushes cached camera info to json file
// if target pathfile parameter is supplied.
func (st *CamInfo) SaveInfo(filePathOnvif string, filePathAxis string) error {
	if filePathOnvif != "" {
		infoJSON, err := json.Marshal(&st.OnvifCameras)
		if err != nil {
			return errors.Wrap(err, "Saving ONVIF Camera Info as JSON failed. Using in-memory cache only!")
		}
		err = ioutil.WriteFile(filePathOnvif, infoJSON, 0600)
		if err != nil {
			return errors.Wrap(err, "Saving ONVIF Camera Info file failed. Using in-memory cache only!")
		}
	}
	if filePathAxis != "" {
		infoJSON, err := json.Marshal(&st.AxisCameras)
		if err != nil {
			return errors.Wrap(err, "Saving Axis Camera Info as JSON failed. Using in-memory cache only!")
		}
		err = ioutil.WriteFile(filePathAxis, infoJSON, 0600)
		if err != nil {
			return errors.Wrap(err, "Saving Axis Camera Info file failed. Using in-memory cache only!")
		}
	}
	return nil
}

// TransformCameraInfoToString marshals structured JSON data to string
func (st *CamInfo) TransformCameraInfoToString(deviceVendor string, serialNum string) string {
	var ci []CameraInfo
	if deviceVendor == "axis" {
		ci = st.AxisCameras
	}
	if deviceVendor == "onvif" {
		ci = st.OnvifCameras
	}
	for i := range ci {
		if ci[i].SerialNumber == serialNum {
			b, err := json.Marshal(ci[i])
			if err != nil {
				fmt.Println("Error marshaling camera info from index: ", i, " from CameraInfoCache for SerialNumber: ", serialNum)
			} else {
				fmt.Println("Fetched CameraInfo from CamInfoCache: ", string(b))
				return string(b)
			}
		}
	}
	return "[camera info not found!]"
}

func (st *CamInfo) _appendIfMissing(lc AppLoggingClient, cameras []CameraInfo, item CameraInfo) []CameraInfo {
	for _, ele := range cameras {
		if ele.SerialNumber == item.SerialNumber {
			lc.Debug(fmt.Sprintf("Found EXISTING camera in cache with serialNum: %v", item.SerialNumber))
			return cameras
		}
	}
	lc.Debug(fmt.Sprintf("Adding NEW camera to cache with serialNum: %v", item.SerialNumber))
	return append(cameras, item)
}

// AddOnvifCamera appends camera device to Onvif cache if not existing - cache invalidation tbd
func (st *CamInfo) AddOnvifCamera(lc AppLoggingClient, item CameraInfo) []CameraInfo {
	st.OnvifCameras = st._appendIfMissing(lc, st.OnvifCameras, item)
	return st.OnvifCameras
}

// AddAxisCamera appends camera device to Axis cache if not existing - cache invalidation tbd
func (st *CamInfo) AddAxisCamera(lc AppLoggingClient, item CameraInfo) []CameraInfo {
	st.AxisCameras = st._appendIfMissing(lc, st.AxisCameras, item)
	return st.AxisCameras
}
