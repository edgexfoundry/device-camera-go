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

	e_models "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	ds_models "github.com/edgexfoundry/device-sdk-go/pkg/models"

	"github.com/edgexfoundry-holding/device-camera-go/internal/pkg/digest"
	"github.com/edgexfoundry-holding/device-camera-go/internal/pkg/client"
)

const (
	CONF_ALARM_OVERVIEW     = "0x0c38"
	CONF_IVA_COUNTER_VALUES = "0x0b4a"

	ALARM_TYPE_UNKNOWN                      = 0
	ALARM_TYPE_VCA                          = 1
	ALARM_TYPE_RELAIS                       = 2
	ALARM_TYPE_DIGITAL_INTPUT               = 3
	ALARM_TYPE_AUDIO                        = 4
	ALARM_TYPE_VIRTUAL_INPUT                = 5
	ALARM_TYPE_DEFAULT_TASK                 = 16
	ALARM_TYPE_GLOBAL_CHANGE                = 17
	ALARM_TYPE_SIGNAL_TOO_BRIGHT            = 18
	ALARM_TYPE_SIGNAL_TOO_DARK              = 19
	ALARM_TYPE_REFERENCE_IMAGE_CHECK_FAILED = 23
	ALARM_TYPE_INVALID_CONFIGURATION        = 24
	ALARM_TYPE_FLAME_DETECTED               = 25
	ALARM_TYPE_SMOKE_DETECTED               = 26
	ALARM_TYPE_OBJECT_IN_FIELD              = 32
	ALARM_TYPE_CROSSING_LINE                = 33
	ALARM_TYPE_LOITERING                    = 34
	ALARM_TYPE_CONDITION_CHANGE             = 35
	ALARM_TYPE_FOLLOWING_ROUTE              = 36
	ALARM_TYPE_TAMPERING                    = 37
	ALARM_TYPE_REMOVED_OBJECT               = 38
	ALARM_TYPE_IDLE_OBJECT                  = 39
	ALARM_TYPE_ENTERING_FIELD               = 40
	ALARM_TYPE_LEAVING_FIELD                = 41
	ALARM_TYPE_SIMILARITY_SEARCH            = 42
	ALARM_TYPE_CROWD_DETECTION              = 43
	ALARM_TYPE_FLOW_IN_FIELD                = 44
	ALARM_TYPE_COUNTER_FLOW_IN_FIELD        = 45
	ALARM_TYPE_MOTION_IN_FIELD              = 46
	ALARM_TYPE_MAN_OVERBOARD                = 47
	ALARM_TYPE_COUNTER                      = 48
	ALARM_TYPE_BEV_PEOPLE_COUNTER           = 49
	ALARM_TYPE_OCCUPANCY                    = 50

	ALARM_ADD_FLAG       = 0x80
	ALARM_DELETE_FLAG    = 0x40
	ALARM_STATE_FLAG     = 0x20
	ALARM_STATE_SET_FLAG = 0x10

	RCP_FMT_URL = "http://%s/rcp.xml?%s=%s"
)

type Alarm struct {
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

type CounterData struct {
	ID    uint8
	Type  uint8
	Name  string
	Value uint32
}

type MsgList struct {
	XMLName xml.Name `xml:"message_list"`
	Msgs    []Msg    `xml:"msg"`
}

type Msg struct {
	Command string `xml:"command"`
	Num     string `xml:"num"`
	Cltid   string `xml:"cltid"`
	Hex     string `xml:"hex"`
}

func getXML(dc digest.Client, url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
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

func (p *packet) Byte(i int) uint8 {
	return uint8(binary.BigEndian.Uint32([]byte{0, 0, 0, p.buffer[i]}))
}

func (p *packet) Uint16(i int) uint16 {
	return binary.BigEndian.Uint16(p.buffer[i : i+2])
}

func (p *packet) Uint32(i int) uint32 {
	return binary.BigEndian.Uint32(p.buffer[i : i+4])
}

func (p *packet) UTF16String(i int, n int) string {
	ints := make([]uint16, n/2)
	if err := binary.Read(bytes.NewReader(p.buffer[i:i+n]), binary.BigEndian, &ints); err != nil {
		return ""
	}
	return string(utf16.Decode(ints))
}

func parseAlarms(bytes []byte) (alarms []Alarm) {
	packet := packet{buffer: bytes}

	readout := (packet.Byte(0) & 0x80) != 0

	// Alarm entries begin 4 bytes into payload
	for i := 4; i < len(packet.buffer); {
		var alarm Alarm
		alarm.EntryID = packet.Uint16(i)
		alarm.EntryLength = packet.Uint16(i + 2)

		flags := packet.Byte(i + 4)
		alarm.FlagAdd = (flags & ALARM_ADD_FLAG) != 0
		alarm.FlagDelete = (flags & ALARM_DELETE_FLAG) != 0
		alarm.FlagState = (flags & ALARM_STATE_FLAG) != 0
		alarm.FlagStateSet = (flags & ALARM_STATE_SET_FLAG) != 0

		alarm.AlarmSource = uint16(packet.Byte(i + 6))
		alarm.AlarmType = uint16(packet.Byte(i + 7))
		alarm.AlarmName = packet.UTF16String(i+8, int(alarm.EntryLength-8))
		i = i + int(alarm.EntryLength)

		if !readout || alarm.FlagState {
			alarms = append(alarms, alarm)
		}
	}
	return
}

func parseCounters(bytes []byte) (counters []CounterData) {
	packet := packet{buffer: bytes[1:]}

	for i := 0; i < len(packet.buffer); {
		var counter CounterData
		counter.ID = packet.Byte(i)
		counter.Type = packet.Byte(i + 1)
		counter.Name = packet.UTF16String(i+2, 64)
		counter.Value = packet.Uint32(i + 66)
		i = i + 70
		counters = append(counters, counter)
	}
	return
}

type RcpClient struct {
	client    digest.Client
	asyncChan chan<- *ds_models.AsyncValues
	lc        logger.LoggingClient

	alarms   map[int]e_models.DeviceResource
	counters map[string]e_models.DeviceResource

	alarmStates   map[int]bool
	counterStates map[string]int

	stop    chan bool
	stopped chan bool
}

func NewClient(asyncCh chan<- *ds_models.AsyncValues, lc logger.LoggingClient) client.Client {
	return &RcpClient{asyncChan: asyncCh, lc: lc}
}

func (rc *RcpClient) CameraRelease(force bool) {
	close(rc.stop)
	if !force {
		<-rc.stopped
	}
}

func (rc *RcpClient) CameraInit(edgexDevice e_models.Device, ipAddress string, username string, password string) {
	if rc.client == nil {
		rc.initializeDClient(username, password)
	}

	if rc.alarms == nil {
		rc.alarms = make(map[int]e_models.DeviceResource)
	}

	if rc.counters == nil {
		rc.counters = make(map[string]e_models.DeviceResource)
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
	deviceResources := edgexDevice.Profile.DeviceResources

	for _, e := range deviceResources {
		alarmType, ok := e.Attributes["alarm_type"]
		if ok {
			val, err := strconv.Atoi(alarmType)
			if err == nil {
				rc.alarms[val] = e
			}

			continue
		}

		counterName := e.Attributes["counter_name"]
		if counterName != "" {
			rc.counters[counterName] = e
		}

	}

	go func() {
		ticks := time.Tick(time.Second * 5)

		var max_errors = 60
		for max_errors > 0 {
			select {
			case <-ticks:
				err := rc.requestEvents(edgexDevice, ipAddress, stopchan)
				if err != nil {
					rc.lc.Error("Error in RCP loop: %s", err.Error())
					max_errors--
				} else {
					max_errors = 60
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

func (rc *RcpClient) HandleReadCommand(req ds_models.CommandRequest) (*ds_models.CommandValue, error) {
	var cv *ds_models.CommandValue
	var err error

	if alarmType, ok := req.Attributes["alarm_type"]; ok {
		alarmType, err := strconv.Atoi(alarmType)
		if err != nil {
			return nil, err
		}
		data := rc.getAlarmState(alarmType)

		cv, err = ds_models.NewBoolValue(req.DeviceResourceName, 0, data)
		if err != nil {
			return nil, err
		}
	} else if counterType, ok := req.Attributes["counter_name"]; ok {
		data := rc.getCounterState(counterType)

		cv, err = ds_models.NewUint32Value(req.DeviceResourceName, 0, uint32(data))
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("rcp: unrecognized read command")
	}

	return cv, nil
}

func (rc *RcpClient) HandleWriteCommand(req ds_models.CommandRequest, param *ds_models.CommandValue) error {
	return fmt.Errorf("rcp: unrecognized write command")
}

func (rc *RcpClient) commandValuesFromAlarms(alarms []Alarm, edgexDevice e_models.Device) ([]*ds_models.CommandValue, error) {
	cvs := make([]*ds_models.CommandValue, 0)
	var err error
	for _, alarm := range alarms {
		deviceResource, ok := rc.alarms[int(alarm.AlarmType)]

		if !ok {
			continue
		}

		alarmValue := alarm.FlagStateSet == alarm.FlagState

		rc.setAlarmState(int(alarm.AlarmType), alarmValue)

		var cv *ds_models.CommandValue
		cv, err = ds_models.NewBoolValue(deviceResource.Name, time.Now().UnixNano()/int64(time.Millisecond), alarmValue)
		if err != nil {
			rc.lc.Error("sendEvent: unable to get new bool value")
			return []*ds_models.CommandValue{}, fmt.Errorf("unable to create CommandValue")
		}
		cvs = append(cvs, cv)
	}

	return cvs, nil
}

func (rc *RcpClient) commandValuesFromCounters(counters []CounterData, edgexDevice e_models.Device) ([]*ds_models.CommandValue, error) {
	cvs := make([]*ds_models.CommandValue, 0)
	var err error
	for _, counter := range counters {
		dr, ok := rc.counters[counter.Name]
		if !ok {
			continue
		}

		rc.setCounterState(dr.Name, int(counter.Value))

		var cv *ds_models.CommandValue
		cv, err = ds_models.NewUint32Value(dr.Name, time.Now().UnixNano()/int64(time.Millisecond), counter.Value)
		if err != nil {
			rc.lc.Error("sendEvent: unable to get new Uint32 value")
			return []*ds_models.CommandValue{}, fmt.Errorf("unable to create CommandValue")
		}
		cvs = append(cvs, cv)
	}

	return cvs, nil
}

func (rc *RcpClient) requestEvents(device e_models.Device, ipAddress string, stopchan chan bool) error {
	url, err := getRcpURL(ipAddress, "message", CONF_ALARM_OVERVIEW+"$"+CONF_IVA_COUNTER_VALUES, map[string]string{"collectms": "5000"})
	if err != nil {
		rc.lc.Error("Error creating event polling url")
		return err
	}

	eventXml, err := getXML(rc.client, url)
	if err != nil {
		rc.lc.Error(fmt.Sprintf("error making request: %v", err.Error()))
		return err
	}
	var msgWrapper MsgList
	err = xml.Unmarshal(eventXml, &msgWrapper)
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

		var cvs []*ds_models.CommandValue
		switch msg.Command {
		case CONF_ALARM_OVERVIEW:
			alarms := parseAlarms(decoded)
			cvs, err = rc.commandValuesFromAlarms(alarms, device)
		case CONF_IVA_COUNTER_VALUES:
			counters := parseCounters(decoded)
			cvs, err = rc.commandValuesFromCounters(counters, device)
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

	formattedString := fmt.Sprintf(RCP_FMT_URL, ip, action, command)
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

func (rc *RcpClient) sendEvent(edgexDevice e_models.Device, cvs []*ds_models.CommandValue) {
	var av ds_models.AsyncValues
	av.DeviceName = edgexDevice.Name

	for _, cv := range cvs {
		av.CommandValues = append(av.CommandValues, cv)
	}

	rc.asyncChan <- &av
}
