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
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/lair-framework/go-nmap"
)

func (p *CameraDiscoveryProvider) nmapDiscover(scanDuration string, ports, ip string, mask string) (map[string][]int, string, error) {
	result := make(map[string][]int)
	startTime := "0"
	p.lc.Info("NMAP IP/Mask: " + ip + mask)
	cmd := exec.Command(
		"nmap", "-T4", "-n", "--open", "-oX", "nmap.xml",
		"--host-timeout", scanDuration, "-p", ports, ip+mask)
	err := cmd.Run()
	if err != nil {
		return nil, startTime, err
	}
	file, err := os.Open("nmap.xml")
	if err != nil {
		return nil, startTime, err
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, startTime, err
	}
	report, err := nmap.Parse(data)
	if err != nil {
		return nil, startTime, err
	}
	startTime = report.StartStr
	for _, host := range report.Hosts {
		ip := host.Addresses[0].Addr
		var discoveredPorts []int
		for _, port := range host.Ports {
			discoveredPorts = append(discoveredPorts, port.PortId)
		}
		result[ip] = discoveredPorts
	}
	return result, startTime, nil
}
