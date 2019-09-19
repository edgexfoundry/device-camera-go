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

	"github.com/edgexfoundry-holding/device-camera-go/internal/pkg/digest"
)

type OnvifClient struct {
	ipAddress   string
	user        string
	password    string
	onvifDevice *onvif4go.OnvifDevice
	lc          logger.LoggingClient
	digestClient digest.Client
}

func NewOnvifClient(ipAddress string, user string, password string, lc logger.LoggingClient) *OnvifClient {
	c := OnvifClient{
		ipAddress: ipAddress,
		user:      user,
		password:  password,
		lc:        lc,
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

	uriJson, err := json.Marshal(uriResp)
	if err != nil {
		return "", fmt.Errorf("error marshaling stream URI to json: %v", err.Error())
	}

	return string(uriJson), nil
}

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

	resp, err := c.digestClient.Do(req)
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

func (c *OnvifClient) GetSystemDateAndTime() (string, error) {
	datetime, err := c.onvifDevice.Device.GetSystemDateAndTime()
	if err != nil {
		return "", err
	}

	datetimeJson, err := json.Marshal(datetime)
	if err != nil {
		return "", err
	}

	return string(datetimeJson), nil
}

func (c *OnvifClient) GetHostname() (string, error) {
	hostname, err := c.onvifDevice.Device.GetHostname()
	if err != nil {
		return "", err
	}

	hostnameJson, err := json.Marshal(hostname)
	if err != nil {
		return "", err
	}

	return string(hostnameJson), nil
}

func (c *OnvifClient) SetHostname(name string) (error) {
	err := c.onvifDevice.Device.SetHostname(name)
	return err
}

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
func (c *OnvifClient) GetDNS() (string, error) {
	dns, err := c.onvifDevice.Device.GetDNS()
	if err != nil {
		return "", err
	}

	dnsJson, err := json.Marshal(dns)
	if err != nil {
		return "", err
	}

	return string(dnsJson), nil
}

func (c *OnvifClient) GetNetworkInterfaces() (string, error) {
	interfaces, err := c.onvifDevice.Device.GetNetworkInterfaces()
	if err != nil {
		return "", err
	}

	interfacesJson, err := json.Marshal(interfaces)
	if err != nil {
		return "", err
	}

	return string(interfacesJson), nil
}

func (c *OnvifClient) GetNetworkProtocols() (string, error) {
	protocols, err := c.onvifDevice.Device.GetNetworkProtocols()
	if err != nil {
		return "", err
	}

	protocolsJson, err := json.Marshal(protocols)
	if err != nil {
		return "", err
	}

	return string(protocolsJson), nil
}

func (c *OnvifClient) GetNetworkDefaultGateway() (string, error) {
	defaultGateway, err := c.onvifDevice.Device.GetNetworkDefaultGateway()
	if err != nil {
		return "", err
	}

	gatewayJson, err := json.Marshal(defaultGateway)
	if err != nil {
		return "", err
	}

	return string(gatewayJson), nil
}

func (c *OnvifClient) GetNTP() (string, error) {
	ntp, err := c.onvifDevice.Device.GetNTP()
	if err != nil {
		return "", err
	}

	ntpJson, err := json.Marshal(ntp)
	if err != nil {
		return "", err
	}

	return string(ntpJson), nil
}

func (c *OnvifClient) Reboot() (string, error) {
	reboot, err := c.onvifDevice.Device.SystemReboot()
	if err != nil {
		return "", err
	}

	rebootJson, err := json.Marshal(reboot)
	if err != nil {
		return "", err
	}

	return string(rebootJson), nil
}

func (c *OnvifClient) GetUsers() (string, error) {
	users, err := c.onvifDevice.Device.GetUsers()
	if err != nil {
		return "", err
	}

	usersJson, err := json.Marshal(users)
	if err != nil {
		return "", err
	}

	return string(usersJson), nil
}

func (c *OnvifClient) CreateUser(user onvif.User) error {
	err := c.onvifDevice.Device.CreateUser(user)
	return err
}