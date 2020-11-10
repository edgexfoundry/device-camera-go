package noop

import (
	"fmt"

	"github.com/edgexfoundry/device-sdk-go/pkg/models"
	e_models "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/device-camera-go/internal/pkg/client"
)

// Client is a camera client for cameras which don't have any other camera or manufacturer
// specific clients to leverage.  All of this client's methods return an error if information is
// requested and otherwise comply silently with direction to Initialize or Release.
type Client struct{}

// HandleReadCommand triggers a protocol Read operation for the specified device, resulting in
// an error for an unrecognized read command.
func (n Client) HandleReadCommand(req models.CommandRequest) (*models.CommandValue, error) {
	return &models.CommandValue{}, fmt.Errorf("device-camera-go: unrecognized read command")
}

// HandleWriteCommand triggers a protocol Write operation; resulting in an error for an unrecognized write command
func (n Client) HandleWriteCommand(req models.CommandRequest, param *models.CommandValue) error {
	return fmt.Errorf("device-camera-go: unrecognized write command")
}

// CameraRelease immediately returns control to the caller
func (n Client) CameraRelease(force bool) {
	return
}

// CameraInit immediately returns control to the caller
func (n Client) CameraInit(edgexDevice e_models.Device, ipAddress string, username string, password string) {
	return
}

// NewClient returns a new noop Client
func NewClient() client.Client {
	return Client{}
}
