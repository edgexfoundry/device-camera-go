module github.com/edgexfoundry-holding/device-camera-go

require (
	github.com/edgexfoundry-holding/device-camera-go/cameradiscoveryprovider v0.0.0
	github.com/edgexfoundry/device-sdk-go v0.0.0-20190529004611-4ec3ceb83e9b
	github.com/gorilla/mux v1.7.0 // indirect
)

replace github.com/edgexfoundry-holding/device-camera-go/cameradiscoveryprovider => ./provider
