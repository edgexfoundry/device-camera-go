module github.com/edgexfoundry-holding/device-camera-go

require (
	github.com/edgexfoundry-holding/device-camera-go/cameradiscoveryprovider v0.0.0
	github.com/edgexfoundry/device-sdk-go v0.0.0-20190415234449-132c5ec382ca
	github.com/gorilla/mux v1.7.0 // indirect
)

replace github.com/edgexfoundry-holding/device-camera-go/cameradiscoveryprovider => ./provider
