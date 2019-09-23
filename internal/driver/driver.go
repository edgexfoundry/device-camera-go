package driver

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/device-sdk-go"
	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/faceterteam/onvif4go/onvif"
	"github.com/pkg/errors"

	"github.com/edgexfoundry-holding/device-camera-go/internal/pkg/axis"
	"github.com/edgexfoundry-holding/device-camera-go/internal/pkg/bosch"
	"github.com/edgexfoundry-holding/device-camera-go/internal/pkg/client"
	"github.com/edgexfoundry-holding/device-camera-go/internal/pkg/noop"
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
func (d *Driver) HandleReadCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error) {
	var responses = make([]*sdkModel.CommandValue, len(reqs))

	addr, err := d.addrFromProtocols(protocols)
	if err != nil {
		return responses, errors.Errorf("handleReadCommands: %v", err.Error())
	}

	// check for existence of both clients
	onvifClient, c, err := d.clientsFromAddr(addr, deviceName)
	if err != nil {
		return responses, errors.Errorf("handleReadCommands: %v", err.Error())
	}

	for i, req := range reqs {
		switch req.DeviceResourceName {
		// ONVIF cases
		case "onvif_device_information":
			data, err := onvifClient.GetDeviceInformation()

			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv := sdkModel.NewStringValue(reqs[i].DeviceResourceName, 0, string(data))
			responses[i] = cv
		case "onvif_profile_information":
			data, err := onvifClient.GetProfileInformation()

			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv := sdkModel.NewStringValue(reqs[i].DeviceResourceName, 0, string(data))
			responses[i] = cv
		case "OnvifDateTime":
			data, err := onvifClient.GetSystemDateAndTime()

			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv := sdkModel.NewStringValue(reqs[i].DeviceResourceName, 0, string(data))
			responses[i] = cv
		case "OnvifHostname":
			data, err := onvifClient.GetHostname()

			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv := sdkModel.NewStringValue(reqs[i].DeviceResourceName, 0, string(data))
			responses[i] = cv
		case "onvif_dns":
			data, err := onvifClient.GetDNS()

			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv := sdkModel.NewStringValue(reqs[i].DeviceResourceName, 0, string(data))
			responses[i] = cv
		case "onvif_network_interfaces":
			data, err := onvifClient.GetNetworkInterfaces()

			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv := sdkModel.NewStringValue(reqs[i].DeviceResourceName, 0, string(data))
			responses[i] = cv
		case "onvif_network_protocols":
			data, err := onvifClient.GetNetworkProtocols()

			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv := sdkModel.NewStringValue(reqs[i].DeviceResourceName, 0, string(data))
			responses[i] = cv
		case "onvif_network_default_gateway":
			data, err := onvifClient.GetNetworkDefaultGateway()

			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv := sdkModel.NewStringValue(reqs[i].DeviceResourceName, 0, string(data))
			responses[i] = cv
		case "onvif_ntp":
			data, err := onvifClient.GetNTP()

			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv := sdkModel.NewStringValue(reqs[i].DeviceResourceName, 0, string(data))
			responses[i] = cv
		case "onvif_system_reboot":
			data, err := onvifClient.Reboot()

			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv := sdkModel.NewStringValue(reqs[i].DeviceResourceName, 0, string(data))
			responses[i] = cv
		case "onvif_users":
			data, err := onvifClient.GetUsers()

			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv := sdkModel.NewStringValue(reqs[i].DeviceResourceName, 0, string(data))
			responses[i] = cv
		case "onvif_snapshot":
			data, err := onvifClient.GetSnapshot()

			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv, err := sdkModel.NewBinaryValue(reqs[i].DeviceResourceName, 0, data)
			if err != nil {
				err = errors.Wrap(err, "error creating binary CommandValue")
				d.lc.Error(err.Error())
				return responses, err
			}
			responses[i] = cv
		case "OnvifStreamURI":
			data, err := onvifClient.GetStreamURI()

			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}

			cv := sdkModel.NewStringValue(reqs[i].DeviceResourceName, 0, string(data))
			responses[i] = cv

		// camera specific cases
		default:
			if c == nil {
				err := errors.New("Non-ONVIF command for camera without secondary client")
				d.lc.Error(err.Error())
				return responses, err
			}

			cv, err := c.HandleReadCommand(req)
			if err != nil {
				d.lc.Error(err.Error())
				return responses, err
			}
			responses[i] = cv
		}
	}

	return responses, nil
}

// HandleWriteCommands passes a slice of CommandRequest struct each representing
// a ResourceOperation for a specific device resource (aka DeviceObject).
// Since the commands are actuation commands, params provide parameters for the individual
// command.
func (d *Driver) HandleWriteCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	addr, err := d.addrFromProtocols(protocols)
	if err != nil {
		return errors.Errorf("handleWriteCommands: %v", err.Error())
	}

	// check for existence of both clients
	onvifClient, c, err := d.clientsFromAddr(addr, deviceName)
	if err != nil {
		return errors.Errorf("handleWriteCommands: %v", err.Error())
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
			fmt.Println(req.DeviceResourceName)
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
		return errors.Errorf("OnvifUser CommandValue missing string value")
	}
	err = json.Unmarshal([]byte(str), v)
	if err != nil {
		return errors.Errorf("error unmarshaling string: %v", err.Error())
	}
	return nil
}

// DisconnectDevice handles protocol-specific cleanup when a device
// is removed.
func (d *Driver) DisconnectDevice(deviceName string, protocols map[string]contract.ProtocolProperties) error {
	addr, err := d.addrFromProtocols(protocols)
	if err != nil {
		return errors.Errorf("no address found for device: %v", err.Error())
	}

	shutdownClient(addr)
	shutdownOnvifClient(addr)
	return nil
}

// Initialize performs protocol-specific initialization for the device
// service.
func (d *Driver) Initialize(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues) error {
	d.lc = lc
	d.asynchCh = asyncCh

	config, err := loadConfigFromFile()
	if err != nil {
		panic(fmt.Errorf("read driver configuration from file failed: %d", err))
	}
	d.config = config

	for _, dev := range device.RunningService().Devices() {
		initializeOnvifClient(dev, config.Camera.User, config.Camera.Password)
		newClient(dev, config.Camera.User, config.Camera.Password)
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
func (d *Driver) AddDevice(deviceName string, protocols map[string]contract.ProtocolProperties, adminState contract.AdminState) error {
	addr, err := d.addrFromProtocols(protocols)
	if err != nil {
		err = errors.Errorf("error adding device: %v", err.Error())
		d.lc.Error(err.Error())
		return err
	}

	_, _, err = d.clientsFromAddr(addr, deviceName)
	if err != nil {
		err = errors.Errorf("error adding device: %v", err.Error())
		d.lc.Error(err.Error())
		return err
	}
	return nil
}

// UpdateDevice is a callback function that is invoked
// when a Device associated with this Device Service is updated
func (d *Driver) UpdateDevice(deviceName string, protocols map[string]contract.ProtocolProperties, adminState contract.AdminState) error {
	return nil
}

// RemoveDevice is a callback function that is invoked
// when a Device associated with this Device Service is removed
func (d *Driver) RemoveDevice(deviceName string, protocols map[string]contract.ProtocolProperties) error {
	addr, err := d.addrFromProtocols(protocols)
	if err != nil {
		return errors.Errorf("no address found for device: %v", err.Error())
	}

	shutdownClient(addr)
	shutdownOnvifClient(addr)
	return nil
}

func newClient(device contract.Device, user string, password string) client.Client {
	labels := device.Profile.Labels
	var c client.Client

	if in("bosch", labels) {
		c = initializeClient(device, user, password)
	} else if in("hanwha", labels) {
		// c = initializeHanwhaClient(device, user, password)
	} else if in("axis", labels) {
		c = initializeAxisClient(device, user, password)
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

func initializeOnvifClient(device contract.Device, user string, password string) *OnvifClient {
	addr := device.Protocols["HTTP"]["Address"]
	c := NewOnvifClient(addr, user, password, driver.lc)
	lock.Lock()
	onvifClients[addr] = c
	lock.Unlock()
	return c
}

func initializeClient(device contract.Device, user string, password string) client.Client {
	addr := device.Protocols["HTTP"]["Address"]

	c := bosch.NewClient(driver.asynchCh, driver.lc)
	c.CameraInit(device, addr, user, password)

	lock.Lock()
	clients[addr] = c
	lock.Unlock()

	return c
}

func initializeAxisClient(device contract.Device, user string, password string) client.Client {
	addr := device.Protocols["HTTP"]["Address"]

	c := axis.NewClient(driver.asynchCh, driver.lc)
	c.CameraInit(device, addr, user, password)

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

func (d *Driver) addrFromProtocols(protocols map[string]contract.ProtocolProperties) (string, error) {
	if _, ok := protocols["HTTP"]; !ok {
		d.lc.Error("No HTTP address found for device. Check configuration file.")
		return "", fmt.Errorf("no HTTP address in protocols map")
	}

	var addr string
	addr, ok := protocols["HTTP"]["Address"]
	if !ok {
		d.lc.Error("No HTTP address found for device. Check configuration file.")
		return "", fmt.Errorf("no HTTP address in protocols map")
	}
	return addr, nil

}

func (d *Driver) clientsFromAddr(addr string, deviceName string) (*OnvifClient, client.Client, error) {
	onvifClient, ok := getOnvifClient(addr)

	if !ok {
		dev, err := device.RunningService().GetDeviceByName(deviceName)
		if err != nil {
			err = fmt.Errorf("device not found: %s", deviceName)
			d.lc.Error(err.Error())

			return nil, nil, err
		}

		onvifClient = initializeOnvifClient(dev, d.config.Camera.User, d.config.Camera.Password)
	}

	c, ok := getClient(addr)

	if !ok {
		dev, err := device.RunningService().GetDeviceByName(deviceName)
		if err != nil {
			err = fmt.Errorf("device not found: %s", deviceName)
			d.lc.Error(err.Error())

			return nil, nil, err
		}
		newClient(dev, d.config.Camera.User, d.config.Camera.Password)
	}

	return onvifClient, c, nil
}
