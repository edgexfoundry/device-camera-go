// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019
//
// SPDX-License-Identifier: Apache-2.0

package bosch

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	sdkModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/edgexfoundry/device-camera-go/internal/pkg/client"
	"github.com/edgexfoundry/device-camera-go/internal/pkg/digest"
)

const (
	// confAlarmOverview and other constants are from the Bosch RCP documentation
	confAlarmOverview    = "0x0c38"
	confIvaCounterValues = "0x0b4a"

	alarmAddFlag      = 0x80
	alarmDeleteFlag   = 0x40
	alarmStateFlag    = 0x20
	alarmStateSetFlag = 0x10

	rcpFmtURL = "http://%s/rcp.xml?%s=%s"
)

type alarm struct {
	EntryID      uint16
	EntryLength  uint16
	FlagAdd      bool
	FlagDelete   bool
	FlagState    bool
	FlagStateSet bool
	AlarmSource  uint16
	AlarmType    uint16
	AlarmName    string
}

type counterData struct {
	ID    uint8
	Type  uint8
	Name  string
	Value uint32
}

type msgList struct {
	XMLName xml.Name `xml:"message_list"`
	Msgs    []msg    `xml:"msg"`
}

type msg struct {
	Command string `xml:"command"`
	Num     string `xml:"num"`
	Cltid   string `xml:"cltid"`
	Hex     string `xml:"hex"`
}

func getXML(dc digest.Client, url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("New request GET Error: %v", err.Error())
	}
	resp, err := dc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET Error: %v", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status Error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read Error: %v", err.Error())
	}
	return data, nil
}

type packet struct {
	buffer []byte
}

func (p *packet) byte(i int) uint8 {
	return uint8(binary.BigEndian.Uint32([]byte{0, 0, 0, p.buffer[i]}))
}

func (p *packet) uint16(i int) uint16 {
	return binary.BigEndian.Uint16(p.buffer[i : i+2])
}

func (p *packet) uint32(i int) uint32 {
	return binary.BigEndian.Uint32(p.buffer[i : i+4])
}

func (p *packet) utf16string(i int, n int) string {
	ints := make([]uint16, n/2)
	if err := binary.Read(bytes.NewReader(p.buffer[i:i+n]), binary.BigEndian, &ints); err != nil {
		return ""
	}
	return string(utf16.Decode(ints))
}

func parseAlarms(bytes []byte) (alarms []alarm) {
	packet := packet{buffer: bytes}

	readout := (packet.byte(0) & 0x80) != 0

	// alarm entries begin 4 bytes into payload
	for i := 4; i < len(packet.buffer); {
		var alarm alarm
		alarm.EntryID = packet.uint16(i)
		alarm.EntryLength = packet.uint16(i + 2)

		flags := packet.byte(i + 4)
		alarm.FlagAdd = (flags & alarmAddFlag) != 0
		alarm.FlagDelete = (flags & alarmDeleteFlag) != 0
		alarm.FlagState = (flags & alarmStateFlag) != 0
		alarm.FlagStateSet = (flags & alarmStateSetFlag) != 0

		alarm.AlarmSource = uint16(packet.byte(i + 6))
		alarm.AlarmType = uint16(packet.byte(i + 7))
		alarm.AlarmName = packet.utf16string(i+8, int(alarm.EntryLength-8))
		i = i + int(alarm.EntryLength)

		if !readout || alarm.FlagState {
			alarms = append(alarms, alarm)
		}
	}
	return
}

func parseCounters(bytes []byte) (counters []counterData) {
	packet := packet{buffer: bytes[1:]}

	for i := 0; i < len(packet.buffer); {
		var counter counterData
		counter.ID = packet.byte(i)
		counter.Type = packet.byte(i + 1)
		counter.Name = packet.utf16string(i+2, 64)
		counter.Value = packet.uint32(i + 66)
		i = i + 70
		counters = append(counters, counter)
	}
	return
}

// RcpClient is a client for accessing some analytics information from Bosch cameras via the RCP api.
// It is assumed that the analytics events are already configured on the device.
type RcpClient struct {
	client    digest.Client
	asyncChan chan<- *sdkModels.AsyncValues
	lc        logger.LoggingClient

	alarms   map[int]models.DeviceResource
	counters map[string]models.DeviceResource

	alarmStates   map[int]bool
	counterStates map[string]int

	stop    chan bool
	stopped chan bool
}

// NewClient creates a new RcpClient
func NewClient(asyncCh chan<- *sdkModels.AsyncValues, lc logger.LoggingClient) client.Client {
	return &RcpClient{asyncChan: asyncCh, lc: lc}
}

// CameraRelease stops the RCP listener routine
func (rc *RcpClient) CameraRelease(force bool) {
	close(rc.stop)
	if !force {
		<-rc.stopped
	}
}

// CameraInit initializes the RCP listener routine
func (rc *RcpClient) CameraInit(edgexDevice models.Device, edgexProfile models.DeviceProfile, ipAddress string, username string, password string) {
	if rc.client == nil {
		rc.initializeDClient(username, password)
	}

	if rc.alarms == nil {
		rc.alarms = make(map[int]models.DeviceResource)
	}

	if rc.counters == nil {
		rc.counters = make(map[string]models.DeviceResource)
	}

	if rc.alarmStates == nil {
		rc.alarmStates = make(map[int]bool)
	}

	if rc.counterStates == nil {
		rc.counterStates = make(map[string]int)
	}

	// a channel to tell us to stop
	stopchan := make(chan bool)

	// a channel to signal that it's stopped
	stoppedchan := make(chan bool)
	defer close(stoppedchan)

	// interrogate device profile for alarms/counters to listen for
	deviceResources := edgexProfile.DeviceResources

	for _, e := range deviceResources {
		alarmType, ok := e.Attributes["alarm_type"].(string)
		if ok {
			val, err := strconv.Atoi(alarmType)
			if err == nil {
				rc.alarms[val] = e
			}

			continue
		}

		counterName := e.Attributes["counter_name"].(string)
		if counterName != "" {
			rc.counters[counterName] = e
		}

	}

	go func() {
		ticks := time.Tick(time.Second * 5) //nolint:staticcheck

		var maxErrors = 60
		for maxErrors > 0 {
			select {
			case <-ticks:
				err := rc.requestEvents(edgexDevice, ipAddress, stopchan)
				if err != nil {
					rc.lc.Error(fmt.Sprintf("Error in RCP loop: %s", err.Error()))
					maxErrors--
				} else {
					maxErrors = 60
				}
			case <-stopchan:
				// stop
				return
			}
		}
	}()

	rc.stop = stopchan
	rc.stopped = stoppedchan
}

func (rc *RcpClient) initializeDClient(username string, password string) {
	rc.client = digest.NewDClient(&http.Client{}, username, password)
}

func (rc *RcpClient) setAlarmState(alarmType int, state bool) {
	rc.alarmStates[alarmType] = state
}

func (rc *RcpClient) getAlarmState(alarmType int) bool {
	return rc.alarmStates[alarmType]
}

func (rc *RcpClient) setCounterState(counter string, count int) {
	rc.counterStates[counter] = count
}

func (rc *RcpClient) getCounterState(counter string) int {
	return rc.counterStates[counter]
}

// HandleReadCommand handles requests to read data from the device via the RCP api
func (rc *RcpClient) HandleReadCommand(req sdkModels.CommandRequest) (*sdkModels.CommandValue, error) {
	var cv *sdkModels.CommandValue
	var err error

	if alarmType, ok := req.Attributes["alarm_type"].(string); ok {
		alarmType, err := strconv.Atoi(alarmType)
		if err != nil {
			return nil, err
		}
		data := rc.getAlarmState(alarmType)

		cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeBool, data)
		if err != nil {
			return nil, err
		}
	} else if counterType, ok := req.Attributes["counter_name"].(string); ok {
		data := rc.getCounterState(counterType)

		cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeUint32, uint32(data))
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("rcp: unrecognized read command")
	}

	return cv, nil
}

// HandleWriteCommand is unimplemented--any requests to it are unexpected
func (rc *RcpClient) HandleWriteCommand(req sdkModels.CommandRequest, param *sdkModels.CommandValue) error {
	return fmt.Errorf("rcp: unrecognized write command")
}

func (rc *RcpClient) commandValuesFromAlarms(alarms []alarm, edgexDevice models.Device) ([]*sdkModels.CommandValue, error) {
	cvs := make([]*sdkModels.CommandValue, 0)
	var err error
	for _, alarm := range alarms {
		deviceResource, ok := rc.alarms[int(alarm.AlarmType)]

		if !ok {
			continue
		}

		alarmValue := alarm.FlagStateSet == alarm.FlagState

		rc.setAlarmState(int(alarm.AlarmType), alarmValue)

		var cv *sdkModels.CommandValue
		cv, err = sdkModels.NewCommandValue(deviceResource.Name, common.ValueTypeBool, alarm)
		if err != nil {
			rc.lc.Error("sendEvent: unable to get new bool value")
			return []*sdkModels.CommandValue{}, fmt.Errorf("unable to create CommandValue")
		}
		cv.Origin = time.Now().UnixNano() / int64(time.Millisecond)
		cvs = append(cvs, cv)
	}

	return cvs, nil
}

func (rc *RcpClient) commandValuesFromCounters(counters []counterData, edgexDevice models.Device) ([]*sdkModels.CommandValue, error) {
	cvs := make([]*sdkModels.CommandValue, 0)
	var err error
	for _, counter := range counters {
		dr, ok := rc.counters[counter.Name]
		if !ok {
			continue
		}

		rc.setCounterState(dr.Name, int(counter.Value))

		var cv *sdkModels.CommandValue
		cv, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeUint32, counter.Value)
		if err != nil {
			rc.lc.Error("sendEvent: unable to get new uint32 value")
			return []*sdkModels.CommandValue{}, fmt.Errorf("unable to create CommandValue")
		}
		cv.Origin = time.Now().UnixNano() / int64(time.Millisecond)
		cvs = append(cvs, cv)
	}

	return cvs, nil
}

func (rc *RcpClient) requestEvents(device models.Device, ipAddress string, stopchan chan bool) error {
	url, err := getRcpURL(ipAddress, "message", confAlarmOverview+"$"+confIvaCounterValues, map[string]string{"collectms": "5000"})
	if err != nil {
		rc.lc.Error("Error creating event polling url")
		return err
	}

	eventXML, err := getXML(rc.client, url)
	if err != nil {
		rc.lc.Error(fmt.Sprintf("error making request: %v", err.Error()))
		return err
	}
	var msgWrapper msgList
	err = xml.Unmarshal(eventXML, &msgWrapper)
	if err != nil {
		rc.lc.Error(fmt.Sprintf("error unmarshaling: %v", err.Error()))
		return err
	}

	msgs := msgWrapper.Msgs

	for _, msg := range msgs {
		decoded, err := hex.DecodeString(msg.Hex[2:]) // Ignore 0x
		if err != nil {
			rc.lc.Error(fmt.Sprintf("error decoding: %v", err.Error()))
			continue
		}

		var cvs []*sdkModels.CommandValue
		switch msg.Command {
		case confAlarmOverview:
			alarms := parseAlarms(decoded)
			cvs, _ = rc.commandValuesFromAlarms(alarms, device)
		case confIvaCounterValues:
			counters := parseCounters(decoded)
			cvs, _ = rc.commandValuesFromCounters(counters, device)
		default:
			rc.lc.Warn("Unknown Command type in RCP Message")
		}

		if len(cvs) > 0 {
			rc.sendEvent(device, cvs)
		}
	}

	return nil
}

func getRcpURL(ip string, action string, command string, params map[string]string) (string, error) {
	if ip == "" || action == "" || command == "" {
		return "", fmt.Errorf("getRcpURL failed: required argument missing")
	}

	formattedString := fmt.Sprintf(rcpFmtURL, ip, action, command)
	var formattedArgs []string
	for k, v := range params {
		formattedArgs = append(formattedArgs, fmt.Sprintf("%s=%s", k, v))
	}
	paramString := strings.Join(formattedArgs, "&")

	if paramString != "" {
		paramString = "&" + paramString
	}

	return formattedString + paramString, nil
}

func (rc *RcpClient) sendEvent(edgexDevice models.Device, cvs []*sdkModels.CommandValue) {
	var av sdkModels.AsyncValues
	av.DeviceName = edgexDevice.Name

	av.CommandValues = append(av.CommandValues, cvs...)

	rc.asyncChan <- &av
}
