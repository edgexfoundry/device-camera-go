package driver

import (
	"fmt"

	sdk "github.com/edgexfoundry/device-sdk-go/v2/pkg/service"
)

type configuration struct {
	Camera cameraInfo
}

type cameraInfo struct {
	User       string
	Password   string
	AuthMethod string
}

const (
	USER        = "User"
	PASSWORD    = "Password"
	AUTH_METHOD = "AuthMethod"
)

// loadCameraConfig loads the camera configuration
func loadCameraConfig() (*configuration, error) {
	config := new(configuration)
	if val, ok := sdk.DriverConfigs()[USER]; ok {
		config.Camera.User = val
	} else {
		return config, fmt.Errorf("driver config undefined: %s", USER)
	}
	if val, ok := sdk.DriverConfigs()[PASSWORD]; ok {
		config.Camera.Password = val
	} else {
		return config, fmt.Errorf("driver config undefined: %s", PASSWORD)
	}
	if val, ok := sdk.DriverConfigs()[AUTH_METHOD]; ok {
		config.Camera.AuthMethod = val
	} else {
		return config, fmt.Errorf("driver config undefined: %s", AUTH_METHOD)
	}

	return config, nil
}
