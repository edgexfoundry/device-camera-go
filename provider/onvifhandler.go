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
	"errors"
	"fmt"
	"github.com/atagirov/goonvif"
	"github.com/atagirov/goonvif/Device"
	"github.com/atagirov/goonvif/Media"
	"github.com/atagirov/goonvif/xsd/onvif"
	"github.com/beevik/etree"
	"io/ioutil"
	"net/http"
	"strconv"
)

// VendorInfoAxis holds the overloads for Axis support; inheriting base from ONVIF for now
type VendorInfoAxis struct {
	SupportedResolutions []string `json:"supportedresolutions,omitempty"`
	SupportedFormats     []string `json:"supportedformats,omitempty"`
	RTSPAuthEnabled      bool     `json:"rtspauthenabled,omitempty"`
	RTSPPath             string   `json:"rtsppath,omitempty"`
	ImagePath            string   `json:"imagepath,omitempty"`
}

// CameraInfo holds ONVIF camera device details.
// Locate available conformant device manufacturers at https://www.onvif.org/conformant-products/
type CameraInfo struct {
	VendorInfoAxis
	IPAddress       string                 `json:"ip"`
	ProductName     string                 `json:"productname"`
	FirmwareVersion string                 `json:"firmwareversion"`
	SerialNumber    string                 `json:"serialnumber"`
	Profiles        []Profile              `json:"profiles,omitempty"`
	Tags            map[string]interface{} `json:"tags,omitempty"`
}

// Profile holds ONVIF profile details.
// Ref: https://www.onvif.org/profiles/
type Profile struct {
	ProfileName  string
	Formats      string
	Resolutions  []string
	RTSPPath     string
	ImagePath    string
	ProfileToken string
}

const (
	profileTokenKey = "token"
)

// getOnvifCameraDetails populates CameraInfo structure for ONVIF compliant cameras
// nolint: gocyclo
func (p *CameraDiscoveryProvider) getOnvifCameraDetails(addr string, credentials CredentialsInfo) (CameraInfo, error) {
	//Getting an camera instance
	p.lc.Trace(fmt.Sprintf("Invoking ONVIF.NewDevice"))
	dev, err := goonvif.NewDevice(addr)
	if err != nil {
		// a prospective host we scanned likely just not a camera
		p.lc.Debug(err.Error())
		return CameraInfo{}, errors.New("Host " + addr + " not responding or does not support ONVIF services")
	}
	//Authorization
	dev.Authenticate(credentials.User, credentials.Pass)

	//Populate Device Info
	var listProfiles []Profile
	cameraDetailsMap := make(map[string]string)
	aDeviceInformation := Device.GetDeviceInformation{}
	p.lc.Trace(fmt.Sprintf("Invoking ONVIF.GetDeviceInformation"))
	GetDeviceInformationResponse, err := dev.CallMethod(aDeviceInformation)
	p.lc.Info(fmt.Sprintf("ONVIF.GetDeviceInformation yielded: %v", GetDeviceInformationResponse))
	if err != nil && GetDeviceInformationResponse.StatusCode != http.StatusOK && GetDeviceInformationResponse.StatusCode != http.StatusUnauthorized {
		p.lc.Error(err.Error())
	} else {
		doc := etree.NewDocument()
		if b, err2 := ioutil.ReadAll(GetDeviceInformationResponse.Body); err2 != nil {
			p.lc.Error(fmt.Sprintf("Error reading ONVIF device body: %v", err2))
		} else {
			if err = doc.ReadFromBytes(b); err != nil {
				p.lc.Error(fmt.Sprintf("Error reading ONVIF device information response: %v", err))
			}
			getResponseElements := doc.FindElements("./Envelope/Body/GetDeviceInformationResponse/*")
			for _, j := range getResponseElements {
				cameraDetailsMap[j.Tag] = j.Text()
			}
			p.lc.Debug(fmt.Sprintf("Camera Details: %v", cameraDetailsMap))
		}
	}

	// Populate Profiles
	profiles := Media.GetProfiles{}
	profilesresponse, err := dev.CallMethod(profiles)
	if err != nil {
		p.lc.Error(err.Error())
	} else {
		doc := etree.NewDocument()
		b, err := ioutil.ReadAll(profilesresponse.Body)
		if err != nil {
			p.lc.Error(fmt.Sprintf("Error reading ONVIF profile response body: %v", err))
		} else {
			if err = doc.ReadFromBytes(b); err != nil {
				p.lc.Error(fmt.Sprintf("Error reading ONVIF profiles response details: %v", err))
			}
		}
		for i1, getProfileResponseElements := range doc.FindElements("./Envelope/Body/GetProfilesResponse/*") {
			var name string
			var formats string
			var resolutions []string
			var rtspPath string
			var imgPath string
			i1 = i1 + 1
			// Populate profile.name
			for _, getName := range doc.FindElements("./Envelope/Body/GetProfilesResponse/Profiles[" + strconv.Itoa(i1) + "]/Name") {
				name = getName.Text()
			}
			// Populate profile.formats
			for _, getEncoding := range doc.FindElements("./Envelope/Body/GetProfilesResponse/Profiles[" + strconv.Itoa(i1) + "]/VideoEncoderConfiguration/Encoding") {
				formats = getEncoding.Text()
			}
			// Populate profile.resolutions
			for _, getresolution := range doc.FindElements("./Envelope/Body/GetProfilesResponse/Profiles[" + strconv.Itoa(i1) + "]/VideoEncoderConfiguration/Resolution/*") {
				resolutions = append(resolutions, getresolution.Text())
			}
			profileToken := getProfileResponseElements.SelectAttr(profileTokenKey)
			if profileToken == nil || len(profileToken.Value) < 1 {
				continue
			}
			// Populate RTSPPath
			uri := Media.GetStreamUri{
				StreamSetup: onvif.StreamSetup{
					Stream:    onvif.StreamType("RTP-Unicast"),
					Transport: onvif.Transport{Protocol: "RTSP"},
				},
				ProfileToken: onvif.ReferenceToken(profileToken.Value),
			}
			if uriresponse, err := dev.CallMethod(uri); err == nil {
				rtspPath = p.getRTSPPath(uriresponse)
			} else {
				p.lc.Error(err.Error())
			}
			// Populate ImagePath
			uri2 := Media.GetSnapshotUri{
				ProfileToken: onvif.ReferenceToken(profileToken.Value),
			}
			if uriresponse2, err2 := dev.CallMethod(uri2); err2 == nil {
				imgPath = p.getImagePath(uriresponse2)
			} else {
				p.lc.Error(err2.Error())
			}
			// Perform validity check - ignore any profile not containing video stream URI
			if rtspPath == "" {
				continue
			}
			profiles := Profile{
				ProfileName:  name,
				Formats:      formats,
				Resolutions:  resolutions,
				RTSPPath:     rtspPath,
				ImagePath:    imgPath,
				ProfileToken: profileToken.Value,
			}
			listProfiles = append(listProfiles, profiles)
		}
	}
	cameraInfo := CameraInfo{
		IPAddress:       addr,
		ProductName:     cameraDetailsMap["Manufacturer"],
		FirmwareVersion: cameraDetailsMap["FirmwareVersion"],
		SerialNumber:    cameraDetailsMap["SerialNumber"],
		Profiles:        listProfiles,
	}
	return cameraInfo, nil
}

func (p *CameraDiscoveryProvider) getRTSPPath(uriresponse *http.Response) string {
	rtspPath := ""
	doc1 := etree.NewDocument()
	b, err := ioutil.ReadAll(uriresponse.Body)
	if err != nil {
		p.lc.Error(fmt.Sprintf("Error reading RTP body: %v", err))
	} else {
		if err = doc1.ReadFromBytes(b); err != nil {
			// May need to quash some of these
			p.lc.Error(fmt.Sprintf("Error reading ONVIF RTP URI response: %v", err))
		}
	}
	getStreamURIResponseElements := doc1.FindElements("./Envelope/Body/*")
	for _, j := range getStreamURIResponseElements {
		if j.Tag == "Fault" {
			// swallow error (treat as debug) and move on
			p.lc.Trace("Error: Incomplete configuration")
		}
		if j.Tag == "GetStreamUriResponse" {
			getURIElement := doc1.FindElement("./Envelope/Body/GetStreamUriResponse/MediaUri/Uri")
			if getURIElement != nil {
				rtspPath = getURIElement.Text()
			}
		}
	}
	return rtspPath
}

func (p *CameraDiscoveryProvider) getImagePath(uriresponse2 *http.Response) string {
	imgPath := ""
	doc2 := etree.NewDocument()
	b, err := ioutil.ReadAll(uriresponse2.Body)
	if err != nil {
		p.lc.Error(fmt.Sprintf("Error reading ONVIF image body: %v", err))
	} else {
		if err = doc2.ReadFromBytes(b); err != nil {
			// May need to quash some of these
			p.lc.Error(fmt.Sprintf("Error reading ONVIF Snapshot URI response: %v", err))
		}
	}
	getImageURIResponseElements := doc2.FindElements("./Envelope/Body/*")
	for _, j := range getImageURIResponseElements {
		if j.Tag == "Fault" {
			// swallow error (treat as debug) and move on
			p.lc.Trace("Error: Incomplete configuration")
		}
		if j.Tag == "GetSnapshotUriResponse" {
			getURIElement := doc2.FindElement("./Envelope/Body/GetSnapshotUriResponse/MediaUri/Uri")
			if (getURIElement != nil) {
				imgPath = getURIElement.Text()
			}
		}
	}
	return imgPath
}
