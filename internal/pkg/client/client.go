package client

import (
	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	e_models "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type Client interface {
	CameraInit(edgexDevice e_models.Device, ipAddress string, username string, password string)
	HandleReadCommand(req sdkModel.CommandRequest) (*sdkModel.CommandValue, error)
	HandleWriteCommand(req sdkModel.CommandRequest, param *sdkModel.CommandValue) error
	CameraRelease(force bool)
}