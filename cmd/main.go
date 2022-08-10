// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 Dell Technologies
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/startup"

	device_camera "github.com/edgexfoundry/device-camera-go"
	"github.com/edgexfoundry/device-camera-go/internal/driver"
)

const (
	serviceName string = "device-camera"
)

// Deprecated - this service is deprecated as of Aug 2022.  It is replaced by two new device services as of the Kamakura release
// See
// ONVIF Camera DS - https://github.com/edgexfoundry/device-onvif-camera
// USB Camera DS - https://github.com/edgexfoundry/device-usb-camera

func main() {
	sd := driver.NewProtocolDriver()
	startup.Bootstrap(serviceName, device_camera.Version, sd)
}
