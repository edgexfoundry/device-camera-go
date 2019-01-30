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
	"bytes"
	"errors"
	"fmt"
	"github.com/edgexfoundry/device-sdk-go"
	e_models "github.com/edgexfoundry/edgex-go/pkg/models"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	scanOnStartup         = true
	showTerminalCountdown = true
	countdownFreqSeconds  = 1
)

type CameraVendor int
const (
	ONVIF    CameraVendor = 0
	Axis     CameraVendor = 1
)

type CameraSource struct {
	Index            CameraVendor
	Name             string
	DeviceNamePrefix string
	ProfileName      string
	DefaultPort      int
}

// Options contains input parameters and runtime objects to simplify access by device service methods.
// Since Options is a member of CameraDiscoveryProvider, consider  as appropriate.
type Options struct {
	Interval         int      // Interval, in seconds, between device scans
	ScanDuration     string   // Duration to permit for network scan - "40s"  "10m"
	IP               string   // IP subnet(s) to use for network scan
	NetMask          string   // Network mask to use for network scan
	SourceFlags      []string // Array holding vendor(s) and port(s) for scan
	Credentials      string   // Credentials to connect to cameras
	SupportedSources []CameraSource
}

func (p *CameraDiscoveryProvider) _terminalCountdown(doneChan *chan bool, ticker **time.Ticker, activity string, durCountdown string) error {
	*ticker = time.NewTicker(countdownFreqSeconds * time.Second)
	*doneChan = make(chan bool)
	go func(doneChan chan bool) {
		var pct int
		countdownDur, err := time.ParseDuration(durCountdown)
		if err != nil {
			p.lc.Error(fmt.Sprintf("Error parsing duration: %v", err.Error()))
		}
		countdownSecs := countdownDur.Seconds()
		remaining := "remaining"
		progressBar := ""
		divScan := 20
		for {
			if *ticker != nil {
				select {
				case <-(*ticker).C:
					if countdownSecs >= 0 {
						pct = 100 - int(100*(countdownSecs/countdownDur.Seconds()))
					} else {
						remaining = "\033[41moverdue\033[0m"
						pct = 99
					}
					progressBar = strings.Repeat("#", ((divScan/10)*pct)/10)
					fmt.Printf("\033[2K\r[%-20s] %s %d%% complete (%v seconds %s)", progressBar, activity, pct, countdownSecs, remaining)
					countdownSecs -= countdownFreqSeconds
				case <-doneChan:
					if countdownSecs >= 0 {
						remaining = "early"
					}
					progressBar = strings.Repeat("#", divScan)
					fmt.Printf("\033[2K\r[%-20s] %s 100%% complete (%v seconds %s)", progressBar, activity, countdownSecs, remaining)
					fmt.Println()
					return
				}
			}
		}
	}(*doneChan)
	return nil
}

// buildPortList constructs a flattened string of ports across sources/vendors.
// It is used to determine the full set of ports to scan for prospective hosts.
func (p *CameraDiscoveryProvider) buildPortList(options Options) (string, error) {
	var portList string
	p.lc.Info("Building flattened port list for scan...")
	if len(options.SourceFlags) > 0 {
		for i := range options.SourceFlags {
			var port string
			var source string
			if strings.Contains(options.SourceFlags[i], ":") {
				s := strings.Split(options.SourceFlags[i], ":")
				source, port = s[0], s[1]
				if port == "" {
					if source == p.options.SupportedSources[Axis].Name {
						port = strconv.Itoa(p.options.SupportedSources[Axis].DefaultPort)
					}
					if source == p.options.SupportedSources[ONVIF].Name {
						port = strconv.Itoa(p.options.SupportedSources[ONVIF].DefaultPort)
					}
				}
				if port == "" {
					return portList, errors.New("FATAL: valid vendor missing (must be onvif and/or axis, format example '-source onvif:80')")
				}
			} else {
				if strings.Contains(options.SourceFlags[i], p.options.SupportedSources[Axis].Name) {
					port = strconv.Itoa(p.options.SupportedSources[Axis].DefaultPort)
				} else if strings.Contains(options.SourceFlags[i], p.options.SupportedSources[ONVIF].Name) {
					port = strconv.Itoa(p.options.SupportedSources[ONVIF].DefaultPort)
				} else {
					return portList, errors.New("FATAL: valid vendor missing (must be onvif and/or axis, format example '-source onvif:80 -source axis')")
				}
			}
			portList += port + ","
		}
	} else {
		return portList, errors.New("FATAL: valid vendor missing (must be onvif and/or axis, format example '-source onvif:80')")
	}
	return portList, nil
}

func createKeyValuePairString(m map[string]interface{}) string {
	b := new(bytes.Buffer)
	counter := 0
	for k, v := range m {
		counter++
		if counter > 1 {
			if _, err := fmt.Fprintf(b, ","); err != nil {
				log.Fatalf("failed to write to console: %s", err)
			}
		}
		switch v.(type) {
		case float64:
			if _, err := fmt.Fprintf(b, "'%s':%v", k, v); err != nil {
				log.Fatalf("failed to write float64 to console: %s", err)
			}
		case bool:
			if _, err := fmt.Fprintf(b, "'%s':%v", k, v); err != nil {
				log.Fatalf("failed to write bool to console: %s", err)
			}
		case string:
			if _, err := fmt.Fprintf(b, "'%s':'%v'", k, v); err != nil {
				log.Fatalf("failed to write string to console: %s", err)
			}
		}
	}
	return b.String()
}

func (p *CameraDiscoveryProvider) isVendorRequested(sourceFlags []string, vendor string) bool {
	for i := range sourceFlags {
		if strings.Contains(sourceFlags[i], vendor) {
			return true
		}
	}
	return false
}

func (p *CameraDiscoveryProvider) discoverHosts(portList string, options Options) (map[string][]int, error) {
	p.lc.Info(fmt.Sprintf("\033[34mScanning network hosts with %v timeout...\033[0m", options.ScanDuration))
	var doneChan chan bool
	if showTerminalCountdown {
		err := p._terminalCountdown(&doneChan, &(p.scanDurationTicker), "Scanning", options.ScanDuration)
		if err != nil {
			fmt.Println("ERROR: during terminal progress bar" + err.Error())
		}
	}
	discoveredHosts, startTime, err := p.nmapDiscover(options.ScanDuration, portList, options.IP, options.NetMask)
	if showTerminalCountdown {
		p.scanDurationTicker.Stop()
		doneChan <- true
		close(doneChan)
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		p.lc.Error(fmt.Sprintf("For scan initiated at %v, encountered NMAP Discovery error: %v", startTime, err.Error()))
		return nil, err
	}
	p.lc.Debug(fmt.Sprintf("Scan found %d potential hosts", len(discoveredHosts)))
	return discoveredHosts, err
}

// DiscoverDevices is a scheduled action that performs holistic camera device discovery including a scan
// for potential hosts, evaluation of requested compliance with one or more camera interface
// APIs (e.g., ONVIF, Axis), population of memoized in memory cache, registration with EdgeX and
// persistance of cache to disk. Data is segmented by each vendor/interface and device profile.
func (p *CameraDiscoveryProvider) DiscoverDevices(portList string, options Options) (int, error) {
	deviceCountTotal := 0
	discoveredHosts, err := p.discoverHosts(portList, *p.options)
	if err != nil {
		p.lc.Error(err.Error())
		return 0, err
	}
	if len(discoveredHosts) > 0 {
		p.lc.Info(fmt.Sprintf("Scanning %d discovered hosts for requested camera device signatures...", len(discoveredHosts)))
		// First check for cameras compliant with ONVIF standard
		deviceCountOnvif, err := p.addCameraDevices(p.options.SupportedSources[ONVIF].Name, p.options.SupportedSources[ONVIF].ProfileName, discoveredHosts)
		if err != nil {
			p.lc.Error(err.Error())
			return 0, err
		}
		deviceCountTotal += deviceCountOnvif
		// Then facilitate additional camera vendor interfaces such as Axis API
		deviceCountAxis, err2 := p.addCameraDevices(p.options.SupportedSources[Axis].Name, p.options.SupportedSources[Axis].ProfileName, discoveredHosts)
		if err2 != nil {
			p.lc.Error(err2.Error())
			return deviceCountTotal, err2
		}
		deviceCountTotal += deviceCountAxis
	} else {
		p.lc.Debug("No hosts discovered listening on requested ports. Aborted the scan for requested compatible camera devices.")
	}
	return deviceCountTotal, nil
}

// addCameraDevice populates the provided discoveredCameras empty array of CameraInfo structure with metadata about the requested camera vendor's device.
func (p *CameraDiscoveryProvider) addCameraDevices(camVendor string, vendorProfile string, discoveredHosts map[string][]int) (int, error) {
	// Note that CameraInfo is currently a shared structure between ONVIF and Axis types.
	//discoveredCameras := []CameraInfo{}
	deviceCountVendor := 0
	if p.isVendorRequested(p.options.SourceFlags, camVendor) {
		p.lc.Info(fmt.Sprintf("Checking for [%s] camera devices...", camVendor))
		discoveredCameras := p.getCameraMetadata(camVendor, p.options.SourceFlags, discoveredHosts, p.options.Credentials)
		p.lc.Info(fmt.Sprintf("\033[32mDiscovered: %v %s cameras\033[0m", len(discoveredCameras), camVendor))
		p.lc.Debug(fmt.Sprintf("Attaching Tags to discovered %s cameras...", camVendor))
		p.assignTags(discoveredCameras, p.ac.TagCache)
		p.lc.Debug(fmt.Sprintf("Storing %s metadata to cache...", camVendor))
		if camVendor == p.options.SupportedSources[ONVIF].Name {
			p.memoizeCamInfoOnvif(discoveredCameras, p.ac.CamInfoCache)
		} else if camVendor == p.options.SupportedSources[Axis].Name {
			p.memoizeCamInfoAxis(discoveredCameras, p.ac.CamInfoCache)
		}
		p.lc.Debug(fmt.Sprintf("Adding EdgeX %s cameras...", camVendor))
		ids, err := p.addEdgeXCameraDevices(discoveredCameras, vendorProfile)
		if err != nil {
			p.lc.Debug(err.Error())
			return deviceCountVendor, err
		}
		deviceCountVendor += len(ids)
		p.lc.Debug(fmt.Sprintf("Finished adding %d new %s Cameras to EdgeX: %v", len(ids), camVendor, strings.Join(ids, ",")))
		err = p.saveToFile(camVendor)
		if err != nil {
			p.lc.Debug("Error persisting device info: " + err.Error())
			return deviceCountVendor, err
		}
	}
	return deviceCountVendor, nil
}

func (p *CameraDiscoveryProvider) saveToFile(camVendor string) error {
	var err error
	p.lc.Debug(fmt.Sprintf("Saving %s CameraInfo cache to disk", camVendor))
	if camVendor == p.options.SupportedSources[ONVIF].Name {
		err = p.ac.CamInfoCache.SaveInfo(p.ac.InfoFileOnvif, "")
	} else if camVendor == p.options.SupportedSources[Axis].Name {
		err = p.ac.CamInfoCache.SaveInfo("", p.ac.InfoFileAxis)
	}
	if err != nil {
		p.lc.Error(err.Error())
	} else {
		p.lc.Info(fmt.Sprintf("Saved %s CameraInfo cache to disk", camVendor))
	}
	return err
}

func (p *CameraDiscoveryProvider) addEdgeXCameraDevices(cameras []CameraInfo, deviceVendorProfileName string) ([]string, error) {
	var idstr string
	var err error
	var ids []string
	var counter int
	// Add each discovered camera as an EdgeX device, managed by this device service
	deviceNamePrefix := p.options.SupportedSources[ONVIF].DeviceNamePrefix
	if deviceVendorProfileName == p.options.SupportedSources[Axis].ProfileName {
		deviceNamePrefix = p.options.SupportedSources[Axis].DeviceNamePrefix
	}
	for i := range cameras {
		var edgexDevice e_models.Device
		if cameras[i].SerialNumber == "" {
			p.lc.Error(fmt.Sprintf("ERROR adding EdgeX camera device. Check credentials. No serial number at index: %v", i))
		} else {
			// Transfer all Tags to EdgeX device labels...
			labels := make([]string, len(cameras[i].Tags))
			var value string
			counter = 0
			for tagKey, tagVal := range cameras[i].Tags {
				value = fmt.Sprintf("%v", tagVal)
				labels[counter] = tagKey + ":" + value
				p.lc.Debug(fmt.Sprintf("processing tag @ idx: %v", tagKey))
				counter++
			}
			deviceName := deviceNamePrefix + cameras[i].SerialNumber
			edgexDevice, err = device.RunningService().GetDeviceByName(deviceName)
			if err != nil {
				// In this case expect EdgeX error:
				// "Device 'edgex-camera-<interface>-<SerialNumber>' cannot be found in cache"
				p.lc.Info("GetDeviceByName for device " + deviceName + " ErrResponse: " + err.Error())
				edgexDevice = e_models.Device{
					Name:           deviceName,
					AdminState:     "unlocked",
					OperatingState: "enabled",
					Addressable: e_models.Addressable{
						Name:      deviceName + "-addressable",
						Protocol:  "HTTP",
						Address:   "172.17.0.1", //functions equally well using localhost/127.0.0.1, or external ip of host, with Delhi release
						Port:      49990,
						Path:      "/cameradiscoveryprovider",
						Publisher: "none",
						User:      "none",
						Password:  "none",
						Topic:     "none",
					},
					Labels: labels,
					//Location: tag.deviceLocation,
					Profile: e_models.DeviceProfile{
						Name: deviceVendorProfileName,
					},
					Service: e_models.DeviceService{
						AdminState: "unlocked",
						Service: e_models.Service{
							Name:           "device-camera-go",
							OperatingState: "enabled",
						},
					},
				}
				p.lc.Debug(fmt.Sprintf("Adding NEW EdgeX device named: %s", deviceName))
				idstr, err = device.RunningService().AddDevice(edgexDevice)
				if err != nil {
					p.lc.Error(fmt.Sprintf("ERROR adding device named: %s", deviceName))
				}
				ids = append(ids, idstr)
				p.lc.Info(fmt.Sprintf("Added NEW EdgeX device named: %s", deviceName))
			} else {
				p.lc.Debug(fmt.Sprintf("Updating EXISTING EdgeX device named: %s", deviceName))
				err = device.RunningService().UpdateDevice(edgexDevice)
				if err != nil {
					p.lc.Error(fmt.Sprintf("ERROR adding/updating device named: %s", deviceName))
				} else {
					p.lc.Info(fmt.Sprintf("Updated EXISTING EdgeX device named: %s", deviceName))
				}
			}
			// TODO: Update operational state to 'disabled'
			//  a) Query EdgeX (or our local caches) for all devices by name prefix.
			//  b) Filter out the discovered devices from above
			//  c) Assign remaining devices as 'disabled' if not discovered after [configurable #] scans.
		}
	}
	return ids, err
}

func (p *CameraDiscoveryProvider) assignTags(cameras []CameraInfo, tagCache *Tags) {
	if tagCache == nil {
		p.lc.Warn(fmt.Sprintf("TagCache is not available. Ignoring tag assignment"))
		return
	}
	for i := range cameras {
		p.lc.Debug(fmt.Sprintf("Attaching tagCache tags for serialNumber: %v to camera %d", cameras[i].SerialNumber, i))
		cameras[i].Tags = tagCache.Tags[cameras[i].SerialNumber]
	}
}

func (p *CameraDiscoveryProvider) memoizeCamInfoOnvif(cameras []CameraInfo, camInfoCache *CamInfo) {
	if camInfoCache == nil {
		p.lc.Warn(fmt.Sprintf("CamInfoCache is not available. Ignoring camInfo storage"))
		return
	}
	for i := range cameras {
		p.lc.Debug(fmt.Sprintf("Storing CamInfo to CamInfoCache.Onvif.. for device associated with serialNumber: %v camera index %d", cameras[i].SerialNumber, i))
		camInfoCache.AddOnvifCamera(p.lc, cameras[i])
	}
}

func (p *CameraDiscoveryProvider) memoizeCamInfoAxis(cameras []CameraInfo, camInfoCache *CamInfo) {
	if camInfoCache == nil {
		p.lc.Warn(fmt.Sprintf("CamInfoCache is not available. Ignoring camInfo storage"))
		return
	}
	for i := range cameras {
		p.lc.Debug(fmt.Sprintf("Storing CamInfo to CamInfoCache.Axis.. for device associated with serialNumber: %v camera index %d", cameras[i].SerialNumber, i))
		camInfoCache.AddAxisCamera(p.lc, cameras[i])
	}
}

// getCameraMetadata loops across requested camera vendors to resolve potential hosts on potential ports.
// This builds an array of CameraInfo elements containing ONVIF and Axis details from N hosts listening on N ports.
// --source=axis:554,555,556,80 --source=onvif:80,81,82,555,8000
func (p *CameraDiscoveryProvider) getCameraMetadata(camVendor string, sourceFlags []string, discoveredHosts map[string][]int, credentialsPath string) []CameraInfo {
	cameras := []CameraInfo{}
	p.lc.Debug(fmt.Sprintf("CameraDiscovery.GetCameraMetadata scanning discovered hosts for requested camera device signatures"))
	for host, ports := range discoveredHosts {
		for _, port := range ports {
			for i := range sourceFlags {
				if strings.Contains(sourceFlags[i], camVendor) {
					sourcePorts, source, err := p.getRequestedSourcePorts(sourceFlags[i])
					if err != nil {
						p.lc.Error(err.Error())
						// would return error but refactoring to build internal port lists earlier
						// (Initialize should handle any input errors)
					}
					// assert(source==requestedVendorSource)
					if source != camVendor {
						p.lc.Error("Invalid source encountered when looping for camera devices: source [%v] != requested [%v]", source, camVendor)
					}
					if isRequestedPort(sourcePorts, port) {
						p.lc.Trace(fmt.Sprintf("RequestingCameraDetails - DiscoveredHost [%s] port [%d] matches requested vendor source [%s] sourcePort[%v] for sourceFlagsVendor [%s]", host, port, camVendor, sourcePorts, source))
						cameras = p.getRequestedCameraDetails(source, port, host, credentialsPath, cameras)
					}
				}
			}
		}
	}
	return cameras
}

func isRequestedPort(sourcePorts []int, port int) bool {
	for _, p := range sourcePorts {
		if p == port {
			return true
		}
	}
	return false
}

func (p *CameraDiscoveryProvider) splitPorts(requestedPorts string) ([]int, error) {
	var err error
	a := strings.Split(requestedPorts, ",")
	b := make([]int, len(a))
	for i, v := range a {
		b[i], err = strconv.Atoi(v)
		if err != nil {
			return b, err
		}
	}
	return b, nil
}

// getRequestedSourcePorts reliably returns source and sourceport values for the provided sourceFlag input.
// This is used to assign default port as appropriate. It is also the current method by which we retain
// segmentation of desired set of ports requested per camera vendor.
// sourceFlags input parameter holds these varieties of values:
// --source=onvif
// --source=onvif:80
// --source=onvif:80,81,82
// --source=onvif --source=axis
// .. additional vendor:port assignments..
func (p *CameraDiscoveryProvider) getRequestedSourcePorts(sourceFlagInfo string) ([]int, string, error) {
	var err error
	var source string
	var sourcePorts []int
	if strings.Contains(sourceFlagInfo, ":") {
		s := strings.Split(sourceFlagInfo, ":")
		source = s[0]
		sourcePorts, err = p.splitPorts(s[1])
		if len(sourcePorts) == 0 {
			if source == p.options.SupportedSources[Axis].Name {
				sourcePorts = append(sourcePorts, p.options.SupportedSources[Axis].DefaultPort)
			}
			if source == p.options.SupportedSources[ONVIF].Name {
				sourcePorts = append(sourcePorts, p.options.SupportedSources[ONVIF].DefaultPort)
			}
		}
	} else {
		if strings.Contains(sourceFlagInfo, p.options.SupportedSources[Axis].Name) {
			source = p.options.SupportedSources[Axis].Name
			sourcePorts = append(sourcePorts, p.options.SupportedSources[Axis].DefaultPort)
		} else if strings.Contains(sourceFlagInfo, p.options.SupportedSources[ONVIF].Name) {
			source = p.options.SupportedSources[ONVIF].Name
			sourcePorts = append(sourcePorts, p.options.SupportedSources[ONVIF].DefaultPort)
		}
	}
	return sourcePorts, source, err
}

func (p *CameraDiscoveryProvider) getRequestedCameraDetails(source string, port int, host string, credentialsPath string, cameras []CameraInfo) []CameraInfo {
	//onvif camera
	if source == p.options.SupportedSources[ONVIF].Name {
		deviceAddr := host + ":" + strconv.Itoa(port)
		p.lc.Trace(fmt.Sprintf("CameraDiscovery.GetCameraMetadata invoking getOnvifCameraDetails(%s)", deviceAddr))
		cameraInfo, err := p.getOnvifCameraDetails(deviceAddr, credentialsPath)
		if err != nil {
			// Treat error as debug to avoid spamming logs, since many prospective hosts are simply not cameras.
			p.lc.Debug(err.Error())
		} else {
			cameras = append(cameras, cameraInfo)
		}
	}
	//axis camera
	if source == p.options.SupportedSources[Axis].Name {
		deviceAddr := host + ":" + strconv.Itoa(port)
		p.lc.Trace(fmt.Sprintf("CameraDiscovery.GetCameraMetadata invoking getAxisCameraDetails(%s)", deviceAddr))
		cameraInfo, err := p.getAxisCameraDetails(host, credentialsPath)
		if err != nil {
			// Continue on error - other cameras may return valid response
			p.lc.Debug(err.Error())
		} else {
			cameras = append(cameras, cameraInfo)
		}
	}
	return cameras
}
