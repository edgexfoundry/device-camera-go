package client

import (
	sdkModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

// Client is an interface that can be implemented to allow cameras to pass HandleReadCommand
// and HandleWriteCommand requests to device- or manufacturer-specific handlers.  CameraInit
// and CameraRelease allow cameras to spin up long-running goroutines to manage the camera
type Client interface {
	CameraInit(edgexDevice models.Device, edgexProfile models.DeviceProfile, ipAddress string, username string, password string)
	HandleReadCommand(req sdkModels.CommandRequest) (*sdkModels.CommandValue, error)
	HandleWriteCommand(req sdkModels.CommandRequest, param *sdkModels.CommandValue) error
	CameraRelease(force bool)
}
