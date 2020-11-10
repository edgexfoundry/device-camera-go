package driver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/faceterteam/onvif4go"
	"github.com/faceterteam/onvif4go/device"
	"github.com/faceterteam/onvif4go/onvif"

	"github.com/edgexfoundry/device-camera-go/internal/pkg/digest"
)

// OnvifClient manages the state required to issue ONVIF requests to a camera
type OnvifClient struct {
	ipAddress    string
	user         string
	password     string
	cameraAuth   string
	onvifDevice  *onvif4go.OnvifDevice
	lc           logger.LoggingClient
	digestClient digest.Client
}

// NewOnvifClient returns an OnvifClient for a single camera
func NewOnvifClient(ipAddress string, user string, password string, cameraAuth string, lc logger.LoggingClient) *OnvifClient {
	c := OnvifClient{
		ipAddress:  ipAddress,
		user:       user,
		password:   password,
		cameraAuth: cameraAuth,
		lc:         lc,
	}

	dev := onvif4go.NewOnvifDevice(c.ipAddress)
	dev.Auth(user, password)
	err := dev.Initialize()
	if err != nil {
		lc.Error(fmt.Sprintf("Error initializing ONVIF Client: %v", err.Error()))
		return nil
	}

	c.onvifDevice = dev

	c.digestClient = digest.NewDClient(&http.Client{}, user, password)
	return &c
}

// GetDeviceInformation makes an ONVIF GetDeviceInformation request to the camera
func (c *OnvifClient) GetDeviceInformation() (string, error) {
	info, err := c.onvifDevice.Device.GetDeviceInformation()
	if err != nil {
		return "", err
	}

	deviceInfo, err := json.Marshal(info)
	if err != nil {
		return "", err
	}

	return string(deviceInfo), nil
}

// GetProfileInformation makes an ONVIF GetProfiles request to the camera
func (c *OnvifClient) GetProfileInformation() (string, error) {
	profiles, err := c.onvifDevice.Media.GetProfiles()
	if err != nil {
		return "", err
	}

	mediaProfiles, err := json.Marshal(profiles)
	if err != nil {
		return "", err
	}

	return string(mediaProfiles), nil
}

// GetStreamURI returns the RTSP URI for the first media profile returned by the camera
func (c *OnvifClient) GetStreamURI() (string, error) {
	profilesResp, err := c.onvifDevice.Media.GetProfiles()
	if err != nil {
		return "", err
	}

	if len(profilesResp.Profiles) == 0 {
		return "", fmt.Errorf("no onvif profiles found")
	}

	token := profilesResp.Profiles[0].Token

	uriResp, err := c.onvifDevice.Media.GetStreamURI(string(token), "RTP-Unicast", "RTSP")
	if err != nil {
		return "", fmt.Errorf("GetStreamURI failed: %v", err.Error())
	}

	uriJSON, err := json.Marshal(uriResp)
	if err != nil {
		return "", fmt.Errorf("error marshaling stream URI to json: %v", err.Error())
	}

	return string(uriJSON), nil
}

// GetSnapshot returns a snapshot from the camera as a slice of bytes
func (c *OnvifClient) GetSnapshot() ([]byte, error) {
	profilesResp, err := c.onvifDevice.Media.GetProfiles()
	if err != nil {
		return nil, err
	}

	if len(profilesResp.Profiles) == 0 {
		return nil, fmt.Errorf("no onvif profiles found")
	}

	token := profilesResp.Profiles[0].Token

	uriResponse, err := c.onvifDevice.Media.GetSnapshotURI(string(token))
	if err != nil {
		return nil, err
	}

	url := uriResponse.MediaUri.Uri

	req, err := http.NewRequest(http.MethodGet, string(url), nil)
	if err != nil {
		return nil, err
	}

	var resp *http.Response

	if c.cameraAuth == "digest" {
		resp, err = c.digestClient.Do(req)
	} else if c.cameraAuth == "basic" {
		req.SetBasicAuth(c.user, c.password)
		httpClient := http.Client{}
		resp, err = httpClient.Do(req)
	} else {
		httpClient := http.Client{}
		resp, err = httpClient.Do(req)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http request for image failed with status %v", resp.StatusCode)
	}

	buf, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("error reading http request: %v", err)
	}

	return buf, nil
}

// GetSystemDateAndTime returns the current date and time as reported by the ONVIF GetSystemDateAndTime command
func (c *OnvifClient) GetSystemDateAndTime() (string, error) {
	datetime, err := c.onvifDevice.Device.GetSystemDateAndTime()
	if err != nil {
		return "", err
	}

	datetimeJSON, err := json.Marshal(datetime)
	if err != nil {
		return "", err
	}

	return string(datetimeJSON), nil
}

// GetHostname returns the hostname reported by the device via the ONVIF GetHostname command
func (c *OnvifClient) GetHostname() (string, error) {
	hostname, err := c.onvifDevice.Device.GetHostname()
	if err != nil {
		return "", err
	}

	hostnameJSON, err := json.Marshal(hostname)
	if err != nil {
		return "", err
	}

	return string(hostnameJSON), nil
}

// SetHostname requests a change to the camera's hostname via the ONFVIF SetHostname command
func (c *OnvifClient) SetHostname(name string) error {
	err := c.onvifDevice.Device.SetHostname(name)
	return err
}

// SetHostnameFromDHCP requests the camera to base its hostname from DHCP
func (c *OnvifClient) SetHostnameFromDHCP() error {
	_, err := c.onvifDevice.Device.SetHostnameFromDHCP(true)

	return err
}

// SetSystemDateAndTime changes the camera's system time via the SetSystemDateAndTime ONVIF command
func (c *OnvifClient) SetSystemDateAndTime(datetime time.Time) error {
	req, err := device.NewSetSystemDateAndTimeManual(datetime, "", false)
	if err != nil {
		c.lc.Error(fmt.Sprintf("Error creating SetSystemDateAndTime object: %v", err.Error()))
		return err
	}

	err = c.onvifDevice.Device.SetSystemDateAndTime(req)
	if err != nil {
		c.lc.Error(fmt.Sprintf("Error calling SetSystemDateAndTime: %v", err.Error()))
	}
	return nil
}

// GetDNS returns the DNS settings as reported by the ONVIF GetDNS command
func (c *OnvifClient) GetDNS() (string, error) {
	dns, err := c.onvifDevice.Device.GetDNS()
	if err != nil {
		return "", err
	}

	dnsJSON, err := json.Marshal(dns)
	if err != nil {
		return "", err
	}

	return string(dnsJSON), nil
}

// GetNetworkInterfaces returns the results of the ONVIF GetNetworkInterfaces command
func (c *OnvifClient) GetNetworkInterfaces() (string, error) {
	interfaces, err := c.onvifDevice.Device.GetNetworkInterfaces()
	if err != nil {
		return "", err
	}

	interfacesJSON, err := json.Marshal(interfaces)
	if err != nil {
		return "", err
	}

	return string(interfacesJSON), nil
}

// GetNetworkProtocols returns the resutls of the ONVIF GetNetworkProtocols command
func (c *OnvifClient) GetNetworkProtocols() (string, error) {
	protocols, err := c.onvifDevice.Device.GetNetworkProtocols()
	if err != nil {
		return "", err
	}

	protocolsJSON, err := json.Marshal(protocols)
	if err != nil {
		return "", err
	}

	return string(protocolsJSON), nil
}

// GetNetworkDefaultGateway returns the results of the ONVIF GetNetworkDefaultGateway command
func (c *OnvifClient) GetNetworkDefaultGateway() (string, error) {
	defaultGateway, err := c.onvifDevice.Device.GetNetworkDefaultGateway()
	if err != nil {
		return "", err
	}

	gatewayJSON, err := json.Marshal(defaultGateway)
	if err != nil {
		return "", err
	}

	return string(gatewayJSON), nil
}

// GetNTP returns the results of the ONVIF GetNTP command
func (c *OnvifClient) GetNTP() (string, error) {
	ntp, err := c.onvifDevice.Device.GetNTP()
	if err != nil {
		return "", err
	}

	ntpJSON, err := json.Marshal(ntp)
	if err != nil {
		return "", err
	}

	return string(ntpJSON), nil
}

// Reboot requests a device system reboot via ONVIF
func (c *OnvifClient) Reboot() (string, error) {
	reboot, err := c.onvifDevice.Device.SystemReboot()
	if err != nil {
		return "", err
	}

	rebootJSON, err := json.Marshal(reboot)
	if err != nil {
		return "", err
	}

	return string(rebootJSON), nil
}

// GetUsers requests the Users associated with the device via ONVIF
func (c *OnvifClient) GetUsers() (string, error) {
	users, err := c.onvifDevice.Device.GetUsers()
	if err != nil {
		return "", err
	}

	usersJSON, err := json.Marshal(users)
	if err != nil {
		return "", err
	}

	return string(usersJSON), nil
}

// CreateUser creates a new ONVIF User for the device
func (c *OnvifClient) CreateUser(user onvif.User) error {
	err := c.onvifDevice.Device.CreateUser(user)
	return err
}
