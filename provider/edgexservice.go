// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

// Package cameradiscoveryprovider implements a CameraDiscovery provider which scans
// network for existence of compatible/requested IP cameras and registers these
// as EdgeX devices, with corresponding management/control interfaces.
package cameradiscoveryprovider

// Implements EdgeX ProtocolDriver interface
import (
	"fmt"
	"github.com/edgexfoundry/device-sdk-go"
	ds_models "github.com/edgexfoundry/device-sdk-go/pkg/models"
	logger "github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	e_models "github.com/edgexfoundry/edgex-go/pkg/models"
	"strings"
	"time"
)

const (
	// CameraManagementProfileName currently reuses Axis device profile to create a mock Camera Management Provider
	// as a placeholder EdgeX device (i.e., for the device service itself).
	CameraManagementProfileName string = "camera-profile-axis"
)

// AppLoggingClient overrides the implementation of the EdgeX logger.LoggingClient interface.
// It delegates logging to centralize any errors thrown by act of logging itself.
// It will always return nil or panic.
type AppLoggingClient interface {
	SetLogLevel(logLevel string) error
	Debug(msg string, labels ...string)
	Error(msg string, labels ...string)
	Info(msg string, labels ...string)
	Trace(msg string, labels ...string)
	Warn(msg string, labels ...string)
}
type appLogger struct {
	lc        logger.LoggingClient
	msgPrefix string
	count     int
	thresh    int
}

func newAppLogger(lc logger.LoggingClient) AppLoggingClient {
	return &appLogger{lc: lc, count: 0, thresh: 0, msgPrefix: fmt.Sprintf("[%s] - ", device.RunningService().Name())}
}

func (l *appLogger) SetLogLevel(level string) error {
	return l.lc.SetLogLevel(level)
}
func (l *appLogger) Error(msg string, labels ...string) {
	if err := l.lc.Error(l.msgPrefix+msg, labels...); err != nil {
		l.count++
		panic(err)
	}
}

func (l *appLogger) Warn(msg string, labels ...string) {
	if err := l.lc.Warn(l.msgPrefix+msg, labels...); err != nil {
		l.count++
		panic(err)
	}
}

func (l *appLogger) Info(msg string, labels ...string) {
	if err := l.lc.Info(l.msgPrefix+msg, labels...); err != nil {
		l.count++
		panic(err)
	}
}

func (l *appLogger) Debug(msg string, labels ...string) {
	if err := l.lc.Debug(l.msgPrefix+msg, labels...); err != nil {
		l.count++
		panic(err)
	}
}

func (l *appLogger) Trace(msg string, labels ...string) {
	if err := l.lc.Trace(l.msgPrefix+msg, labels...); err != nil {
		l.count++
		panic(err)
	}
}

// AppCache holds elements related to application layer's device service caches.
// Currently segmented by tags and camera device types (i.e., a separate cache per device profile).
type AppCache struct {
	CamInfoCache  *CamInfo // In Memory device cache for discovered cameras
	InfoFileOnvif string   // File to back onvif device cache
	InfoFileAxis  string   // File to back axis device cache
	TagCache      *Tags    // Tags to be attached to cameras
	TagsFile      string   // File to initialize TagCache
}

// CameraDiscoveryProvider holds service level objects
type CameraDiscoveryProvider struct {
	lc                 AppLoggingClient
	asyncChan          chan<- *ds_models.AsyncValues
	options            *Options
	intervalTicker     *time.Ticker
	scanDurationTicker *time.Ticker
	ac                 *AppCache
}

// New instantiates CameraDiscoveryProvider
func New(options *Options, ac *AppCache) *CameraDiscoveryProvider {
	var p CameraDiscoveryProvider
	p.options = options
	p.ac = ac
	return &p
}

// DisconnectDevice is called by the SDK for protocol specific disconnection from device service.
func (p *CameraDiscoveryProvider) DisconnectDevice(address *e_models.Addressable) error {
	p.lc.Warn(fmt.Sprintf("DisconnectDevice CALLED: We can set state of devices, and update CoreMetadata..."))
	return nil
}

// Initialize performs protocol-specific initialization for the device
// service. The given *AsyncValues channel can be used to push asynchronous
// events and readings to EdgeX Core Data.
func (p *CameraDiscoveryProvider) Initialize(lc logger.LoggingClient, asyncCh chan<- *ds_models.AsyncValues) error {
	p.lc = newAppLogger(lc)
	p.asyncChan = asyncCh
	p.lc.Trace(fmt.Sprintf("CameraDiscoveryProvider Initialize called with options: %v", p.options))
	// ==============================
	// Validate and normalize inputs
	// ==============================
	// ScanDuration and Interval
	duration, err := time.ParseDuration(p.options.ScanDuration)
	if err != nil {
		p.lc.Error(fmt.Sprintf("Invalid ScanDuration. See help for examples."))
		return err
	}
	minWaitSeconds := 10
	if p.options.Interval <= int(duration.Seconds())+minWaitSeconds {
		err = fmt.Errorf("Must provide more than %d seconds between discovery scans!  Interval[%d] > ScanDuration[%v]", minWaitSeconds, p.options.Interval, duration.Seconds())
		return err
	}
	// IP and Mask - prepend '/'' to netmask if any value is specified
	if len(p.options.NetMask) > 0 {
		if p.options.NetMask[0] != '/' {
			p.options.NetMask = "/" + p.options.NetMask
		}
	}
	// PortList to scan; fed from supplied SourceFlags parameter(s)
	var portList string
	portList, err = p.buildPortList(*p.options)
	if err != nil {
		p.lc.Error(err.Error())
		return err
	}
	// ==============================
	// Load Camera Info cache(s)
	// ==============================
	err = p.loadCameraInfoCaches()
	if err != nil {
		p.lc.Warn(err.Error())
		// Existence of a camera info cache is not mandatory, continue with Initialize
		err = nil
	}
	// ==============================
	// Load Camera Tags cache
	// ==============================
	err = p.ac.TagCache.LoadTags(p.ac.TagsFile)
	if err != nil {
		p.lc.Warn(err.Error())
		// Existence of a tag cache is not mandatory, continue with Initialize
		err = nil
	}
	// ==============================
	// Add service as EdgeX Device
	// ==============================
	labels := []string{"cameradiscovery"}
	deviceName := "camera-management-provider"
	err = p.registerDeviceManagementProvider(deviceName, labels)
	// ==============================
	// Schedule discovery scans
	// ==============================
	if err == nil {
		// Kick off our requested Discovery Schedule
		p.schedulePortScans(portList)
	}
	return err
}

// schedulePortScans declares a goroutine to initiate scans at requested interval
func (p *CameraDiscoveryProvider) schedulePortScans(portList string) {
	p.intervalTicker = time.NewTicker(time.Second * time.Duration(p.options.Interval))
	intervalCount := 0
	intervalStart := time.Now()
	go func() {
		for ; scanOnStartup; <-p.intervalTicker.C {
			intervalCount++
			deviceCount, err := p.DiscoverDevices(portList, *p.options)
			if err != nil {
				p.lc.Error(err.Error())
			}
			time.Sleep(1 * time.Second) // permits device sdk log entries to settle (affects presentation only)
			p.lc.Info(fmt.Sprintf("%v new camera devices registered during this scan", deviceCount))
			// Report next anticipated interval trigger
			nextScan := intervalStart.Local().Add(time.Second * time.Duration(p.options.Interval*intervalCount))
			remainSec := time.Until(nextScan)
			p.lc.Info(fmt.Sprintf("Next scan triggers in %.f seconds (%v), and each %v seconds thereafter.", remainSec.Seconds(), nextScan.Format(time.Stamp), p.options.Interval))
		}
	}()
}

func (p *CameraDiscoveryProvider) loadCameraInfoCaches() error {
	var fileOnvif string
	var fileAxis string
	if p.isVendorRequested(p.options.SourceFlags, "onvif") {
		fileOnvif = p.ac.InfoFileOnvif
	}
	if p.isVendorRequested(p.options.SourceFlags, "axis") {
		fileAxis = p.ac.InfoFileAxis
	}
	return p.ac.CamInfoCache.LoadInfo(p.lc, fileOnvif, fileAxis)
}

func (p *CameraDiscoveryProvider) registerDeviceManagementProvider(deviceName string, labels []string) error {
	p.lc.Info(fmt.Sprintf("Adding CameraDeviceProvider as a proxy/manager EdgeX device"))
	// device.RunningService().RemoveDeviceByName(deviceName)
	edgexDevice, err := device.RunningService().GetDeviceByName(deviceName)
	if err != nil {
		idstr, err2 := device.RunningService().AddDevice(e_models.Device{
			Name:           deviceName,
			AdminState:     "unlocked",
			OperatingState: "enabled",
			Addressable: e_models.Addressable{
				Name:      "device-camera-go-management-provider",
				Protocol:  "HTTP",
				Address:   "localhost", //10.0.2.15  localhost  172.17.0.1
				Port:      49990,
				Path:      "/api/v1/devices/{deviceId}/tags",
				Publisher: "none",
				User:      "none",
				Password:  "none",
				Topic:     "none",
			},
			Labels:   labels,
			Location: "gateway",
			Profile: e_models.DeviceProfile{
				Name: CameraManagementProfileName,
			},
			Service: e_models.DeviceService{
				AdminState: "unlocked",
				Service: e_models.Service{
					Name:           "device-camera-go",
					OperatingState: "enabled",
				},
			},
		})
		err = err2
		if err2 != nil {
			p.lc.Error("Error registering CameraDiscoveryProvider as EdgeX device: " + err2.Error())
		}
		// Upon success, edgex-core-metadata should also respond with a corresponding log message similar to:
		// INFO: 2018/11/14 18:30:46 AddDevice returned ID:5bec69d69f8fc20001fd3a6b
		time.Sleep(1 * time.Second)
		p.lc.Info("CameraDiscoveryProvider assigned EdgeX ID:" + idstr)
	} else {
		p.lc.Info(fmt.Sprintf("CameraDiscoveryProvider was previously registered and has EdgeX ID: %s", edgexDevice.Id.Hex()))
	}
	return err
}

// HandleReadCommands processes GET commands
func (p *CameraDiscoveryProvider) HandleReadCommands(addr *e_models.Addressable,
	reqs []ds_models.CommandRequest) (res []*ds_models.CommandValue, err error) {
	if len(reqs) != 1 {
		p.lc.Error(fmt.Sprintf("CameraDiscoveryDriver.HandleReadCommands; too many command requests; only one supported"))
		return
	}
	p.lc.Info(fmt.Sprintf("RECEIVED COMMAND REQUEST: %s", reqs[0].RO.Object))
	if reqs[0].RO.Object == "onvif_profiles" || reqs[0].RO.Object == "axis_info" {
		// These EdgeX commands are distinct for each device class.
		// ONVIF and vendor-specific APIs (e.g., Axis) provide different interfaces, often to the same physical device.
		var serialNum string
		var camInfo string
		if reqs[0].RO.Object == "axis_info" {
			serialNum = strings.TrimPrefix(addr.Name, p.options.SupportedSources[Axis].DeviceNamePrefix)
			camInfo = p.ac.CamInfoCache.TransformCameraInfoToString("axis", serialNum)
		}
		if reqs[0].RO.Object == "onvif_profiles" {
			serialNum = strings.TrimPrefix(addr.Name, p.options.SupportedSources[ONVIF].DeviceNamePrefix)
			camInfo = p.ac.CamInfoCache.TransformCameraInfoToString("onvif", serialNum)
		}
		res = make([]*ds_models.CommandValue, 1)
		now := time.Now().UnixNano() / int64(time.Millisecond)
		cv := ds_models.NewStringValue(&reqs[0].RO, now, camInfo)
		res[0] = cv
	} else if reqs[0].RO.Object == "tags" {
		// This EdgeX Command is common between two device classes (ONVIF and Axis)
		p.lc.Info(fmt.Sprintf("CameraDiscoveryProvider.HandleReadCommands: Returning Tags associated with device: %s", addr.Name))
		serialNum := strings.TrimPrefix(addr.Name, p.options.SupportedSources[ONVIF].DeviceNamePrefix)
		if len(serialNum) == len(addr.Name) {
			serialNum = strings.TrimPrefix(addr.Name, p.options.SupportedSources[Axis].DeviceNamePrefix)
		}
		camTags := createKeyValuePairString(p.ac.TagCache.Tags[serialNum])
		res = make([]*ds_models.CommandValue, 1)
		now := time.Now().UnixNano() / int64(time.Millisecond)
		cv := ds_models.NewStringValue(&reqs[0].RO, now, camTags)
		res[0] = cv
	} else if reqs[0].RO.Object == "get_user" {
		// Vendor specific command (Axis user CRUD example)
		p.lc.Info(fmt.Sprintf("CameraDiscoveryProvider.HandleReadCommands: TODO: Return EdgeX Video Users associated with device: %s", addr.Name))
	}
	return
}

// HandleWriteCommands processes PUT commands, and is passed a slice of CommandRequest struct
// each representing a ResourceOperation for a specific device resource (aka DeviceObject).
// As these are actuation commands, params will provide parameters distinct to the command.
func (p *CameraDiscoveryProvider) HandleWriteCommands(addr *e_models.Addressable, reqs []ds_models.CommandRequest,
	params []*ds_models.CommandValue) error {
	if len(reqs) != 1 {
		err := fmt.Errorf("CameraDiscoveryDriver.HandleWriteCommands; too many command requests; only one supported")
		return err
	}
	p.lc.Info(fmt.Sprintf("TODO: CameraDiscoveryDriver.HandleWriteCommands: dev: %s op: %v attrs: %v", addr.Name, reqs[0].RO.Operation, reqs[0].DeviceObject.Attributes))
	p.lc.Info(fmt.Sprintf("with params: %v", params))
	if reqs[0].RO.Object == "tags" {
		p.lc.Info(fmt.Sprintf("CameraDiscoveryProvider.HandleWriteCommands: TODO: PUT tags caller wants associated with device: %s", addr.Name))
	} else if reqs[0].RO.Object == "user" {
		// TODO: To support CRUD add commands for /add_user, /update_user, /remove_user
		p.lc.Info(fmt.Sprintf("CameraDiscoveryProvider.HandleWriteCommands: TODO: PUT user group and credentials for EdgeX Video User that caller wants associated with device: %s", addr.Name))
	}
	return nil
}

//Stop is called on termination of service. Perform any needed cleanup for graceful/forced shutdown here.
func (p *CameraDiscoveryProvider) Stop(force bool) error {
	p.lc.Debug("Stopping intervalTicker")
	p.intervalTicker.Stop()

	p.lc.Debug(fmt.Sprintf("Stop Called: force=%v", force))
	return nil
}
