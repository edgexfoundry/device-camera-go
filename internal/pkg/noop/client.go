package noop

import (
	"fmt"

	"github.com/edgexfoundry/device-sdk-go/pkg/models"
	e_models "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry-holding/device-camera-go/internal/pkg/client"
)

type NoopClient struct {}

func (n NoopClient) HandleReadCommand(req models.CommandRequest) (*models.CommandValue, error) {
	return &models.CommandValue{}, fmt.Errorf("device-camera-go: unrecognized read command")
}

func (n NoopClient) HandleWriteCommand(req models.CommandRequest, param *models.CommandValue) error {
	return fmt.Errorf("device-camera-go: unrecognized write command")
}

func (n NoopClient) CameraRelease(force bool) {
	return
}

func (n NoopClient) CameraInit(edgexDevice e_models.Device, ipAddress string, username string, password string) {
	return
}

func NewClient() client.Client {
	return NoopClient{}
}