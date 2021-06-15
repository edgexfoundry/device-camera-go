/*******************************************************************************
 * Copyright 2021 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 *******************************************************************************/

package driver

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	sdkModel "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	sdk "github.com/edgexfoundry/device-sdk-go/v2/pkg/service"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/faceterteam/onvif4go/onvif"

	"github.com/edgexfoundry/device-camera-go/internal/pkg/axis"
	"github.com/edgexfoundry/device-camera-go/internal/pkg/bosch"
	"github.com/edgexfoundry/device-camera-go/internal/pkg/client"
	"github.com/edgexfoundry/device-camera-go/internal/pkg/noop"
)

var once sync.Once
var lock sync.Mutex

var onvifClients map[string]*OnvifClient
var clients map[string]client.Client

var driver *Driver

// Driver implements the sdkModel.ProtocolDriver interface for
// the device service
type Driver struct {
	lc       logger.LoggingClient
	asynchCh chan<- *sdkModel.AsyncValues
	config   *configuration
}

// NewProtocolDriver initializes the singleton Driver and
// returns it to the caller
func NewProtocolDriver() *Driver {
	once.Do(func() {
		driver = new(Driver)
		onvifClients = make(map[string]*OnvifClient)
		clients = make(map[string]client.Client)
	})

	return driver
}

// HandleReadCommands triggers a protocol Read operation for the specified device.
func (d *Driver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error) {
	var err error
	var responses = make([]*sdkModel.CommandValue, len(reqs))

	_, err = d.addrFromProtocols(protocols)
	if err != nil {
		return responses, fmt.Errorf("handleReadCommands: %w", err)
	}

	cameraConfig, err := CreateCameraInfo(protocols)
	if err != nil {
		return responses, fmt.Errorf("handleReadCommands: %w", err)
	}

	// check for existence of both clients
	onvifClient, c, err := d.clientsFromCameraConfig(cameraConfig, deviceName)
	if err != nil {
		return responses, fmt.Errorf("handleReadCommands: %w", err)
	}

	var data string
	var cv *sdkModel.CommandValue
	for i, req := range reqs {
		switch req.DeviceResourceName {
		// ONVIF cases
		case "onvif_device_information":
			data, err = onvifClient.GetDeviceInformation()
			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv, err = sdkModel.NewCommandValue(reqs[i].DeviceResourceName, common.ValueTypeString, string(data))
		case "onvif_profile_information":
			data, err = onvifClient.GetProfileInformation()
			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv, err = sdkModel.NewCommandValue(reqs[i].DeviceResourceName, common.ValueTypeString, string(data))
		case "OnvifDateTime":
			data, err = onvifClient.GetSystemDateAndTime()
			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv, err = sdkModel.NewCommandValue(reqs[i].DeviceResourceName, common.ValueTypeString, string(data))
		case "OnvifHostname":
			data, err = onvifClient.GetHostname()
			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv, err = sdkModel.NewCommandValue(reqs[i].DeviceResourceName, common.ValueTypeString, string(data))
		case "onvif_dns":
			data, err = onvifClient.GetDNS()
			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv, err = sdkModel.NewCommandValue(reqs[i].DeviceResourceName, common.ValueTypeString, string(data))
		case "onvif_network_interfaces":
			data, err = onvifClient.GetNetworkInterfaces()
			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv, err = sdkModel.NewCommandValue(reqs[i].DeviceResourceName, common.ValueTypeString, string(data))
		case "onvif_network_protocols":
			data, err = onvifClient.GetNetworkProtocols()
			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv, err = sdkModel.NewCommandValue(reqs[i].DeviceResourceName, common.ValueTypeString, string(data))
		case "onvif_network_default_gateway":
			data, err = onvifClient.GetNetworkDefaultGateway()
			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv, err = sdkModel.NewCommandValue(reqs[i].DeviceResourceName, common.ValueTypeString, string(data))
		case "onvif_ntp":
			data, err = onvifClient.GetNTP()
			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv, err = sdkModel.NewCommandValue(reqs[i].DeviceResourceName, common.ValueTypeString, string(data))
		case "onvif_system_reboot":
			data, err = onvifClient.Reboot()
			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv, err = sdkModel.NewCommandValue(reqs[i].DeviceResourceName, common.ValueTypeString, string(data))
		case "onvif_users":
			data, err = onvifClient.GetUsers()
			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv, err = sdkModel.NewCommandValue(reqs[i].DeviceResourceName, common.ValueTypeString, string(data))
		case "onvif_snapshot":
			var bytes []byte
			bytes, err = onvifClient.GetSnapshot()
			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv, err = sdkModel.NewCommandValue(reqs[i].DeviceResourceName, common.ValueTypeBinary, bytes)
		case "OnvifStreamURI":
			data, err = onvifClient.GetStreamURI()
			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv, err = sdkModel.NewCommandValue(reqs[i].DeviceResourceName, common.ValueTypeString, string(data))
		// camera specific cases
		default:
			if c == nil {
				err := errors.New("Non-ONVIF command for camera without secondary client")
				d.lc.Error(err.Error())
				return responses, err
			}

			cv, err = c.HandleReadCommand(req)
			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}
		}
		if err != nil {
			d.lc.Errorf("Error creating CommandValue: %s", err.Error())
			return responses, err
		}
		responses[i] = cv
	}

	return responses, nil
}

// HandleWriteCommands passes a slice of CommandRequest struct each representing
// a ResourceOperation for a specific device resource (aka DeviceObject).
// Since the commands are actuation commands, params provide parameters for the individual
// command.
func (d *Driver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	_, err := d.addrFromProtocols(protocols)
	if err != nil {
		return fmt.Errorf("handleWriteCommands: %w", err)
	}

	cameraConfig, err := CreateCameraInfo(protocols)
	if err != nil {
		return fmt.Errorf("while write commands, failed to create cameraInfo for device %s: %w", deviceName, err)
	}

	// check for existence of both clients
	onvifClient, c, err := d.clientsFromCameraConfig(cameraConfig, deviceName)
	if err != nil {
		return fmt.Errorf("handleWriteCommands: %w", err)
	}

	for i, req := range reqs {
		switch req.DeviceResourceName {
		case "OnvifUser":
			user := struct {
				Username  string
				Password  string
				UserLevel string
				Extension *string
			}{}

			err := structFromParam(params[i], &user)
			if err != nil {
				d.lc.Error(err.Error())
				return err
			}

			onvifUser := onvif.User{
				Username:  user.Username,
				Password:  user.Password,
				UserLevel: onvif.UserLevel(user.UserLevel),
				Extension: (*onvif.UserExtension)(user.Extension),
			}

			err = onvifClient.CreateUser(onvifUser)
			if err != nil {
				d.lc.Error(fmt.Sprintf("handleWriteCommands error: %v", err.Error()))
				return err
			}

		case "OnvifReboot":
			shouldReboot, err := params[i].BoolValue()
			if err != nil {
				err := errors.New("non-binary value passed to OnvifReboot command")
				d.lc.Error(err.Error())
				return err
			}
			if !shouldReboot {
				continue
			}

			_, err = onvifClient.Reboot()
			if err != nil {
				return err
			}

		case "OnvifHostname":
			hostname, err := params[i].StringValue()
			if err != nil {
				err := errors.New("non-string value passed to OnvifHostname command")
				d.lc.Error(err.Error())
				return err
			}

			err = onvifClient.SetHostname(hostname)
			if err != nil {
				d.lc.Error(err.Error())
				return err
			}

		case "OnvifHostnameFromDHCP":
			err := onvifClient.SetHostnameFromDHCP()
			if err != nil {
				d.lc.Error(err.Error())
				return err
			}

		case "OnvifDateTime":
			dateTime := struct {
				Year   int
				Month  int
				Day    int
				Hour   int
				Minute int
				Second int
			}{}

			err := structFromParam(params[i], &dateTime)
			if err != nil {
				d.lc.Error(err.Error())
				return err
			}

			t := time.Date(dateTime.Year, time.Month(dateTime.Month), dateTime.Day, dateTime.Hour, dateTime.Minute, dateTime.Second, 0, time.UTC)
			err = onvifClient.SetSystemDateAndTime(t)
			if err != nil {
				d.lc.Error(err.Error())
				return err
			}

		default:
			if c == nil {
				err := errors.New("non-onvif command for camera without secondary client")
				d.lc.Error(err.Error())
				return err
			}

			err := c.HandleWriteCommand(reqs[i], params[i])
			if err != nil {
				d.lc.Error(err.Error())
				return err
			}
		}

	}

	return nil
}

type stringer interface {
	StringValue() (string, error)
}

func structFromParam(s stringer, v interface{}) error {
	str, err := s.StringValue()
	if err != nil {
		return errors.New("OnvifUser CommandValue missing string value")
	}
	err = json.Unmarshal([]byte(str), v)
	if err != nil {
		return fmt.Errorf("error unmarshaling string: %w", err)
	}
	return nil
}

// DisconnectDevice handles protocol-specific cleanup when a device
// is removed.
func (d *Driver) DisconnectDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	addr, err := d.addrFromProtocols(protocols)
	if err != nil {
		return fmt.Errorf("no address found for device: %w", err)
	}

	shutdownClient(addr)
	shutdownOnvifClient(addr)
	return nil
}

// Initialize performs protocol-specific initialization for the device
// service.
func (d *Driver) Initialize(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues,
	deviceCh chan<- []sdkModel.DiscoveredDevice) error {
	d.lc = lc
	d.asynchCh = asyncCh

	camConfig, err := loadCameraConfig(sdk.DriverConfigs())
	if err != nil {
		panic(fmt.Errorf("load camera configuration failed: %w", err))
	}
	d.config = camConfig

	deviceService := sdk.RunningService()

	for _, dev := range deviceService.Devices() {
		camInfo, err := CreateCameraInfo(dev.Protocols)

		if err != nil {
			return fmt.Errorf("failed to create cameraInfo for camera %s: %w", dev.Name, err)
		}

		var creds config.Credentials

		// need to retrieve credentials from secret provider when auth method is either basic or digest
		if authMethod := camInfo.AuthMethod; authMethod == BASIC_AUTH || authMethod == DIGEST_AUTH {
			// each camera can have different credentials
			creds, err = GetCredentials(camInfo.CredentialPaths)
			if err != nil {
				return fmt.Errorf("failed to get credentials for camera %s: %w", dev.Name, err)
			}
		}

		initializeOnvifClient(dev, creds.Username, creds.Password, camInfo.AuthMethod)
		newClient(dev, creds.Username, creds.Password)
	}

	return nil
}

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (d *Driver) Stop(force bool) error {
	for _, c := range clients {
		c.CameraRelease(force)
	}

	close(d.asynchCh)

	return nil
}

// AddDevice is a callback function that is invoked
// when a new Device associated with this Device Service is added
func (d *Driver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	_, err := d.addrFromProtocols(protocols)
	if err != nil {
		err = fmt.Errorf("error adding device: %w", err)
		d.lc.Error(err.Error())
		return err
	}

	cameraConfig, err := CreateCameraInfo(protocols)
	if err != nil {
		return fmt.Errorf("while add device, failed to create cameraInfo for device %s: %w", deviceName, err)
	}

	_, _, err = d.clientsFromCameraConfig(cameraConfig, deviceName)
	if err != nil {
		err = fmt.Errorf("error adding device: %w", err)
		d.lc.Error(err.Error())
		return err
	}
	return nil
}

// UpdateDevice is a callback function that is invoked
// when a Device associated with this Device Service is updated
func (d *Driver) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	return nil
}

// RemoveDevice is a callback function that is invoked
// when a Device associated with this Device Service is removed
func (d *Driver) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	addr, err := d.addrFromProtocols(protocols)
	if err != nil {
		return fmt.Errorf("no address found for device: %w", err)
	}

	shutdownClient(addr)
	shutdownOnvifClient(addr)
	return nil
}

func newClient(device models.Device, user string, password string) client.Client {
	profile, _ := sdk.RunningService().GetProfileByName(device.ProfileName)
	labels := profile.Labels
	var c client.Client

	if in("bosch", labels) {
		c = initializeClient(device, profile, user, password)
	} else if in("hanwha", labels) {
		// c = initializeHanwhaClient(device, user, password)
	} else if in("axis", labels) {
		c = initializeAxisClient(device, profile, user, password)
	} else {
		c = initializeNoopClient()
	}

	return c
}

func getOnvifClient(addr string) (*OnvifClient, bool) {
	lock.Lock()
	c, ok := onvifClients[addr]
	lock.Unlock()
	return c, ok
}

func getClient(addr string) (client.Client, bool) {
	lock.Lock()
	c, ok := clients[addr]
	lock.Unlock()
	return c, ok
}

func initializeOnvifClient(device models.Device, user string, password string, authMethod string) *OnvifClient {
	addr := device.Protocols[HTTP_PROTOCOL][ADDRESS]

	// go to secretstore with credential path to get username and password
	c := NewOnvifClient(addr, user, password, authMethod, driver.lc)
	if c != nil {
		// Only add the ONVIF client if it could be initialized. if it's offline then we might try again in an autoevent
		lock.Lock()
		onvifClients[addr] = c
		lock.Unlock()
	}
	return c
}

func initializeClient(device models.Device, profile models.DeviceProfile, user string, password string) client.Client {
	addr := device.Protocols[HTTP_PROTOCOL][ADDRESS]

	c := bosch.NewClient(driver.asynchCh, driver.lc)
	c.CameraInit(device, profile, addr, user, password)

	lock.Lock()
	clients[addr] = c
	lock.Unlock()

	return c
}

func initializeAxisClient(device models.Device, profile models.DeviceProfile, user string, password string) client.Client {
	addr := device.Protocols[HTTP_PROTOCOL][ADDRESS]

	c := axis.NewClient(driver.asynchCh, driver.lc)
	c.CameraInit(device, profile, addr, user, password)

	lock.Lock()
	clients[addr] = c
	lock.Unlock()

	return c
}

func initializeNoopClient() client.Client {
	c := noop.NewClient()
	return c
}

func shutdownOnvifClient(addr string) {
	// nothing much to do here at the moment
	lock.Lock()
	delete(onvifClients, addr)
	lock.Unlock()
}

func shutdownClient(addr string) {
	lock.Lock()

	clients[addr].CameraRelease(true)
	delete(clients, addr)

	lock.Unlock()
}

func in(needle string, haystack []string) bool {
	for _, e := range haystack {
		if needle == e {
			return true
		}
	}
	return false
}

func (d *Driver) addrFromProtocols(protocols map[string]models.ProtocolProperties) (string, error) {
	if _, ok := protocols[HTTP_PROTOCOL]; !ok {
		d.lc.Error("No HTTP address found for device. Check configuration file.")
		return "", errors.New("no HTTP address in protocols map")
	}

	var addr string
	addr, ok := protocols[HTTP_PROTOCOL][ADDRESS]
	if !ok {
		d.lc.Error("No HTTP address found for device. Check configuration file.")
		return "", errors.New("no HTTP address in protocols map")
	}
	return addr, nil

}

func (d *Driver) clientsFromCameraConfig(cameraConfig *cameraInfo, deviceName string) (*OnvifClient, client.Client, error) {
	onvifClient, ok := getOnvifClient(cameraConfig.Address)

	if !ok {
		dev, err := sdk.RunningService().GetDeviceByName(deviceName)
		if err != nil {
			err = fmt.Errorf("device not found: %s", deviceName)
			d.lc.Error(err.Error())

			return nil, nil, err
		}

		var creds config.Credentials
		// need to retrieve credentials from secret provider when auth method is either basic or digest
		if authMethod := cameraConfig.AuthMethod; authMethod == BASIC_AUTH || authMethod == DIGEST_AUTH {
			creds, err = GetCredentials(cameraConfig.CredentialPaths)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get credentials for %s: %w", deviceName, err)
			}
		}

		onvifClient = initializeOnvifClient(dev, creds.Username, creds.Password, cameraConfig.AuthMethod)

		if onvifClient == nil {
			err := fmt.Errorf("ONVIF client could not be initialized: %s", deviceName)
			d.lc.Error(err.Error())
			return nil, nil, err
		}

	}

	c, ok := getClient(cameraConfig.Address)

	if !ok {
		dev, err := sdk.RunningService().GetDeviceByName(deviceName)
		if err != nil {
			err = fmt.Errorf("device not found: %s", deviceName)
			d.lc.Error(err.Error())

			return nil, nil, err
		}

		var creds config.Credentials
		// need to retrieve credentials from secret provider when auth method is either basic or digest
		if authMethod := cameraConfig.AuthMethod; authMethod == BASIC_AUTH || authMethod == DIGEST_AUTH {
			creds, err = GetCredentials(cameraConfig.CredentialPaths)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get credentials for %s: %w", deviceName, err)
			}
		}

		newClient(dev, creds.Username, creds.Password)
	}

	return onvifClient, c, nil
}

func GetCredentials(secretPath string) (config.Credentials, error) {
	credentials := config.Credentials{}
	deviceService := sdk.RunningService()

	timer := startup.NewTimer(driver.config.CredentialsRetryTime, driver.config.CredentialsRetryWait)

	var secretData map[string]string
	var err error
	for timer.HasNotElapsed() {
		secretData, err = deviceService.SecretProvider.GetSecret(secretPath, secret.UsernameKey, secret.PasswordKey)
		if err == nil {
			break
		}

		driver.lc.Warnf(
			"Unable to retrieve camera credentials from SecretProvider at path '%s': %s. Retrying for %s",
			secretPath,
			err.Error(),
			timer.RemainingAsString())
		timer.SleepForInterval()
	}

	if err != nil {
		return credentials, err
	}

	credentials.Username = secretData[secret.UsernameKey]
	credentials.Password = secretData[secret.PasswordKey]

	return credentials, nil
}
