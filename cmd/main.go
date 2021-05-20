// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 Dell Technologies
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/edgexfoundry/device-camera-go"
	"github.com/edgexfoundry/device-camera-go/internal/driver"
	"github.com/edgexfoundry/device-sdk-go/pkg/startup"
)

const (
	serviceName string = "device-camera"
)

func main() {
	sd := driver.NewProtocolDriver()
	startup.Bootstrap(serviceName, device_camera.Version, sd)
}
